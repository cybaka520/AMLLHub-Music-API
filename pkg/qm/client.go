package qm

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client 是 QM 解析客户端。
//
// 同一账号在同一进程内只应使用一个 Client（或让多个 Client 共享同一个 CookieStore）：
// refresh_key 一次性轮换，并发刷新会互相踢掉。
type Client struct {
	mu     sync.Mutex
	cli    *http.Client
	store  CookieStore
	cookie *cookieSet

	// gen 在凭据刷新时自增，用于丢弃刷新前发出的 session 结果。
	gen uint64

	session         sessionInfo
	sessionInflight chan struct{}
	sessionErr      error

	refreshInflight chan struct{}
	refreshErr      error
}

// Option 客户端选项。
type Option func(*Client)

// NewClient 创建客户端。
//
// 不传 CookieStore 时使用内存实现，NewClient 的入参作为初值，续期只在本 Client 内生效。
func NewClient(cookie string, opts ...Option) *Client {
	c := &Client{
		cli:    newHTTPClient(),
		store:  newMemoryStore(cookie),
		cookie: parseCookieSet(cookie),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	// 传了 CookieStore 时以 Store 为准，入参 cookie 只在 Store 读取失败时兜底。
	if c.store != nil {
		if s, err := c.store.Load(); err == nil && strings.TrimSpace(s) != "" {
			c.cookie = parseCookieSet(s)
		}
	}
	return c
}

// WithCookieStore 指定凭据存储，用于跨进程复用续期后的 cookie。
func WithCookieStore(s CookieStore) Option {
	return func(c *Client) {
		if s != nil {
			c.store = s
		}
	}
}

// WithTimeout 设置单次请求的整体超时（默认 30s）。
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.cli.Timeout = d
		}
	}
}

// WithHTTPClient 替换底层 HTTP 客户端（代理、连接池等）。
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.cli = h
		}
	}
}

// Cookie 返回当前生效的 cookie 串（可能已被续期更新）。
func (c *Client) Cookie() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cookie == nil {
		return ""
	}
	return c.cookie.String()
}

// RefreshCredential 手动触发一次凭据续期。
// 正常情况下无需调用，请求命中 1000/2001 时会自动续期一次。
func (c *Client) RefreshCredential(ctx context.Context) error {
	return c.refresh(ctx)
}

// newHTTPClient 构造默认 HTTP 客户端，复用连接并限制整体超时。
func newHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   defaultTimeout,
	}
}
