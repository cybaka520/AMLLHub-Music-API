package qm

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

// searchParam 是搜索接口参数，字段名与上游逐字对齐。
type searchParam struct {
	SearchID    string `json:"search_id"`
	RemotePlace string `json:"remoteplace"`
	Query       string `json:"query"`
	PageNum     int    `json:"page_num"`
	NumPerPage  int    `json:"num_per_page"`
	SearchType  int    `json:"search_type"`
	Highlight   int    `json:"highlight"`
	NQCFlag     int    `json:"nqc_flag"`
	PageID      int    `json:"page_id"`
	Grp         int    `json:"grp"`
}

// Search 搜索歌曲，limit 会被收敛到 1..100。
func (c *Client) Search(ctx context.Context, keywords string, limit int) (*SearchResponse, error) {
	keywords = strings.TrimSpace(keywords)
	if keywords == "" {
		return nil, models.NewAPIError(400, "keywords 不能为空")
	}
	if limit < 1 {
		limit = 1
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	param := searchParam{
		SearchID:    newSearchID(),
		RemotePlace: "search.android.keyboard",
		Query:       keywords,
		PageNum:     1,
		NumPerPage:  limit,
	}
	opt := callOptions{
		auth:   false,
		signed: true,
		comm: commPatch{
			tmeAppID: strPtr(searchAppID),
			cv:       strPtr(searchCV),
			v:        strPtr(searchCV),
		},
	}

	data, err := c.call(ctx, "music.search.SearchCgiService", "DoSearchForQQMusicLite", param, opt)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Body struct {
			ItemSong []rawSong `json:"item_song"`
		} `json:"body"`
		Meta struct {
			Sum int `json:"sum"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("qm: 解析搜索响应失败: %w", err)
	}

	// 空结果不是错误，返回空数组。
	out := &SearchResponse{
		Code:  200,
		Total: resp.Meta.Sum,
		Data:  make([]Song, 0, len(resp.Body.ItemSong)),
		Time:  time.Now().Format("2006-01-02 15:04:05"),
	}
	for _, raw := range resp.Body.ItemSong {
		if raw.Mid == "" && raw.ID == 0 {
			continue
		}
		out.Data = append(out.Data, normalizeSong(raw))
	}
	return out, nil
}

// newSearchID 生成搜索会话 ID：group * 2^54 + random * 2^32 + 当日毫秒。
// 结果约 17~18 位十进制，必须用 big.Int，int64 会溢出。
func newSearchID() string {
	const group = 1
	random := rand.Intn(4194304)

	id := new(big.Int).SetUint64(group)
	id.Mul(id, new(big.Int).Lsh(big.NewInt(1), 54))
	id.Add(id, new(big.Int).SetUint64(uint64(random)<<32))
	id.Add(id, big.NewInt(time.Now().UnixMilli()%86400000))
	return id.String()
}
