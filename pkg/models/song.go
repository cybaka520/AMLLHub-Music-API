// Package models 定义网易云音乐API的数据模型
package models

// Artist 歌手
type Artist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Album 专辑
type Album struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PicURL      string `json:"picUrl"`
	PublishTime string `json:"publishTime,omitempty"`
}

// Song 歌曲基本信息
type Song struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Fee       int      `json:"fee"`
	Pop       float64  `json:"pop"`
	StarCount int      `json:"starCount"`
	PicURL    string   `json:"picUrl"`
	Artists   []Artist `json:"artists"`
	Album     Album    `json:"album"`
	Duration  string   `json:"duration"`
	Alias     string   `json:"alias,omitempty"`
}
