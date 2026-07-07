package models

// Creator 歌单创建者
type Creator struct {
	UserID    int64  `json:"userId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

// Track 歌单中的歌曲
type Track struct {
	Name string   `json:"name"`
	ID   int64    `json:"id"`
	Ar   []Artist `json:"ar"`
	Al   Album    `json:"al"`
	Dt   int64    `json:"dt"`
	Fee  int      `json:"fee"`
}

// PlaylistDetail 歌单详情
type PlaylistDetail struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	CoverImgURL string   `json:"coverImgUrl"`
	UserID      int64    `json:"userId"`
	CreateTime  int64    `json:"createTime"`
	TrackCount  int      `json:"trackCount"`
	Creator     Creator  `json:"creator"`
	Tracks      []Track  `json:"tracks"`
}
