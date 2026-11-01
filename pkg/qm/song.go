package qm

import (
	"fmt"
	"strings"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/utils"
)

// rawSinger 是上游的歌手结构。
type rawSinger struct {
	ID   int64  `json:"id"`
	Mid  string `json:"mid"`
	Name string `json:"name"`
}

// rawAlbum 是上游的专辑结构，不同接口字段名略有差异。
type rawAlbum struct {
	ID    int64  `json:"id"`
	Mid   string `json:"mid"`
	PMID  string `json:"pmid"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

// rawFile 是上游的文件结构，体积字段均为字节。
type rawFile struct {
	MediaMid      string  `json:"media_mid"`
	Size128MP3    int64   `json:"size_128mp3"`
	Size320MP3    int64   `json:"size_320mp3"`
	SizeFLAC      int64   `json:"size_flac"`
	SizeHires     int64   `json:"size_hires"`
	SizeAPE       int64   `json:"size_ape"`
	Size192OGG    int64   `json:"size_192ogg"`
	HiresSample   int     `json:"hires_sample"`
	HiresBitdepth int     `json:"hires_bitdepth"`
	SizeNew       []int64 `json:"size_new"`
}

// rawPay 是上游的付费标记。
type rawPay struct {
	PayMonth   int `json:"pay_month"`
	PriceAlbum int `json:"price_album"`
	PayPlay    int `json:"pay_play"`
}

// rawSong 是搜索与详情共用的上游歌曲结构。
type rawSong struct {
	ID       int64       `json:"id"`
	Mid      string      `json:"mid"`
	Title    string      `json:"title"`
	Name     string      `json:"name"`
	Interval int         `json:"interval"`
	IndexCD  int         `json:"index_cd"`
	Singer   []rawSinger `json:"singer"`
	Album    rawAlbum    `json:"album"`
	File     rawFile     `json:"file"`
	Pay      rawPay      `json:"pay"`
}

// normalizeSong 把上游结构归一化为对外模型。
func normalizeSong(raw rawSong) Song {
	albumMid := firstNonEmpty(raw.Album.Mid, raw.Album.PMID)

	artists := make([]Artist, 0, len(raw.Singer))
	names := make([]string, 0, len(raw.Singer))
	for _, s := range raw.Singer {
		artists = append(artists, Artist{ID: s.ID, Mid: s.Mid, Name: s.Name})
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}

	durationMs := int64(raw.Interval) * 1000

	song := Song{
		ID:         raw.ID,
		Mid:        raw.Mid,
		MediaMid:   raw.File.MediaMid,
		Name:       firstNonEmpty(raw.Title, raw.Name),
		Artists:    artists,
		ArtistName: strings.Join(names, " / "),
		Album: Album{
			ID:    raw.Album.ID,
			Mid:   albumMid,
			Name:  firstNonEmpty(raw.Album.Name, raw.Album.Title),
			Cover: coverURL(albumMid, 800),
		},
		Disc:       raw.IndexCD + 1,
		DurationMs: durationMs,
		Duration:   utils.FormatDuration(durationMs),
		Cover:      coverURL(albumMid, 300),
		CoverLarge: coverURL(albumMid, 800),
		Pay: Pay{
			PayAlbum: payAlbumFlag(raw.Pay),
			PayPlay:  raw.Pay.PayPlay,
		},
		Sizes: Sizes{
			HiRes:           hiResSize(raw.File),
			Flac:            raw.File.SizeFLAC,
			Mp3_320:         raw.File.Size320MP3,
			Mp3_128:         raw.File.Size128MP3,
			Ape:             raw.File.SizeAPE,
			Ogg:             raw.File.Size192OGG,
			HiResSampleRate: raw.File.HiresSample,
			HiResBitDepth:   raw.File.HiresBitdepth,
		},
	}
	return song
}

// payAlbumFlag 判断是否数字专辑付费：未单独售卖但专辑价格大于 0。
func payAlbumFlag(p rawPay) int {
	if p.PayMonth == 0 && p.PriceAlbum > 0 {
		return 1
	}
	return 0
}

// hiResSize hires 体积在部分接口走 size_new[0]。
func hiResSize(f rawFile) int64 {
	if f.SizeHires != 0 {
		return f.SizeHires
	}
	if len(f.SizeNew) > 0 {
		return f.SizeNew[0]
	}
	return 0
}

// coverURL 拼接封面地址，albumMid 缺失返回空串。
func coverURL(albumMid string, size int) string {
	if albumMid == "" {
		return ""
	}
	return fmt.Sprintf(coverURLTemplate, size, size, albumMid)
}
