package qm

import (
	"strconv"
	"strings"
)

// Artist 歌手
type Artist struct {
	ID   int64  `json:"id"`
	Mid  string `json:"mid"`
	Name string `json:"name"`
}

// Album 专辑
type Album struct {
	ID    int64  `json:"id"`
	Mid   string `json:"mid"`
	Name  string `json:"name"`
	Cover string `json:"cover"` // T002R800x800，缺失为空
}

// Pay 付费标记
type Pay struct {
	PayAlbum int `json:"payAlbum"` // 1 = 数字专辑付费
	PayPlay  int `json:"payPlay"`  // 1 = 会员播放
}

// Sizes 各档位文件体积（字节），0 表示该档不可用
type Sizes struct {
	HiRes           int64 `json:"hiRes"`
	Flac            int64 `json:"flac"`
	Mp3_320         int64 `json:"mp3_320"`
	Mp3_128         int64 `json:"mp3_128"`
	Ape             int64 `json:"ape"`
	Ogg             int64 `json:"ogg"`
	HiResSampleRate int   `json:"hiResSampleRate"`
	HiResBitDepth   int   `json:"hiResBitDepth"`
}

// Song 归一化歌曲信息
type Song struct {
	ID         int64    `json:"id"`
	Mid        string   `json:"mid"`
	MediaMid   string   `json:"mediaMid"` // 直链 filename 用
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	ArtistName string   `json:"artistName"` // "A / B"
	Album      Album    `json:"album"`
	Disc       int      `json:"disc"` // 从 1 开始
	DurationMs int64    `json:"durationMs"`
	Duration   string   `json:"duration"`   // "04:12"
	Cover      string   `json:"cover"`      // T002R300x300
	CoverLarge string   `json:"coverLarge"` // T002R800x800
	Pay        Pay      `json:"pay"`
	Sizes      Sizes    `json:"sizes"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Code  int    `json:"code"`
	Total int    `json:"total"`
	Data  []Song `json:"data"`
	Time  string `json:"time"`
}

// MusicResponse 单曲解析响应
type MusicResponse struct {
	Code        int    `json:"code"`
	ID          int64  `json:"id"`
	Mid         string `json:"mid"`
	Name        string `json:"name"`
	ArtistName  string `json:"artistName"`
	AlbumName   string `json:"albumName"`
	Cover       string `json:"cover"`
	Level       string `json:"level"`       // 请求档位
	ActualLevel string `json:"actualLevel"` // 实际命中档位（可能降级）
	Format      string `json:"format"`      // flac / mp3 / m4a
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	Trial       bool   `json:"trial"` // 恒 false，本期不做试听片段识别
	Lyric       string `json:"lyric"` // LRC 文本
	QRC         string `json:"qrc"`   // QRC 原文（逐字）
	Translation string `json:"translation"`
	Romaji      string `json:"romaji"`
	Time        string `json:"time"`
}

// PlaylistDetail 歌单详情
type PlaylistDetail struct {
	ID          string `json:"id"` // dissid，字符串
	Name        string `json:"name"`
	Cover       string `json:"cover"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
	PlayCount   int64  `json:"playCount"`
	TrackCount  int    `json:"trackCount"`
	Songs       []Song `json:"songs"`
}

// PlaylistResponse 歌单响应
type PlaylistResponse struct {
	Code     int            `json:"code"`
	Playlist PlaylistDetail `json:"playlist"`
	Time     string         `json:"time"`
}

// Lyric 歌词
type Lyric struct {
	LRC         string `json:"lrc"`
	QRC         string `json:"qrc"`
	Translation string `json:"translation"`
	Romaji      string `json:"romaji"`
}

// flexInt 兼容上游把数字写成字符串的情况（如 expired_at、musicid）。
type flexInt int64

// UnmarshalJSON 同时接受数字、数字字符串与 null。
func (f *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*f = flexInt(v)
	return nil
}

// int64 转为标准类型。
func (f flexInt) int64() int64 { return int64(f) }
