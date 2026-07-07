package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 计算字符串的 MD5 哈希，返回十六进制字符串
func MD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
