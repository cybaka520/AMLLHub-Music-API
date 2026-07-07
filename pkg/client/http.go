// Package client 提供网易云音乐API的HTTP客户端
package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/utils"
)

// HTTPClient 网易云音乐API HTTP客户端
type HTTPClient struct {
	client    *http.Client
	token     string
	userAgent string
	fakeIP    string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(token string) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	return &HTTPClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		token:     token,
		userAgent: utils.RandomUA(),
		fakeIP:    utils.RandomIP(),
	}
}

// buildCommonHeaders 构建通用请求头（含伪造IP/UA）
func (h *HTTPClient) buildCommonHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", h.userAgent)
	headers.Set("X-Real-IP", h.fakeIP)
	headers.Set("Client-IP", h.fakeIP)
	headers.Set("X-Forwarded-For", h.fakeIP)
	headers.Set("Referer", "https://music.163.com")
	return headers
}

// buildCookieString 构建Cookie字符串
func (h *HTTPClient) buildCookieString(extraCookies map[string]string) string {
	defaultCookies := map[string]string{
		"os":       "android",
		"appver":   "9.3.90",
		"osver":    "",
		"deviceId": "pyncm!",
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
func (h *HTTPClient) PostEAPI(reqURL, paramsHex string, cookies map[string]string) ([]byte, error) {
	formData := url.Values{}
	formData.Set("params", paramsHex)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", h.buildCookieString(cookies))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// PostForm 发送普通表单POST请求（用于 /api/ 接口）
func (h *HTTPClient) PostForm(reqURL string, data url.Values, cookies map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cookieStr := h.buildCookieString(cookies)
	if cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Get 发送GET请求
func (h *HTTPClient) Get(reqURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	cookieStr := h.buildCookieString(nil)
	if cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// PostSearch 发送搜索专用POST请求（空body，参数在URL中）
func (h *HTTPClient) PostSearch(reqURL string) ([]byte, error) {
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = h.buildCommonHeaders()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "Keep-Alive")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// RandomRequestID 生成随机requestId（20000000-30000000）
func RandomRequestID() string {
	return fmt.Sprintf("%d", 20000000+rand.Intn(10000001))
}
