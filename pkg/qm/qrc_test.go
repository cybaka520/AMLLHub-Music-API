package qm

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"testing"
)

// encryptQRC 用与解密相反的过程自造样本：zlib → 8 字节对齐 → 3DES-ECB → hex。
func encryptQRC(t *testing.T, plain string) string {
	t.Helper()

	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write([]byte(plain)); err != nil {
		t.Fatalf("压缩失败: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("关闭压缩流失败: %v", err)
	}

	data := buf.Bytes()
	if r := len(data) % 8; r != 0 {
		data = append(data, make([]byte, 8-r)...)
	}

	// 必须用 QQ 专有的 3DES 实现加密，标准库 crypto/des 与它不是同一套算法
	return hex.EncodeToString(qrcTripleDESEncrypt(data, qrcKey))
}

// TestDecryptQRCRoundTrip 验证 zlib 分支的加解密闭环。
func TestDecryptQRCRoundTrip(t *testing.T) {
	const plain = `[ti:测试]` + "\n" + `[0,1000](0,500,0)测(500,500,0)试`

	got, err := decryptQRC(encryptQRC(t, plain))
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if got != plain {
		t.Fatalf("明文中含尾部填充时不一致:\n got: %q\nwant: %q", got, plain)
	}
}

// TestDecryptQRCInvalid 覆盖三类异常输入。
func TestDecryptQRCInvalid(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{name: "非法 hex", in: "zzzz"},
		{name: "奇数长度", in: "abc"},
		{name: "空密文", in: ""},
		{name: "长度非 8 的倍数", in: "00112233445566"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decryptQRC(tc.in); err == nil {
				t.Fatalf("期望返回错误，实际为 nil")
			}
		})
	}
}
