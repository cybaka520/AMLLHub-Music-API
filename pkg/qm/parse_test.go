package qm

import "testing"

// TestParseSongInput 覆盖 mid、数字 id 与六种链接形态。
func TestParseSongInput(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantMid string
		wantID  int64
		wantKnd inputKind
	}{
		{
			name: "纯 mid", in: "0039MnYb0qxYhV",
			wantMid: "0039MnYb0qxYhV", wantKnd: kindMid,
		},
		{
			name: "纯数字 id", in: "123456789",
			wantID: 123456789, wantKnd: kindID,
		},
		{
			name: "songDetail 链接", in: "https://y.qq.com/n/ryqq/songDetail/0039MnYb0qxYhV",
			wantMid: "0039MnYb0qxYhV", wantKnd: kindMid,
		},
		{
			name: "缺 scheme 的链接", in: "y.qq.com/n/ryqq/songDetail/0039MnYb0qxYhV",
			wantMid: "0039MnYb0qxYhV", wantKnd: kindMid,
		},
		{
			name: "song.html 查询参数", in: "https://i.y.qq.com/n2/m/share/details/song.html?songmid=0039MnYb0qxYhV",
			wantMid: "0039MnYb0qxYhV", wantKnd: kindMid,
		},
		{
			name: "c.y.qq.com 查询参数", in: "https://c.y.qq.com/base/fcgi-bin/fcg_music_express_mobile3.fcg?songmid=0039MnYb0qxYhV&format=json",
			wantMid: "0039MnYb0qxYhV", wantKnd: kindMid,
		},
		{
			name: "空串", in: " ",
			wantKnd: kindUnknown,
		},
		{
			name: "无法识别", in: "hello",
			wantKnd: kindUnknown,
		},
		{
			name: "超长输入", in: "0123456789012345678901",
			wantKnd: kindUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mid, id, kind := parseSongInput(tc.in)
			if kind != tc.wantKnd || mid != tc.wantMid || id != tc.wantID {
				t.Fatalf("got (%q, %d, %d), want (%q, %d, %d)",
					mid, id, kind, tc.wantMid, tc.wantID, tc.wantKnd)
			}
		})
	}
}

// TestParsePlaylistInput 覆盖歌单 ID 与三种分享链接。
func TestParsePlaylistInput(t *testing.T) {
	ok := []struct {
		name string
		in   string
		want string
	}{
		{name: "纯数字", in: "8672369", want: "8672369"},
		{name: "playlist 路径", in: "https://y.qq.com/n/ryqq/playlist/8672369", want: "8672369"},
		{name: "带查询参数", in: "https://y.qq.com/n/ryqq/playlist/8672369?from=share", want: "8672369"},
		{name: "taoge 链接", in: "https://i.y.qq.com/n2/m/share/details/taoge.html?id=8672369", want: "8672369"},
		{name: "缺 scheme", in: "y.qq.com/n/ryqq/playlist/8672369", want: "8672369"},
	}

	for _, tc := range ok {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePlaylistInput(tc.in)
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}

	bad := []string{"", "https://y.qq.com/n/ryqq/songDetail/0039MnYb0qxYhV", "abc"}
	for _, in := range bad {
		if _, err := parsePlaylistInput(in); err == nil {
			t.Fatalf("%q 应返回错误", in)
		}
	}
}
