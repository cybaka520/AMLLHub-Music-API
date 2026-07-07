package models

// SearchResponse 搜索响应
type SearchResponse struct {
	Code  int    `json:"code"`
	Data  []Song `json:"data"`
	Time  string `json:"time"`
	Error string `json:"error,omitempty"`
}

// MusicResponse 单曲解析响应
type MusicResponse struct {
	Code       int    `json:"code"`
	Name       string `json:"name"`
	Pic        string `json:"pic"`
	ArtistName string `json:"ar_name"`
	AlbumName  string `json:"al_name"`
	Level      string `json:"level"`
	Size       string `json:"size"`
	URL        string `json:"url"`
	Lyric      string `json:"lyric"`
	TLyric     string `json:"tlyric"`
	Error      string `json:"error,omitempty"`
}

// PlaylistResponse 歌单解析响应
type PlaylistResponse struct {
	Code     int            `json:"code"`
	Playlist PlaylistDetail `json:"playlist"`
	Time     string         `json:"time"`
	Error    string         `json:"error,omitempty"`
}
