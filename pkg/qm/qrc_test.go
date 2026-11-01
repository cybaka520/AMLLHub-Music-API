package qm

import (
	"bytes"
	"compress/zlib"
	"crypto/des"
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
	if r := len(data) % des.BlockSize; r != 0 {
		data = append(data, make([]byte, des.BlockSize-r)...)
	}

	block, err := des.NewTripleDESCipher(qrcKey)
	if err != nil {
		t.Fatalf("构造 3DES 失败: %v", err)
	}

	out := make([]byte, len(data))
	for i := 0; i < len(data); i += block.BlockSize() {
		block.Encrypt(out[i:i+block.BlockSize()], data[i:i+block.BlockSize()])
	}
	return hex.EncodeToString(out)
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
