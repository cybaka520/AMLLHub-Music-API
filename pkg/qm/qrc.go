package qm

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"strings"
)

// qrcKey 是 QRC 的 3DES 密钥（24 字节 ASCII），逆向固定值，不可改。
var qrcKey = []byte("!@#)(*$%123ZXC!@!@#)(NHL")

// decryptQRC 解密 QRC：hex → 3DES-ECB（QQ 专有实现）→ zlib。
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
	if len(cipherText)%8 != 0 {
		return "", errors.New("密文长度不是 8 的倍数")
	}

	return inflateAny(qrcTripleDESDecrypt(cipherText, qrcKey))
}

// normalizeQRC 把上游的 QRC XML 容器拆成纯文本 QRC。
//
// 形态：<QrcInfos><QrcHeadInfo/><LyricInfo><Lyric_1 LyricContent="[ti:..]\n[0,500].."/></LyricInfo></QrcInfos>
// LyricContent 里的换行可能是裸 \n，也可能是 &#10;，两种都要还原。
func normalizeQRC(s string) string {
	if !strings.Contains(s, "LyricContent=") {
		return s
	}

	const key = `LyricContent="`
	idx := strings.Index(s, key)
	if idx < 0 {
		return s
	}
	body := s[idx+len(key):]

	// 属性结束标记；找不到就退回到最后一个引号
	if end := strings.Index(body, `"/>`); end >= 0 {
		body = body[:end]
	} else if end := strings.LastIndex(body, `"`); end >= 0 {
		body = body[:end]
	}
	return xmlUnescape(body)
}

// xmlUnescape 还原歌词里常见的 XML 实体。
func xmlUnescape(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '&' {
			b.WriteByte(s[i])
			continue
		}
		end := strings.IndexByte(s[i:], ';')
		if end < 0 {
			b.WriteByte(s[i])
			continue
		}
		ent := s[i : i+end+1]
		switch ent {
		case "&#10;":
			b.WriteByte('\n')
		case "&#13;":
			b.WriteByte('\r')
		case "&#9;":
			b.WriteByte('\t')
		case "&quot;":
			b.WriteByte('"')
		case "&apos;":
			b.WriteByte('\'')
		case "&lt;":
			b.WriteByte('<')
		case "&gt;":
			b.WriteByte('>')
		case "&amp;":
			b.WriteByte('&')
		default:
			if n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(ent, "&#"), ";")); err == nil &&
				strings.HasPrefix(ent, "&#") {
				b.WriteRune(rune(n))
			} else {
				b.WriteString(ent)
			}
		}
		i += end
	}
	return b.String()
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
