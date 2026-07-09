package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestMD5(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"test", "098f6bcd4621d373cade4e832627b4f6"},
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
	}
	for _, c := range cases {
		if got := MD5(c.in); got != c.want {
			t.Errorf("MD5(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMD5_LengthAndHex(t *testing.T) {
	got := MD5("任意中文")
	if len(got) != 32 {
		t.Fatalf("MD5 长度 = %d, want 32", len(got))
	}
	if _, err := hex.DecodeString(got); err != nil {
		t.Errorf("MD5 结果不是合法十六进制: %v", err)
	}
}

func TestAESEncryptECB_DeterministicAndHex(t *testing.T) {
	a, err := AESEncryptECB("hello")
	if err != nil {
		t.Fatalf("AESEncryptECB 返回错误: %v", err)
	}
	b, _ := AESEncryptECB("hello")
	if a != b {
		t.Errorf("AES-ECB 不确定: %q != %q", a, b)
	}
	if _, err := hex.DecodeString(a); err != nil {
		t.Errorf("AES 结果不是合法十六进制: %v", err)
	}
	if a != strings.ToLower(a) {
		t.Errorf("AES 结果应是小写十六进制: %q", a)
	}
}

func TestAESEncryptECB_BlockAlignment(t *testing.T) {
	// PKCS7 填充后长度应为 16 的倍数；密文长度同样
	out, err := AESEncryptECB("a")
	if err != nil {
		t.Fatalf("AESEncryptECB 错误: %v", err)
	}
	raw, err := hex.DecodeString(out)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	if len(raw)%16 != 0 {
		t.Errorf("密文长度 %d 不是 16 的倍数", len(raw))
	}
	// 空串也应对齐到 16 字节（一整块填充）
	out2, _ := AESEncryptECB("")
	raw2, _ := hex.DecodeString(out2)
	if len(raw2) != 16 {
		t.Errorf("空串密文长度 = %d, want 16", len(raw2))
	}
}

func TestEncryptAPIRequest(t *testing.T) {
	got, err := EncryptAPIRequest("/api/song/enhance/player/url/v1", map[string]interface{}{
		"ids":        []string{"123"},
		"level":      "lossless",
		"encodeType": "flac",
	})
	if err != nil {
		t.Fatalf("EncryptAPIRequest 错误: %v", err)
	}
	if _, err := hex.DecodeString(got); err != nil {
		t.Errorf("结果不是合法十六进制: %v", err)
	}
	// 同输入应确定
	got2, _ := EncryptAPIRequest("/api/song/enhance/player/url/v1", map[string]interface{}{
		"ids":        []string{"123"},
		"level":      "lossless",
		"encodeType": "flac",
	})
	if got != got2 {
		t.Errorf("EncryptAPIRequest 不确定")
	}
	// 不同 apiPath 应产生不同密文
	other, _ := EncryptAPIRequest("/api/v3/song/detail", map[string]interface{}{
		"ids":        []string{"123"},
		"level":      "lossless",
		"encodeType": "flac",
	})
	if got == other {
		t.Errorf("不同 apiPath 产生了相同密文，可能未参与加密")
	}
}
