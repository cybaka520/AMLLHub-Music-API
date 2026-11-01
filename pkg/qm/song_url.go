package qm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

// vkeyParam 是直链接口参数。
type vkeyParam struct {
	GUID     string   `json:"guid"`
	SongMid  []string `json:"songmid"`
	Filename []string `json:"filename"`
	SongType []int    `json:"songtype"`
	Uin      string   `json:"uin"`
	Ctx      int      `json:"ctx"`
}

// ParseMusic 解析单曲：详情 + 直链 + 歌词。
//
// level 取 hi-res / lossless / hq / sq / lq，空则按 hq。
// 歌词失败不影响主流程。
func (c *Client) ParseMusic(ctx context.Context, idOrMid, level string) (*MusicResponse, error) {
	level = strings.TrimSpace(level)
	if level == "" {
		level = defaultLevel
	}
	if levelIndex(level) < 0 {
		return nil, models.NewAPIError(400, "不支持的音质: "+level)
	}

	song, err := c.GetSong(ctx, idOrMid)
	if err != nil {
		return nil, err
	}

	var (
		urlStr string
		actual string
		format string
		urlErr error
		lyric  *Lyric
		wg     sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		urlStr, actual, format, urlErr = c.resolveURL(ctx, song.Mid, song.MediaMid, level)
	}()
	go func() {
		defer wg.Done()
		lyric, _ = c.GetLyric(ctx, song.ID, song.Name, song.ArtistName, song.Album.Name, int(song.DurationMs/1000))
	}()
	wg.Wait()

	if urlErr != nil {
		return nil, urlErr
	}

	resp := &MusicResponse{
		Code:        200,
		ID:          song.ID,
		Mid:         song.Mid,
		Name:        song.Name,
		ArtistName:  song.ArtistName,
		AlbumName:   song.Album.Name,
		Cover:       song.Cover,
		Level:       level,
		ActualLevel: actual,
		Format:      format,
		Size:        sizeForLevel(song.Sizes, actual),
		URL:         urlStr,
		Time:        time.Now().Format("2006-01-02 15:04:05"),
	}
	if lyric != nil {
		resp.Lyric = lyric.LRC
		resp.QRC = lyric.QRC
		resp.Translation = lyric.Translation
		resp.Romaji = lyric.Romaji
	}
	return resp, nil
}

// resolveURL 按档位请求直链，拿不到时顺序降级。
func (c *Client) resolveURL(ctx context.Context, mid, mediaMid, level string) (string, string, string, error) {
	idx := levelIndex(level)
	if idx < 0 {
		idx = levelIndex(defaultLevel)
	}
	candidates := levelTable[idx:]

	// mediaMid 缺失时用 mid 重复两遍，这是上游的命名约定。
	base := mediaMid
	if base == "" {
		base = mid + mid
	}

	mids := make([]string, 0, len(candidates))
	filenames := make([]string, 0, len(candidates))
	types := make([]int, 0, len(candidates))
	for _, l := range candidates {
		mids = append(mids, mid)
		filenames = append(filenames, l.prefix+base+l.ext)
		types = append(types, 0)
	}

	param := vkeyParam{
		GUID:     newGUID(),
		SongMid:  mids,
		Filename: filenames,
		SongType: types,
		Uin:      c.currentUin(),
	}
	opt := callOptions{
		reqKey:  reqKeyVkey,
		rawComm: newVkeyComm(c.currentUin(), c.currentMusicKey()),
	}

	data, err := c.call(ctx, "music.vkey.GetVkey", "UrlGetVkey", param, opt)
	if err != nil {
		return "", "", "", err
	}

	var resp struct {
		Sip        []string `json:"sip"`
		MidURLInfo []struct {
			Filename string `json:"filename"`
			Purl     string `json:"purl"`
		} `json:"midurlinfo"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", "", "", fmt.Errorf("qm: 解析直链响应失败: %w", err)
	}

	sip := pickSIP(resp.Sip)
	for i, name := range filenames {
		for _, info := range resp.MidURLInfo {
			if info.Filename != name || info.Purl == "" {
				continue
			}
			u := info.Purl
			if !strings.HasPrefix(u, "http") {
				u = sip + strings.TrimPrefix(u, "/")
			}
			return u, candidates[i].name, strings.TrimPrefix(candidates[i].ext, "."), nil
		}
	}
	return "", "", "", ErrNoPlayableURL
}

// pickSIP 取第一个以 http 开头的 CDN 前缀，缺失时回落默认值。
func pickSIP(sip []string) string {
	for _, s := range sip {
		if strings.HasPrefix(s, "http") {
			if !strings.HasSuffix(s, "/") {
				s += "/"
			}
			return s
		}
	}
	return defaultSIP
}

// sizeForLevel 取实际命中档位对应的体积。
func sizeForLevel(s Sizes, level string) int64 {
	switch level {
	case "hi-res":
		return s.HiRes
	case "lossless":
		return s.Flac
	case "hq":
		return s.Mp3_320
	case "sq":
		return s.Mp3_128
	default:
		// lq 无对应体积字段
		return 0
	}
}
