package qm

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

// inputKind 表示输入被识别成的形态。
type inputKind int

const (
	kindUnknown inputKind = iota
	kindMid
	kindID
)

var (
	digitsRe = regexp.MustCompile(`^\d{1,20}$`)
	base62Re = regexp.MustCompile(`^[0-9A-Za-z]{10,20}$`)
)

// parseSongInput 解析歌曲输入。
// 判定顺序必须是 URL → 纯数字 → base62，因为纯数字也是 base62 的子集。
func parseSongInput(s string) (mid string, id int64, kind inputKind) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", 0, kindUnknown
	}

	if u, ok := parseInputURL(s); ok {
		// 查询参数优先于路径
		if v := u.Query().Get("songmid"); v != "" {
			return v, 0, kindMid
		}
		if v := lastPathSegment(u.Path); v != "" {
			if n, err := parseID(v); err == nil {
				return "", n, kindID
			}
			if base62Re.MatchString(v) {
				return v, 0, kindMid
			}
		}
		return "", 0, kindUnknown
	}

	if digitsRe.MatchString(s) {
		if n, err := parseID(s); err == nil {
			return "", n, kindID
		}
		return "", 0, kindUnknown
	}
	if base62Re.MatchString(s) {
		return s, 0, kindMid
	}
	return "", 0, kindUnknown
}

// parsePlaylistInput 解析歌单 ID，支持纯数字与多种分享链接。
func parsePlaylistInput(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", models.NewAPIError(400, "歌单 ID 为空")
	}
	if digitsRe.MatchString(s) {
		return s, nil
	}

	if u, ok := parseInputURL(s); ok {
		if v := u.Query().Get("id"); digitsRe.MatchString(v) {
			return v, nil
		}
		if v := lastPathSegment(u.Path); digitsRe.MatchString(v) {
			return v, nil
		}
	}

	// 兜底：直接从原文里抠 playlist/<id> 与 id=<id>
	if v, ok := segmentAfter(s, "playlist/", "/?#"); ok && digitsRe.MatchString(v) {
		return v, nil
	}
	if v, ok := segmentAfter(s, "id=", "&?#"); ok && digitsRe.MatchString(v) {
		return v, nil
	}

	return "", models.NewAPIError(400, "无法从输入解析歌单 ID")
}

// parseInputURL 把输入按 URL 解析。
// 允许缺 scheme（如 y.qq.com/n/ryqq/...），此时补 https://。
func parseInputURL(raw string) (*url.URL, bool) {
	if !strings.Contains(raw, "qq.com") && !strings.HasPrefix(raw, "http") {
		return nil, false
	}
	candidate := raw
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}
	u, err := url.Parse(candidate)
	if err != nil || u.Host == "" {
		return nil, false
	}
	return u, true
}

// lastPathSegment 取路径最后一段。
func lastPathSegment(p string) string {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}

// segmentAfter 取 prefix 之后、任一 stopChars 之前的内容。
func segmentAfter(s, prefix, stopChars string) (string, bool) {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return "", false
	}
	rest := s[idx+len(prefix):]
	if i := strings.IndexAny(rest, stopChars); i >= 0 {
		rest = rest[:i]
	}
	return rest, rest != ""
}

// parseID 解析数字 ID，溢出 int64 视为非法。
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
