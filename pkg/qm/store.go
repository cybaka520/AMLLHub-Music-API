package qm

import (
	"strconv"
	"strings"
	"sync"
)

// CookieStore 凭据的持久化接口。
//
// Save 必须整体替换 cookie 串，不允许只写 musickey：
// refresh_key 会随续期轮换，丢了就永久失去续期能力。
type CookieStore interface {
	Load() (string, error)
	Save(cookie string) error
}

// memoryStore 进程内存储，默认实现。
type memoryStore struct {
	mu     sync.Mutex
	cookie string
}

// newMemoryStore 创建内存存储。
func newMemoryStore(cookie string) *memoryStore {
	return &memoryStore{cookie: cookie}
}

// Load 返回当前 cookie。
func (m *memoryStore) Load() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cookie, nil
}

// Save 覆盖写入 cookie。
func (m *memoryStore) Save(cookie string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cookie = cookie
	return nil
}

// cookieSet 保序的 cookie 集合。
//
// 值保持原文不做 URL 转义：musickey 含 * 与 @，转义后上游会判定无效。
type cookieSet struct {
	keys []string
	vals map[string]string
}

// parseCookieSet 解析 "k=v; k2=v2" 形式的 cookie 串，保持出现顺序。
func parseCookieSet(s string) *cookieSet {
	cs := &cookieSet{vals: make(map[string]string)}
	for _, item := range strings.Split(s, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		idx := strings.Index(item, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(item[:idx])
		val := strings.TrimSpace(item[idx+1:])
		if key == "" {
			continue
		}
		cs.Set(key, val)
	}
	return cs
}

// Get 读取 cookie 值。
func (c *cookieSet) Get(key string) string {
	if c == nil {
		return ""
	}
	return c.vals[key]
}

// Set 写入 cookie 值，新键追加到末尾。
func (c *cookieSet) Set(key, val string) {
	if _, ok := c.vals[key]; !ok {
		c.keys = append(c.keys, key)
	}
	c.vals[key] = val
}

// Clone 深拷贝，避免刷新过程中并发读到半成品状态。
func (c *cookieSet) Clone() *cookieSet {
	if c == nil {
		return &cookieSet{vals: make(map[string]string)}
	}
	out := &cookieSet{
		keys: make([]string, len(c.keys)),
		vals: make(map[string]string, len(c.vals)),
	}
	copy(out.keys, c.keys)
	for k, v := range c.vals {
		out.vals[k] = v
	}
	return out
}

// String 拼接为请求头用的 cookie 串。
func (c *cookieSet) String() string {
	if c == nil || len(c.keys) == 0 {
		return ""
	}
	parts := make([]string, 0, len(c.keys))
	for _, k := range c.keys {
		parts = append(parts, k+"="+c.vals[k])
	}
	return strings.Join(parts, "; ")
}

// musicKey 返回鉴权主体 musickey，两个键名二选一。
func (c *cookieSet) musicKey() string {
	return firstNonEmpty(c.Get("qm_keyst"), c.Get("qqmusic_key"))
}

// uin 返回账号 ID，多键兜底并去掉可能存在的 o 前缀。
func (c *cookieSet) uin() string {
	v := firstNonEmpty(
		c.Get("qm_str_musicid"),
		c.Get("uin"),
		c.Get("wxuin"),
		c.Get("p_uin"),
	)
	return strings.TrimPrefix(strings.TrimSpace(v), "o")
}

// loginType 返回登录方式：1 微信，2 QQ。
// cookie 缺失时按 musickey 的 W_X 前缀推断。
func (c *cookieSet) loginType() int {
	if v := c.Get("tmeLoginType"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n != 0 {
			return n
		}
	}
	if strings.HasPrefix(c.musicKey(), "W_X") {
		return loginTypeWechat
	}
	return loginTypeQQ
}

// 登录方式常量。
const (
	loginTypeWechat = 1
	loginTypeQQ     = 2
)
