// Package client 提供网易云音乐API的HTTP客户端
package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/utils"
)

// 客户端常量（网易云 eapi/android 通用参数），供 api 包复用避免不一致
const (
	DefaultOS      = "android"
	DefaultAppVer  = "9.3.90"
	DefaultDevice  = "pyncm!"
	maxRespBody    = 10 << 20 // 限制响应体最大 10MB，防止内存耗尽
	requestIDBase  = 20000000
	requestIDRange = 10000001
)

// HTTPClient 网易云音乐API HTTP客户端
type HTTPClient struct {
	client *http.Client
	token  string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(token string) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	return &HTTPClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		token: token,
	}
}

// buildCommonHeaders 构建通用请求头（每请求随机 UA/IP）
func (h *HTTPClient) buildCommonHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", utils.RandomUA())
	headers.Set("X-Real-IP", utils.RandomIP())
	headers.Set("Client-IP", headers.Get("X-Real-IP"))
	headers.Set("X-Forwarded-For", headers.Get("X-Real-IP"))
	headers.Set("Referer", "https://music.163.com")
	return headers
}

// buildCookieString 构建Cookie字符串
func (h *HTTPClient) buildCookieString(extraCookies map[string]string) string {
	defaultCookies := map[string]string{
		"os":       DefaultOS,
		"appver":   DefaultAppVer,
		"osver":    "",
		"deviceId": DefaultDevice,
	}

	// 合并用户传入的cookie
	for k, v := range extraCookies {
		defaultCookies[k] = v
	}

	// 解析token中的cookie
	tokenCookies := utils.ParseCookie(h.token)
	for k, v := range tokenCookies {
		defaultCookies[k] = v
	}

	return utils.BuildCookieString(defaultCookies)
}

// PostEAPI 发送eapi加密POST请求（用于需要加密的接口）
func (h *HTTPClient) PostEAPI(ctx context.Context, reqURL, paramsHex string, cookies map[string]string) ([]byte, error) {
	formData := url.Values{}
	formData.Set("params", paramsHex)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", h.buildCookieString(cookies))

	return h.do(req)
}

// PostForm 发送普通表单POST请求（用于 /api/ 接口）
func (h *HTTPClient) PostForm(ctx context.Context, reqURL string, data url.Values, cookies map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cookieStr := h.buildCookieString(cookies)
	if cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}

	return h.do(req)
}

// Get 发送GET请求
func (h *HTTPClient) Get(ctx context.Context, reqURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	cookieStr := h.buildCookieString(nil)
	if cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}

	return h.do(req)
}

// PostSearch 发送搜索专用POST请求（空body，参数在URL中）
func (h *HTTPClient) PostSearch(ctx context.Context, reqURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "Keep-Alive")

	return h.do(req)
}

// do 执行请求并校验响应状态码，限制读取大小
func (h *HTTPClient) do(req *http.Request) ([]byte, error) {
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRespBody))
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

// RandomRequestID 生成随机requestId（20000000-30000000）
func RandomRequestID() string {
	return fmt.Sprintf("%d", requestIDBase+rand.Intn(requestIDRange))
}
