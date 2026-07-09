package models

import "fmt"

// APIError 表示 API 返回的业务/系统错误。
// 当 API 调用无法返回有效数据时（参数非法、上游返回非 200、网络失败等），
// 统一通过此类型作为 error 返回，调用方可断言获取错误码。
type APIError struct {
	Code    int
	Message string
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.Code, e.Message)
}

// NewAPIError 构造一个 APIError
func NewAPIError(code int, message string) *APIError {
	return &APIError{Code: code, Message: message}
}
