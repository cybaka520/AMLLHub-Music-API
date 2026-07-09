package models

import "testing"

func TestAPIError_Error(t *testing.T) {
	e := NewAPIError(400, "参数为空")
	if e.Code != 400 || e.Message != "参数为空" {
		t.Fatalf("字段不正确: %+v", e)
	}
	want := "api error 400: 参数为空"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_AsError(t *testing.T) {
	var err error = NewAPIError(500, "内部错误")
	ae, ok := err.(*APIError)
	if !ok {
		t.Fatalf("无法断言为 *APIError")
	}
	if ae.Code != 500 {
		t.Errorf("Code = %d, want 500", ae.Code)
	}
}
