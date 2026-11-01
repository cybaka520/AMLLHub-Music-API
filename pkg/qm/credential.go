package qm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// loginParamQQ 是 QQ 型账号的续期参数。
// 字段即使用时才存在，缺失也要传空值 —— 所以不用 omitempty，改用两个独立结构体。
type loginParamQQ struct {
	OpenID       string `json:"openid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiredIn    int64  `json:"expired_in"`
	MusicID      int64  `json:"musicid"`
	MusicKey     string `json:"musickey"`
	RefreshKey   string `json:"refresh_key"`
	LoginMode    int    `json:"loginMode"`
}

// loginParamWechat 是微信型账号的续期参数。
type loginParamWechat struct {
	OpenID       string `json:"openid"`
	RefreshToken string `json:"refresh_token"`
	StrMusicID   string `json:"str_musicid"`
	UnionID      string `json:"unionid"`
	MusicKey     string `json:"musickey"`
	RefreshKey   string `json:"refresh_key"`
	LoginMode    int    `json:"loginMode"`
}

// loginResult 是续期响应，字段名与上游一致。
type loginResult struct {
	MusicKey           string  `json:"musickey"`
	OpenID             string  `json:"openid"`
	UnionID            string  `json:"unionid"`
	RefreshToken       string  `json:"refresh_token"`
	AccessToken        string  `json:"access_token"`
	ExpiredAt          flexInt `json:"expired_at"`
	MusicID            flexInt `json:"musicid"`
	StrMusicID         string  `json:"str_musicid"`
	RefreshKey         string  `json:"refresh_key"`
	MusicKeyCreateTime flexInt `json:"musickeyCreateTime"`
	EncryptUin         string  `json:"encryptUin"`
	LoginType          int     `json:"loginType"`
}

// loggedIn 判断当前是否持有可用凭据。
func (c *Client) loggedIn() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cookie != nil && c.cookie.musicKey() != ""
}

// currentUin 返回当前账号 ID。
func (c *Client) currentUin() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cookie == nil {
		return ""
	}
	return c.cookie.uin()
}

// currentMusicKey 返回当前 musickey。
func (c *Client) currentMusicKey() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cookie == nil {
		return ""
	}
	return c.cookie.musicKey()
}

// refresh 单飞续期凭据：并发触发时后到者等待同一结果。
func (c *Client) refresh(ctx context.Context) error {
	c.mu.Lock()
	if ch := c.refreshInflight; ch != nil {
		c.mu.Unlock()
		select {
		case <-ch:
			c.mu.Lock()
			err := c.refreshErr
			c.mu.Unlock()
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	ch := make(chan struct{})
	c.refreshInflight = ch
	c.mu.Unlock()

	err := c.doRefresh(ctx)

	c.mu.Lock()
	c.refreshErr = err
	c.refreshInflight = nil
	if err == nil {
		// 旧 session 必须作废，gen 自增让在途的 session 结果被丢弃。
		c.gen++
		c.session = sessionInfo{}
	}
	c.mu.Unlock()
	close(ch)
	return err
}

// doRefresh 发起 LoginServer.Login 并回写 cookie。
func (c *Client) doRefresh(ctx context.Context) error {
	c.mu.Lock()
	jar := c.cookie.Clone()
	c.mu.Unlock()

	musicKey := jar.musicKey()
	if musicKey == "" {
		return ErrCredentialExpired
	}

	loginType := jar.loginType()
	uin := jar.uin()
	refreshKey := jar.Get("qm_refresh_key")

	var param any
	if loginType == loginTypeWechat {
		param = loginParamWechat{
			OpenID:       firstNonEmpty(jar.Get("wxopenid"), jar.Get("psrf_qqopenid")),
			RefreshToken: firstNonEmpty(jar.Get("wxrefresh_token"), jar.Get("psrf_qqrefresh_token")),
			StrMusicID:   firstNonEmpty(jar.Get("qm_str_musicid"), uin),
			UnionID:      jar.Get("psrf_qqunionid"),
			MusicKey:     musicKey,
			RefreshKey:   refreshKey,
			LoginMode:    2,
		}
	} else {
		param = loginParamQQ{
			OpenID:       firstNonEmpty(jar.Get("psrf_qqopenid"), jar.Get("wxopenid")),
			AccessToken:  jar.Get("psrf_qqaccess_token"),
			RefreshToken: firstNonEmpty(jar.Get("psrf_qqrefresh_token"), jar.Get("wxrefresh_token")),
			ExpiredIn:    parseFlexInt(jar.Get("psrf_access_token_expiresAt")),
			MusicID:      parseFlexInt(uin),
			MusicKey:     musicKey,
			RefreshKey:   refreshKey,
			LoginMode:    2,
		}
	}

	opt := callOptions{auth: true, noRetryAuth: true}
	data, err := c.call(ctx, "music.login.LoginServer", "Login", param, opt)
	if err != nil {
		return err
	}

	var res loginResult
	if err := json.Unmarshal(data, &res); err != nil {
		return fmt.Errorf("qm: 解析续期响应失败: %w", err)
	}
	if res.MusicKey == "" {
		return ErrCredentialExpired
	}

	next := jar.Clone()
	next.Set("qm_keyst", res.MusicKey)
	next.Set("qqmusic_key", res.MusicKey)

	if res.LoginType != 0 {
		loginType = res.LoginType
	}
	next.Set("tmeLoginType", strconv.Itoa(loginType))

	if id := strings.TrimPrefix(firstNonEmpty(res.StrMusicID, strconv.FormatInt(res.MusicID.int64(), 10)), "o"); id != "" {
		next.Set("qm_str_musicid", id)
		next.Set("uin", id)
		if loginType == loginTypeWechat {
			next.Set("wxuin", id)
		}
	}
	setIfNotEmpty(next, "euin", res.EncryptUin)
	if loginType == loginTypeWechat {
		setIfNotEmpty(next, "wxopenid", res.OpenID)
	} else {
		setIfNotEmpty(next, "psrf_qqopenid", res.OpenID)
	}
	setIfNotEmpty(next, "psrf_qqunionid", res.UnionID)
	if loginType == loginTypeWechat {
		setIfNotEmpty(next, "wxrefresh_token", res.RefreshToken)
	} else {
		setIfNotEmpty(next, "psrf_qqrefresh_token", res.RefreshToken)
	}
	setIfNotEmpty(next, "psrf_qqaccess_token", res.AccessToken)
	if res.ExpiredAt != 0 {
		setIfNotEmpty(next, "psrf_access_token_expiresAt", strconv.FormatInt(res.ExpiredAt.int64(), 10))
	}
	if res.MusicKeyCreateTime != 0 {
		setIfNotEmpty(next, "psrf_musickey_createtime", strconv.FormatInt(res.MusicKeyCreateTime.int64(), 10))
	}
	// refresh_key 会轮换，丢失就永久失去续期能力，必须落盘。
	setIfNotEmpty(next, "qm_refresh_key", res.RefreshKey)

	// 先写 Store 成功，再更新内存态，避免两者不一致。
	if c.store != nil {
		if err := c.store.Save(next.String()); err != nil {
			return fmt.Errorf("qm: 保存凭据失败: %w", err)
		}
	}

	c.mu.Lock()
	c.cookie = next
	c.mu.Unlock()
	return nil
}

// setIfNotEmpty 仅在值非空时写入，避免用空值覆盖已有凭据。
func setIfNotEmpty(cs *cookieSet, key, val string) {
	if val != "" {
		cs.Set(key, val)
	}
}

// parseFlexInt 解析可能为字符串的数字。
func parseFlexInt(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
