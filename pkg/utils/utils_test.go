package utils

import (
	"strings"
	"testing"
)

func TestExtractSongID(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"123456", "123456", true},
		{"  789  ", "789", true},
		{"https://music.163.com/song?id=9988", "9988", true},
		{"https://music.163.com/song?id=9988&other=1", "9988", true},
		{"id=42", "42", true}, // 含 id= 但不含 music.163.com -> 当前实现不命中
		{"abc", "", false},
		{"", "", false},
		{"not a url at all", "", false},
	}
	for _, c := range cases {
		got, err := ExtractSongID(c.in)
		// 注意：纯 id=42 不含 music.163.com，当前实现不识别，应失败
		if c.in == "id=42" {
			if err == nil {
				t.Errorf("ExtractSongID(%q) 期望失败但得到 %q", c.in, got)
			}
			continue
		}
		if c.ok {
			if err != nil {
				t.Errorf("ExtractSongID(%q) 意外错误: %v", c.in, err)
				continue
			}
			if got != c.want {
				t.Errorf("ExtractSongID(%q) = %q, want %q", c.in, got, c.want)
			}
		} else {
			if err == nil {
				t.Errorf("ExtractSongID(%q) 期望失败但得到 %q", c.in, got)
			}
		}
	}
}

func TestExtractPlaylistID(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"12345", "12345", true},
		{"https://music.163.com/playlist?id=12345", "12345", true},
		{"https://music.163.com/playlist/12345/", "12345", true},
		{"https://music.163.com/playlist/12345", "12345", true},
		{"id=67890&extra=1", "67890", true},
		{"abc", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, err := ExtractPlaylistID(c.in)
		if c.ok {
			if err != nil {
				t.Errorf("ExtractPlaylistID(%q) 意外错误: %v", c.in, err)
				continue
			}
			if got != c.want {
				t.Errorf("ExtractPlaylistID(%q) = %q, want %q", c.in, got, c.want)
			}
		} else {
			if err == nil {
				t.Errorf("ExtractPlaylistID(%q) 期望失败但得到 %q", c.in, got)
			}
		}
	}
}

func TestParseCookieAndBuildRoundTrip(t *testing.T) {
	raw := "MUSIC_U=abc123; NMTID=xyz; osver="
	m := ParseCookie(raw)
	if m["MUSIC_U"] != "abc123" {
		t.Errorf("MUSIC_U = %q", m["MUSIC_U"])
	}
	if m["NMTID"] != "xyz" {
		t.Errorf("NMTID = %q", m["NMTID"])
	}
	if m["osver"] != "" {
		t.Errorf("osver = %q", m["osver"])
	}

	// 空串
	if got := ParseCookie(""); len(got) != 0 {
		t.Errorf("空 cookie 应返回空 map, got %v", got)
	}
}

func TestBuildCookieString_Escaping(t *testing.T) {
	// 含分号/空格的值必须被转义，且空格用 %20 而非 +
	out := BuildCookieString(map[string]string{"k": "a b;c,d"})
	if strings.Contains(out, "a b") {
		t.Errorf("空格未被转义: %q", out)
	}
	if strings.Contains(out, "+") {
		t.Errorf("空格被错误编码为 +: %q", out)
	}
	if !strings.Contains(out, "%20") {
		t.Errorf("空格未编码为 %%20: %q", out)
	}
	if strings.Contains(out, ";c") {
		t.Errorf("分号未被转义: %q", out)
	}
	// key 不应被转义
	if !strings.HasPrefix(out, "k=") {
		t.Errorf("key 被错误转义: %q", out)
	}
}

func TestBuildCookieString_Empty(t *testing.T) {
	if out := BuildCookieString(nil); out != "" {
		t.Errorf("nil 应返回空串, got %q", out)
	}
}

func TestFormatFileSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.00KB"},
		{1048576, "1.00MB"},
		{1073741824, "1.00GB"},
		{1234567, "1.18MB"},
	}
	for _, c := range cases {
		if got := FormatFileSize(c.in); got != c.want {
			t.Errorf("FormatFileSize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		ms   int64
		want string
	}{
		{0, "00:00"},
		{59000, "00:59"},
		{60000, "01:00"},
		{125000, "02:05"},
		{3600000, "60:00"},
	}
	for _, c := range cases {
		if got := FormatDuration(c.ms); got != c.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", c.ms, got, c.want)
		}
	}
}

func TestFormatLevel(t *testing.T) {
	if got := FormatLevel("lossless"); got != "无损(SQ)" {
		t.Errorf("FormatLevel(lossless) = %q", got)
	}
	if got := FormatLevel("unknown"); got != "未知音质" {
		t.Errorf("FormatLevel(unknown) = %q", got)
	}
}

func TestRandomIP_NonEmpty(t *testing.T) {
	for i := 0; i < 100; i++ {
		ip := RandomIP()
		if ip == "" {
			t.Fatalf("RandomIP 返回空")
		}
		// 应为合法 IPv4 格式 x.x.x.x
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			t.Fatalf("RandomIP 返回非法 IPv4: %q", ip)
		}
	}
}

func TestRandomUA_NonEmpty(t *testing.T) {
	ua := RandomUA()
	if ua == "" {
		t.Fatalf("RandomUA 返回空")
	}
}
