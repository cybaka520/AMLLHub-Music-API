package qm

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// zzc 签名用的下标与掩码，均为逆向得到的固定值。
var (
	zzcPrefixIdx = [7]int{23, 14, 6, 36, 16, 7, 19}
	zzcSuffixIdx = [8]int{16, 1, 32, 12, 19, 27, 8, 5}
	zzcMask      = [20]byte{
		89, 39, 179, 150, 218, 82, 58, 252, 177, 52,
		186, 123, 120, 64, 242, 133, 143, 161, 121, 179,
	}
)

// signZZC 计算搜索网关的 sign 参数。
// payload 必须是最终发送的 JSON 字节，不能是重新序列化的结果。
func signZZC(payload []byte) string {
	sum := sha1.Sum(payload)
	h := hex.EncodeToString(sum[:]) // 40 位小写 hex

	var sb strings.Builder
	sb.WriteString("zzc")
	for _, i := range zzcPrefixIdx {
		sb.WriteByte(h[i])
	}

	raw := make([]byte, len(zzcMask))
	for i, m := range zzcMask {
		v, _ := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
		raw[i] = m ^ byte(v)
	}
	mid := base64.StdEncoding.EncodeToString(raw)
	mid = strings.NewReplacer("/", "", "+", "", "=", "").Replace(mid)
	sb.WriteString(mid)

	for _, i := range zzcSuffixIdx {
		sb.WriteByte(h[i])
	}
	return strings.ToLower(sb.String())
}

// hash33 等价于 JS 的 h = ((h << 5) + h + charCode) >>> 0，返回 h & 0x7fffffff。
// 仅用于 ASCII 输入（musickey / qrsig）。
func hash33(s string, seed uint32) uint32 {
	h := seed
	for i := 0; i < len(s); i++ {
		h = (h << 5) + h + uint32(s[i])
	}
	return h & 0x7fffffff
}

// newGUID 生成 32 位小写 hex（UUID v4 去横线），仅用于直链接口的 guid 参数。
func newGUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		binary.BigEndian.PutUint64(b[:8], uint64(time.Now().UnixNano()))
		binary.BigEndian.PutUint64(b[8:], uint64(time.Now().UnixNano()>>1))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:])
}
