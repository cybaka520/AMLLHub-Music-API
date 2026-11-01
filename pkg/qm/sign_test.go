package qm

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"testing"
)

// TestSignZZCGolden 用黄金向量锁住签名实现与 comm 字段顺序。
// payload 一旦变化，说明 struct 顺序或 JSON 编码行为被改动。
func TestSignZZCGolden(t *testing.T) {
	type testParam struct {
		Text string `json:"text"`
	}

	env := requestEnvelope{
		Comm:    baseComm(),
		Request: &reqBody{Module: "test.module", Method: "Test", Param: testParam{Text: "中文 /+="}},
	}
	payload, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	wantPayload := `{"comm":{"ct":11,"cv":14090008,"v":14090008,"chid":"10003505","os_ver":"15",` +
		`"phonetype":"24122RKC7C","tmeAppID":"qqmusic","nettype":"NETWORK_WIFI","udid":"0",` +
		`"OpenUDID":"0","QIMEI36":"0","uin":"0"},"request":{"module":"test.module",` +
		`"method":"Test","param":{"text":"中文 /+="}}}`
	if string(payload) != wantPayload {
		t.Fatalf("payload 与黄金向量不一致:\n got: %s\nwant: %s", payload, wantPayload)
	}

	sum := sha1.Sum(payload)
	if got := hex.EncodeToString(sum[:]); got != "07d16855b5c5a4e1aca2eb71f9978403eb58f665" {
		t.Fatalf("sha1 不匹配: %s", got)
	}

	const wantSign = "zzc1e5fa52xvbbw2xnh0dllekgdd2hmt5j9ya7ea27b8"
	if got := signZZC(payload); got != wantSign {
		t.Fatalf("sign 不匹配:\n got: %s\nwant: %s", got, wantSign)
	}
}

// TestHash33 校验 g_tk 计算：未登录时 seed 原样返回，有内容时按 JS 的移位规则逐字节累加。
func TestHash33(t *testing.T) {
	if got := hash33("", 5381); got != 5381 {
		t.Fatalf("空串 hash33 应为 seed，实际 %d", got)
	}
	if got := hash33("abc", 5381); got != 193485963 {
		t.Fatalf("hash33 计算异常: %d", got)
	}
}
