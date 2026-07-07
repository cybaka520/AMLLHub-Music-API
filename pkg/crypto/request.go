package crypto

import (
	"encoding/json"
	"fmt"
)

// EncryptAPIRequest 加密网易云音乐 eapi 请求参数
// apiPath 格式如 "/api/song/enhance/player/url/v1"
// payload 为请求参数 map
// 返回加密后的小写十六进制字符串
func EncryptAPIRequest(apiPath string, payload interface{}) (string, error) {
	// 序列化 payload 为 JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// 构建签名文本并计算 MD5
	digestText := fmt.Sprintf("nobody%suse%smd5forencrypt", apiPath, string(payloadJSON))
	digestHex := MD5(digestText)

	// 构建待加密文本
	encryptText := fmt.Sprintf("%s-36cd479b6b5-%s-36cd479b6b5-%s", apiPath, string(payloadJSON), digestHex)

	// AES-128-ECB 加密
	return AESEncryptECB(encryptText)
}
