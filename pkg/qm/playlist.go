package qm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/utils"
)

// rawPlaylistSong 是歌单接口的歌曲结构。
// 字段名与搜索/详情接口完全不同（strMediaMid、size128…），必须单独建结构体。
type rawPlaylistSong struct {
	SongID    int64       `json:"songid"`
	SongMid   string      `json:"songmid"`
	MediaMid  string      `json:"strMediaMid"`
	SongName  string      `json:"songname"`
	Singer    []rawSinger `json:"singer"`
	AlbumName string      `json:"albumname"`
	AlbumMid  string      `json:"albummid"`
	Interval  int         `json:"interval"`
	IndexCD   int         `json:"index_cd"`
	Size128   int64       `json:"size128"`
	Size320   int64       `json:"size320"`
	SizeAPE   int64       `json:"sizeape"`
	SizeFLAC  int64       `json:"sizeflac"`
	SizeOGG   int64       `json:"sizeogg"`
	Pay       struct {
		PayAlbum int `json:"payalbum"`
		PayPlay  int `json:"payplay"`
	} `json:"pay"`
}

// rawPlaylist 是歌单接口的歌单结构。
type rawPlaylist struct {
	DissID   flexInt           `json:"disstid"`
	DissName string            `json:"dissname"`
	Logo     string            `json:"logo"`
	Desc     string            `json:"desc"`
	Nickname string            `json:"nickname"`
	VisitNum flexInt           `json:"visitnum"`
	SongNum  int               `json:"songnum"`
	SongList []rawPlaylistSong `json:"songlist"`
}

var (
	jsonpPrefixRe = regexp.MustCompile(`^[^(]*\(`)
	jsonpSuffixRe = regexp.MustCompile(`\)\s*;?\s*$`)
)

// ParsePlaylist 解析歌单，支持歌单 ID 与分享链接。
func (c *Client) ParsePlaylist(ctx context.Context, idOrURL string) (*PlaylistResponse, error) {
	dissid, err := parsePlaylistInput(idOrURL)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf(
		"%s?type=1&json=1&utf8=1&onlysonglist=0&platform=yqq&needNewCode=0&disstid=%s",
		hostPlaylist, url.QueryEscape(dissid),
	)
	body, err := c.getWeb(ctx, reqURL)
	if err != nil {
		return nil, err
	}

	jsonText, err := unwrapJSONP(string(body))
	if err != nil {
		return nil, err
	}

	var raw struct {
		Code   int           `json:"code"`
		CDList []rawPlaylist `json:"cdlist"`
	}
	if err := json.Unmarshal([]byte(jsonText), &raw); err != nil {
		return nil, fmt.Errorf("qm: 解析歌单响应失败: %w", err)
	}
	if raw.Code != 0 {
		return nil, models.NewAPIError(raw.Code, "歌单请求失败")
	}
	if len(raw.CDList) == 0 {
		return nil, models.NewAPIError(404, "歌单不存在")
	}

	pl := raw.CDList[0]
	songs := make([]Song, 0, len(pl.SongList))
	for _, s := range pl.SongList {
		songs = append(songs, normalizePlaylistSong(s))
	}

	// 服务端可能截断 songlist，曲目数一律以实际长度为准。
	trackCount := pl.SongNum
	if trackCount > len(songs) || trackCount == 0 {
		trackCount = len(songs)
	}

	return &PlaylistResponse{
		Code: 200,
		Playlist: PlaylistDetail{
			ID:          strconv.FormatInt(pl.DissID.int64(), 10),
			Name:        pl.DissName,
			Cover:       pl.Logo,
			Description: pl.Desc,
			Creator:     pl.Nickname,
			PlayCount:   pl.VisitNum.int64(),
			TrackCount:  trackCount,
			Songs:       songs,
		},
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// normalizePlaylistSong 把歌单条目转成统一模型。
func normalizePlaylistSong(raw rawPlaylistSong) Song {
	artists := make([]Artist, 0, len(raw.Singer))
	names := make([]string, 0, len(raw.Singer))
	for _, s := range raw.Singer {
		artists = append(artists, Artist{ID: s.ID, Mid: s.Mid, Name: s.Name})
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}

	durationMs := int64(raw.Interval) * 1000

	return Song{
		ID:         raw.SongID,
		Mid:        raw.SongMid,
		MediaMid:   raw.MediaMid,
		Name:       raw.SongName,
		Artists:    artists,
		ArtistName: strings.Join(names, " / "),
		Album: Album{
			Mid:   raw.AlbumMid,
			Name:  raw.AlbumName,
			Cover: coverURL(raw.AlbumMid, 800),
		},
		Disc:       raw.IndexCD + 1,
		DurationMs: durationMs,
		Duration:   utils.FormatDuration(durationMs),
		Cover:      coverURL(raw.AlbumMid, 300),
		CoverLarge: coverURL(raw.AlbumMid, 800),
		Pay: Pay{
			PayAlbum: raw.Pay.PayAlbum,
			PayPlay:  raw.Pay.PayPlay,
		},
		Sizes: Sizes{
			Flac:    raw.SizeFLAC,
			Mp3_320: raw.Size320,
			Mp3_128: raw.Size128,
			Ape:     raw.SizeAPE,
			Ogg:     raw.SizeOGG,
		},
	}
}

// unwrapJSONP 剥掉 jsonCallback(...) 外壳。
// 兼容括号前后有空格、末尾有无分号的情况。
func unwrapJSONP(s string) (string, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "{") {
		return s, nil
	}
	loc := jsonpPrefixRe.FindStringIndex(s)
	if loc == nil {
		return "", errors.New("qm: 歌单响应不是合法的 JSONP")
	}
	s = jsonpSuffixRe.ReplaceAllString(s[loc[1]:], "")
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return "", errors.New("qm: 歌单响应不是合法的 JSON")
	}
	return s, nil
}

// getWeb 发起 Web 端 GET 请求（歌单老接口用，无需 Cookie）。
func (c *Client) getWeb(ctx context.Context, rawURL string) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uaWeb)
	req.Header.Set("Referer", refererWeb)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, &networkError{err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRespBody))
	if err != nil {
		return nil, &networkError{err: err}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &statusError{Status: resp.StatusCode, Body: truncateBody(body)}
	}
	return body, nil
}
