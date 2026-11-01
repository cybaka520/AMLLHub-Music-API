package qm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

// songDetailParam 是详情接口参数。
// song_mid / song_id 都不能加 omitempty：省掉键会导致部分网关参数校验失败。
type songDetailParam struct {
	SongType int    `json:"song_type"`
	SongMid  string `json:"song_mid"`
	SongID   int64  `json:"song_id"`
}

// GetSong 按 mid、数字 id 或分享链接获取歌曲信息。
func (c *Client) GetSong(ctx context.Context, idOrMid string) (*Song, error) {
	raw := idOrMid
	mid, id, kind := parseSongInput(raw)
	if kind == kindUnknown {
		return nil, models.NewAPIError(400, "无法解析歌曲 ID 或 mid: "+raw)
	}

	song, err := c.getSongDetail(ctx, mid, id)
	if err == nil {
		return song, nil
	}
	if !errors.Is(err, ErrSongNotFound) {
		return nil, err
	}

	// 兜底：id 与 mid 互换再试一次（10 位以上的纯数字同时也是合法 mid）。
	if !base62Re.MatchString(raw) {
		return nil, ErrSongNotFound
	}
	if kind == kindID {
		if song, err := c.getSongDetail(ctx, raw, 0); err == nil {
			return song, nil
		}
	} else if n, perr := strconv.ParseInt(raw, 10, 64); perr == nil {
		if song, err := c.getSongDetail(ctx, "", n); err == nil {
			return song, nil
		}
	}
	return nil, ErrSongNotFound
}

// SongIDByMid 把 mid 转成数字 id。内部就是一次详情请求，两个字段都在响应里。
func (c *Client) SongIDByMid(ctx context.Context, mid string) (int64, error) {
	song, err := c.GetSong(ctx, mid)
	if err != nil {
		return 0, err
	}
	if song.ID == 0 {
		return 0, ErrSongNotFound
	}
	return song.ID, nil
}

// SongMidByID 把数字 id 转成 mid。
func (c *Client) SongMidByID(ctx context.Context, id int64) (string, error) {
	if id <= 0 {
		return "", models.NewAPIError(400, "id 必须大于 0")
	}
	song, err := c.GetSong(ctx, strconv.FormatInt(id, 10))
	if err != nil {
		return "", err
	}
	if song.Mid == "" {
		return "", ErrSongNotFound
	}
	return song.Mid, nil
}

// getSongDetail 发起详情请求。
func (c *Client) getSongDetail(ctx context.Context, mid string, id int64) (*Song, error) {
	param := songDetailParam{SongType: 0, SongMid: mid, SongID: id}

	data, err := c.call(ctx, "music.pf_song_detail_svr", "get_song_detail_yqq", param, defaultCall())
	if err != nil {
		return nil, err
	}

	var resp struct {
		TrackInfo *rawSong `json:"track_info"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("qm: 解析歌曲详情失败: %w", err)
	}
	if resp.TrackInfo == nil || resp.TrackInfo.Mid == "" {
		return nil, ErrSongNotFound
	}

	song := normalizeSong(*resp.TrackInfo)
	return &song, nil
}
