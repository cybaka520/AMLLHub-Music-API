package api

import (
	"encoding/json"
	"testing"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

func TestClamp(t *testing.T) {
	cases := []struct{ v, min, max, want int }{
		{5, 1, 100, 5},
		{0, 1, 100, 1},
		{200, 1, 100, 100},
		{1, 1, 1, 1},
	}
	for _, c := range cases {
		if got := clamp(c.v, c.min, c.max); got != c.want {
			t.Errorf("clamp(%d,%d,%d) = %d, want %d", c.v, c.min, c.max, got, c.want)
		}
	}
}

func TestReplaceHTTPtoHTTPS(t *testing.T) {
	if got := replaceHTTPtoHTTPS("http://example.com/a"); got != "https://example.com/a" {
		t.Errorf("got %q", got)
	}
	if got := replaceHTTPtoHTTPS("https://example.com/a"); got != "https://example.com/a" {
		t.Errorf("已是 https 不应改变: %q", got)
	}
}

func TestSearchFormatResponse(t *testing.T) {
	// 构造一个模拟的搜索原始响应
	raw := &searchRawResponse{Code: 200}
	// 通过 JSON 反序列化构造嵌套匿名结构
	body := []byte(`{
		"code": 200,
		"data": {
			"resources": [
				{
					"resourceType": "song",
					"baseInfo": {
						"simpleSongData": {
							"id": 123,
							"name": "测试歌曲",
							"fee": 8,
							"pop": 95.5,
							"dt": 180000,
							"publishTime": 1609459200000,
							"al": {"id": 1, "name": "测试专辑", "picUrl": "http://p.com/1.jpg"},
							"ar": [{"id": 10, "name": "歌手A"}, {"id": 11, "name": "歌手B"}],
							"alia": ["别名1", "别名2"]
						}
					},
					"extInfo": {"starCount": 5}
				},
				{"resourceType": "album"},
				{"resourceType": "song", "baseInfo": {"simpleSongData": {"id": 456, "name": "第二首", "al": {"picUrl": ""}}}}
			]
		}
	}`)
	// 反序列化到 raw
	if err := json.Unmarshal(body, raw); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	s := &SearchAPI{}
	res := s.formatResponse(raw)

	if res.Code != 200 {
		t.Fatalf("Code = %d", res.Code)
	}
	if len(res.Data) != 2 {
		t.Fatalf("应过滤出 2 首 song, got %d", len(res.Data))
	}
	first := res.Data[0]
	if first.ID != 123 || first.Name != "测试歌曲" {
		t.Errorf("第一首: %+v", first)
	}
	if len(first.Artists) != 2 || first.Artists[0].Name != "歌手A" {
		t.Errorf("歌手: %+v", first.Artists)
	}
	if first.Album.Name != "测试专辑" {
		t.Errorf("专辑: %+v", first.Album)
	}
	if first.Album.PublishTime != "2021-01-01" {
		t.Errorf("发布时间 = %q", first.Album.PublishTime)
	}
	if first.Alias != "别名1" {
		t.Errorf("别名 = %q", first.Alias)
	}
	if first.Duration != "03:00" {
		t.Errorf("时长 = %q", first.Duration)
	}
	if first.StarCount != 5 {
		t.Errorf("StarCount = %d", first.StarCount)
	}
}

func TestSimplifyPlaylistDataWithTracks(t *testing.T) {
	p := &PlaylistAPI{}
	raw := &playlistRawResponse{Code: 200}
	raw.Playlist.ID = 777
	raw.Playlist.Name = "我的歌单"
	raw.Playlist.Creator.Nickname = "用户A"

	tracks := []models.Track{
		{Name: "歌1", ID: 1},
		{Name: "歌2", ID: 2},
	}
	res := p.simplifyPlaylistDataWithTracks(raw, tracks)

	if res.Code != 200 {
		t.Fatalf("Code = %d", res.Code)
	}
	if res.Playlist.ID != 777 || res.Playlist.Name != "我的歌单" {
		t.Errorf("歌单基础信息错误: %+v", res.Playlist)
	}
	if res.Playlist.TrackCount != 2 {
		t.Errorf("TrackCount = %d, want 2", res.Playlist.TrackCount)
	}
	if len(res.Playlist.Tracks) != 2 {
		t.Errorf("Tracks 数量 = %d", len(res.Playlist.Tracks))
	}
}
