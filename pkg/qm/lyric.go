package qm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

// lyricParam 是歌词接口参数。所有文本字段都必须 Base64(UTF-8)。
type lyricParam struct {
	SongID     int64  `json:"songID"`
	SongName   string `json:"songName"`
	SingerName string `json:"singerName"`
	AlbumName  string `json:"albumName"`
	Interval   int    `json:"interval"` // 秒
	Crypt      int    `json:"crypt"`
	CT         int    `json:"ct"`
	CV         int    `json:"cv"`
	LRCT       int    `json:"lrc_t"`
	QRCT       int    `json:"qrc_t"`
	TransT     int    `json:"trans_t"`
	RomaT      int    `json:"roma_t"`
	QRC        int    `json:"qrc"`
	Trans      int    `json:"trans"`
	Roma       int    `json:"roma"`
	Type       int    `json:"type"`
}

// lyricRaw 是歌词接口响应，字段为 hex 密文。
type lyricRaw struct {
	Lyric flexStr `json:"lyric"`
	QRC   flexStr `json:"qrc"`
	QRCT  flexInt `json:"qrc_t"` // 0 表示 lyric 解出来是 LRC，非 0 表示是 QRC
	Trans flexStr `json:"trans"`
	Roma  flexStr `json:"roma"`
}

// flexStr 兼容上游把「无内容」的歌词字段写成数字 0 的情况（匿名请求常见）。
// 若按 string 强解，整个响应会因一个字段类型不符而解析失败，LRC 也就一起丢了。
type flexStr string

// UnmarshalJSON 接受字符串、数字与 null，非字符串一律视为空。
func (f *flexStr) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*f = ""
		return nil
	}
	if s[0] != '"' {
		*f = ""
		return nil
	}
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*f = flexStr(v)
	return nil
}

// GetLyric 获取歌词：LRC + QRC + 翻译 + 罗马音。
//
// 纯音乐返回零值 Lyric 与 nil，不算错误；任一可选字段失败只置空。
func (c *Client) GetLyric(ctx context.Context, songID int64, name, artist, album string, durationSec int) (*Lyric, error) {
	if songID <= 0 {
		return nil, models.NewAPIError(400, "songID 必须大于 0")
	}

	param := lyricParam{
		SongID:     songID,
		SongName:   b64(name),
		SingerName: b64(artist),
		AlbumName:  b64(album),
		Interval:   durationSec,
		Crypt:      1,
		CT:         19,
		CV:         2111,
		QRC:        1,
		Trans:      1,
		Roma:       1,
	}

	raw, err := c.fetchLyric(ctx, param)
	if err != nil {
		return nil, err
	}

	out := &Lyric{
		Translation: decodeLyricField(string(raw.Trans)),
		Romaji:      decodeLyricField(string(raw.Roma)),
	}

	main := decodeLyricField(string(raw.Lyric))
	if raw.QRCT != 0 {
		out.QRC = firstNonEmpty(decodeLyricField(string(raw.QRC)), main)
	} else {
		out.LRC = main
		out.QRC = decodeLyricField(string(raw.QRC))
	}

	// QRC 命中时再要一次 LRC 兜底，失败不影响。
	if out.QRC != "" && out.LRC == "" {
		alt := param
		alt.QRC = 0
		alt.QRCT = 0
		if plain, err := c.fetchLyric(ctx, alt); err == nil {
			out.LRC = decodeLyricField(string(plain.Lyric))
		}
	}
	return out, nil
}

// fetchLyric 发起一次歌词请求。
func (c *Client) fetchLyric(ctx context.Context, param lyricParam) (*lyricRaw, error) {
	data, err := c.call(ctx, "music.musichallSong.PlayLyricInfo", "GetPlayLyricInfo", param, defaultCall())
	if err != nil {
		return nil, err
	}

	var raw lyricRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("qm: 解析歌词响应失败: %w", err)
	}
	return &raw, nil
}

// decodeLyricField 解密单段歌词。
// 解密失败时若字段本身像歌词（含 [ 或 <），按明文返回，兼容上游直出的情况。
func decodeLyricField(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	if out, err := decryptQRC(s); err == nil {
		return out
	}
	if strings.ContainsAny(s, "[<") {
		return s
	}
	return ""
}

// b64 编码歌词请求里的文本字段。
func b64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
