package qm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 信封键：普通接口用 request，直链接口用 req_0。
const (
	reqKeyDefault = "request"
	reqKeyVkey    = "req_0"
)

// 凭据失效相关的内层 code。
const (
	codeNeedLogin = 1000
	codeAuthFail  = 2001
)

// callOptions 控制单次上游调用。
type callOptions struct {
	session     bool      // 是否先取 session，默认 true
	auth        bool      // 是否带上登录态，默认 true
	signed      bool      // 是否走 musics.fcg + zzc 签名
	comm        commPatch // comm 字段覆盖
	reqKey      string    // 信封键，默认 request
	rawComm     any       // 完全替换 comm（直链接口用）
	noRetryAuth bool      // 刷新流程自身禁止再刷新
}

// defaultCall 返回默认选项：取 session、带登录态。
func defaultCall() callOptions {
	return callOptions{session: true, auth: true}
}

// reqBody 是信封里的单个请求体。
type reqBody struct {
	Module string `json:"module"`
	Method string `json:"method"`
	Param  any    `json:"param"`
}

// requestEnvelope 是发送给上游的信封。
type requestEnvelope struct {
	Comm    any      `json:"comm"`
	Request *reqBody `json:"request,omitempty"`
	Req0    *reqBody `json:"req_0,omitempty"`
}

// responseEnvelope 是上游返回的信封。
type responseEnvelope struct {
	Code    int         `json:"code"`
	Request *innerBlock `json:"request"`
	Req0    *innerBlock `json:"req_0"`
}

// innerBlock 是信封内层，code 需要独立校验。
type innerBlock struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

// call 是统一的业务调用入口，返回信封内的 data 字段。
func (c *Client) call(ctx context.Context, module, method string, param any, opt callOptions) (json.RawMessage, error) {
	if opt.session {
		// session 失败不影响主流程：多数接口没有 session 也能用。
		_ = c.ensureSession(ctx)
	}

	refreshed := false
	for attempt := 0; ; attempt++ {
		data, err := c.doCall(ctx, module, method, param, opt)
		if err == nil {
			return data, nil
		}

		var reqErr *requestError
		if errors.As(err, &reqErr) {
			// 仅“已登录 + 命中登录失效码”才续期，且一个请求生命周期内最多一次。
			if isAuthCode(reqErr.Inner) && !refreshed && !opt.noRetryAuth && c.loggedIn() {
				refreshed = true
				if rerr := c.refresh(ctx); rerr != nil {
					return nil, fmt.Errorf("%w: %v", ErrCredentialExpired, rerr)
				}
				// 续期作废了旧 session，重新取一次再重放。
				if opt.session {
					_ = c.ensureSession(ctx)
				}
				continue
			}
			return nil, err
		}

		if !isRetryable(err) || attempt >= maxRetries {
			return nil, err
		}
		if serr := sleepCtx(ctx, retryBackoff); serr != nil {
			return nil, serr
		}
	}
}

// doCall 执行一次完整请求（构造 → 签名 → 发送 → 双层校验）。
func (c *Client) doCall(ctx context.Context, module, method string, param any, opt callOptions) (json.RawMessage, error) {
	payload, err := c.buildPayload(module, method, param, opt)
	if err != nil {
		return nil, err
	}

	reqURL := hostMusicu
	if opt.signed {
		reqURL = searchGateway + "?sign=" + signZZC(payload)
	}

	reqCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", uaFCG)
	req.Header.Set("Cookie", c.cookieHeader())

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, &networkError{err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRespBody))
	if err != nil {
		return nil, &networkError{err: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &statusError{Status: resp.StatusCode, Body: truncateBody(body)}
	}

	var env responseEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("qm: 解析响应失败: %w", err)
	}

	inner := env.Request
	if opt.reqKey == reqKeyVkey {
		inner = env.Req0
	}

	// 双层 code 必须同时为 0。
	if env.Code != 0 || inner == nil {
		return nil, &requestError{Outer: env.Code, Module: module, Method: method}
	}
	if inner.Code != 0 {
		return nil, &requestError{Inner: inner.Code, Module: module, Method: method}
	}
	return inner.Data, nil
}

// buildPayload 序列化信封。
// 同一个请求只序列化一次：签名与发送共用同一份字节。
func (c *Client) buildPayload(module, method string, param any, opt callOptions) ([]byte, error) {
	var comm any = c.buildComm(opt)
	if opt.rawComm != nil {
		comm = opt.rawComm
	}

	body := &reqBody{Module: module, Method: method, Param: param}
	env := requestEnvelope{Comm: comm}
	if opt.reqKey == reqKeyVkey {
		env.Req0 = body
	} else {
		env.Request = body
	}
	return json.Marshal(env)
}

// buildComm 构造 comm：基础值 → 字段覆盖 → 登录态 → session。
func (c *Client) buildComm(opt callOptions) commCommon {
	comm := baseComm()
	comm.apply(opt.comm)

	if opt.auth {
		c.mu.Lock()
		jar := c.cookie
		if jar != nil {
			if key := jar.musicKey(); key != "" {
				if uin := jar.uin(); uin != "" && uin != commZero {
					comm.Uin = uin
					comm.QQ = uin
				}
				comm.Authst = key
				if lt := jar.loginType(); lt != 0 {
					comm.TMELoginType = lt
				}
			}
		}
		c.mu.Unlock()
	}

	if opt.session {
		c.mu.Lock()
		s := c.session
		c.mu.Unlock()
		if s.UID != "" {
			comm.UID = s.UID
			comm.SID = s.SID
			comm.UserIP = s.UserIP
		}
	}
	return comm
}

// cookieHeader 返回请求头 Cookie。
// 有凭据用完整串；无凭据必须带 tmeLoginType=-1，否则匿名搜索会被判 2001。
func (c *Client) cookieHeader() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cookie == nil || c.cookie.musicKey() == "" {
		return anonymousCookie
	}
	return c.cookie.String()
}

// sessionParam 是取 session 的参数。
type sessionParam struct {
	Caller int   `json:"caller"`
	UID    int64 `json:"uid"`
	Vkey   int   `json:"vkey"`
}

// sessionInfo 是上游下发的会话信息，TTL 1 小时。
type sessionInfo struct {
	UID      string
	SID      string
	UserIP   string
	ExpireAt time.Time
}

// ensureSession 单飞获取 session：首次并发只发一次请求，其余等待。
func (c *Client) ensureSession(ctx context.Context) error {
	c.mu.Lock()
	if c.session.UID != "" && time.Now().Before(c.session.ExpireAt) {
		c.mu.Unlock()
		return nil
	}
	if ch := c.sessionInflight; ch != nil {
		c.mu.Unlock()
		select {
		case <-ch:
			c.mu.Lock()
			err := c.sessionErr
			c.mu.Unlock()
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	ch := make(chan struct{})
	c.sessionInflight = ch
	gen := c.gen
	c.mu.Unlock()

	err := c.fetchSession(ctx, gen)

	c.mu.Lock()
	c.sessionErr = err
	c.sessionInflight = nil
	c.mu.Unlock()
	close(ch)
	return err
}

// fetchSession 真正发起取 session 的请求。
func (c *Client) fetchSession(ctx context.Context, gen uint64) error {
	uin, _ := strconv.ParseInt(c.currentUin(), 10, 64)
	opt := callOptions{auth: true, noRetryAuth: true}
	data, err := c.call(ctx, "music.getSession.session", "GetSession", sessionParam{UID: uin}, opt)
	if err != nil {
		return err
	}

	var resp struct {
		Session struct {
			UID    string `json:"uid"`
			SID    string `json:"sid"`
			UserIP string `json:"userip"`
		} `json:"session"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("qm: 解析 session 失败: %w", err)
	}
	if resp.Session.UID == "" {
		return errors.New("qm: session 响应为空")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// 刷新凭据期间发出的请求，结果已过期，直接丢弃。
	if c.gen != gen {
		return nil
	}
	c.session = sessionInfo{
		UID:      resp.Session.UID,
		SID:      resp.Session.SID,
		UserIP:   resp.Session.UserIP,
		ExpireAt: time.Now().Add(sessionTTL),
	}
	return nil
}

// isAuthCode 判断内层 code 是否代表登录态失效。
func isAuthCode(code int) bool {
	return code == codeNeedLogin || code == codeAuthFail
}

// isRetryable 判断错误是否值得重试。
func isRetryable(err error) bool {
	var netErr *networkError
	if errors.As(err, &netErr) {
		return true
	}
	var stErr *statusError
	if errors.As(err, &stErr) {
		return stErr.Status >= 500 || stErr.Status == http.StatusTooManyRequests
	}
	return false
}

// sleepCtx 在支持 context 取消的前提下等待。
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// truncateBody 截断错误响应体，避免日志被大包刷屏。
func truncateBody(b []byte) string {
	const limit = 512
	s := strings.TrimSpace(string(b))
	if len(s) > limit {
		return s[:limit]
	}
	return s
}
