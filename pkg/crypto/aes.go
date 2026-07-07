// Package crypto 提供网易云音乐API所需的加密工具
package crypto

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"strings"
)

// AESKey 网易云音乐 eapi 使用的 16 字节 AES 密钥
const AESKey = "e82ckenh8dichen8"

// AESEncryptECB 使用 AES-128-ECB 模式加密数据，返回小写十六进制字符串
func AESEncryptECB(data string) (string, error) {
	block, err := aes.NewCipher([]byte(AESKey))
	if err != nil {
		return "", err
	}

	// PKCS7 填充
	padded := pkcs7Pad([]byte(data), block.BlockSize())

	// ECB 加密
	encrypted := make([]byte, len(padded))
	for i := 0; i < len(padded); i += block.BlockSize() {
		block.Encrypt(encrypted[i:i+block.BlockSize()], padded[i:i+block.BlockSize()])
	}

	// 转小写十六进制
	return strings.ToLower(hex.EncodeToString(encrypted)), nil
}

// pkcs7Pad 实现 PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}
