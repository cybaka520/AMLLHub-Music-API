package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var numericRegex = regexp.MustCompile(`^\d+$`)

// ExtractSongID 从输入中提取歌曲ID
// 支持纯数字、music.163.com URL、163cn.tv 短链接
// 注意：短链接需要调用方先获取重定向URL，此函数不处理短链接跳转
func ExtractSongID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入为空")
	}

	// 纯数字直接返回
	if numericRegex.MatchString(input) {
		return input, nil
	}

	// 处理 music.163.com URL
	if strings.Contains(input, "music.163.com") {
		// 尝试从查询参数解析
		u, err := url.Parse(input)
		if err == nil && u.Query().Get("id") != "" {
			return u.Query().Get("id"), nil
		}

		// 字符串查找 id=
		if idx := strings.Index(input, "id="); idx != -1 {
			idPart := input[idx+3:]
			if ampIdx := strings.Index(idPart, "&"); ampIdx != -1 {
				return idPart[:ampIdx], nil
			}
			return idPart, nil
		}
	}

	// 无法解析，假定输入已经是ID
	return input, nil
}

// ExtractPlaylistID 从输入中提取歌单ID
// 支持纯数字、带 id= 参数的URL、/playlist/{id} 路径的URL
func ExtractPlaylistID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入为空")
	}

	// 纯数字直接返回
	if numericRegex.MatchString(input) {
		return input, nil
	}

	// 从 URL 查询参数提取 id=数字
	if idx := strings.Index(input, "id="); idx != -1 {
		idPart := input[idx+3:]
		if ampIdx := strings.Index(idPart, "&"); ampIdx != -1 {
			return idPart[:ampIdx], nil
		}
		return idPart, nil
	}

	// 从 /playlist/{id} 路径提取
	if idx := strings.Index(input, "playlist/"); idx != -1 {
		idPart := input[idx+9:]
		if slashIdx := strings.Index(idPart, "/"); slashIdx != -1 {
			return idPart[:slashIdx], nil
		}
		return idPart, nil
	}

	return "", fmt.Errorf("无法从输入中提取有效的歌单ID")
}

// ParseCookie 将 cookie 字符串解析为 map
func ParseCookie(text string) map[string]string {
	cookies := make(map[string]string)
	text = strings.TrimSpace(text)
	if text == "" {
		return cookies
	}

	for _, item := range strings.Split(text, ";") {
		if eqIdx := strings.Index(item, "="); eqIdx != -1 {
			key := strings.TrimSpace(item[:eqIdx])
			value := strings.TrimSpace(item[eqIdx+1:])
			cookies[key] = value
		}
	}

	return cookies
}

// BuildCookieString 将 cookie map 转为字符串
func BuildCookieString(cookies map[string]string) string {
	if len(cookies) == 0 {
		return ""
	}

	var parts []string
	for key, value := range cookies {
		parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}
	return strings.Join(parts, "; ")
}
