package qm

import (
	"errors"
	"fmt"
)

var (
	// ErrNoPlayableURL 表示拿不到可播放直链（会员专享、无版权或地区限制）。
	ErrNoPlayableURL = errors.New("qm: 无法获取播放链接（可能为会员专享、无版权或地区限制）")
	// ErrCredentialExpired 表示登录凭据失效且续期失败。
	ErrCredentialExpired = errors.New("qm: 登录凭据已失效，请重新获取 Cookie")
	// ErrSongNotFound 表示按 id / mid 都查不到歌曲。
	ErrSongNotFound = errors.New("qm: 未找到该歌曲")
)

// requestError 上游双层 code 校验失败。
type requestError struct {
	Outer  int
	Inner  int
	Module string
	Method string
}

// Error 实现 error 接口。
func (e *requestError) Error() string {
	return fmt.Sprintf("qm: 上游返回异常 outer=%d inner=%d (%s.%s)", e.Outer, e.Inner, e.Module, e.Method)
}

// statusError HTTP 非 2xx 响应，5xx 可重试。
type statusError struct {
	Status int
	Body   string
}

// Error 实现 error 接口。
func (e *statusError) Error() string {
	return fmt.Sprintf("qm: HTTP %d: %s", e.Status, e.Body)
}

// networkError 传输层错误（可重试）。
type networkError struct {
	err error
}

// Error 实现 error 接口。
func (e *networkError) Error() string { return "qm: 网络请求失败: " + e.err.Error() }

// Unwrap 支持 errors.Is / errors.As 穿透。
func (e *networkError) Unwrap() error { return e.err }
