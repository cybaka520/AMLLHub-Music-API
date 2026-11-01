package qm

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/des"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

// qrcKey 是 QRC 的 3DES 密钥（24 字节 ASCII），逆向固定值，不可改。
var qrcKey = []byte("!@#)(*$%123ZXC!@!@#)(NHL")

// decryptQRC 解密 QRC：hex → 3DES-ECB → zlib。
// 上游密文长度一定是 8 的倍数，不需要处理 padding。
func decryptQRC(hexText string) (string, error) {
	hexText = strings.TrimSpace(hexText)
	cipherText, err := hex.DecodeString(hexText)
	if err != nil {
		return "", err
	}
	if len(cipherText) == 0 {
		return "", errors.New("空密文")
	}
	if len(cipherText)%des.BlockSize != 0 {
		return "", errors.New("密文长度不是 8 的倍数")
	}

	block, err := des.NewTripleDESCipher(qrcKey)
	if err != nil {
		return "", err
	}

	plain := make([]byte, len(cipherText))
	for i := 0; i < len(cipherText); i += block.BlockSize() {
		block.Decrypt(plain[i:i+block.BlockSize()], cipherText[i:i+block.BlockSize()])
	}
	return inflateAny(plain)
}

// inflateAny 依次尝试 zlib / raw deflate / gzip，任一成功且结果像歌词才返回。
func inflateAny(b []byte) (string, error) {
	const max = 4 << 20 // 4MB，防 zip bomb

	if r, err := zlib.NewReader(bytes.NewReader(b)); err == nil {
		defer r.Close()
		if out, err := io.ReadAll(io.LimitReader(r, max)); err == nil && len(out) > 0 {
			return string(out), nil
		}
	}
	// raw deflate 没有校验和，乱数据也可能“解压成功”，必须再校验内容形态。
	if out, err := io.ReadAll(io.LimitReader(flate.NewReader(bytes.NewReader(b)), max)); err == nil {
		if looksLikeLyric(string(out)) {
			return string(out), nil
		}
	}
	if r, err := gzip.NewReader(bytes.NewReader(b)); err == nil {
		defer r.Close()
		if out, err := io.ReadAll(io.LimitReader(r, max)); err == nil && len(out) > 0 {
			return string(out), nil
		}
	}

	// 三种都失败：可能本身未压缩
	if s := string(b); looksLikeLyric(s) {
		return s, nil
	}
	return "", errors.New("QRC 解压失败")
}

// looksLikeLyric 判断内容是否像 LRC / QRC 文本。
func looksLikeLyric(s string) bool {
	if s == "" {
		return false
	}
	return strings.ContainsAny(s, "[<")
}
