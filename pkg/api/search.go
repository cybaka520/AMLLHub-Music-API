// Package api 实现网易云音乐的核心解析功能
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/client"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/utils"
)

// SearchAPI 音乐搜索API
type SearchAPI struct {
	httpClient *client.HTTPClient
}

// NewSearchAPI 创建搜索API实例
func NewSearchAPI(httpClient *client.HTTPClient) *SearchAPI {
	return &SearchAPI{httpClient: httpClient}
}

// Search 搜索音乐
func (s *SearchAPI) Search(ctx context.Context, keywords string, limit int) (*models.SearchResponse, error) {
	// 参数验证
	if keywords == "" {
		return nil, models.NewAPIError(400, "请提供歌曲名称")
	}
	limit = clamp(limit, 1, 100)

	// 构建请求URL
	reqURL := fmt.Sprintf(
		"https://interface3.music.163.com/api/search/song/list/page?keyword=%s&limit=%d&needCorrect=1&scene=normal",
		url.QueryEscape(keywords), limit,
	)

	// 发送请求
	body, err := s.httpClient.PostSearch(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("搜索请求失败: %w", err)
	}

	// 解析原始响应
	var raw searchRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析搜索响应失败: %w", err)
	}

	// 构建格式化响应
	return s.formatResponse(&raw), nil
}

// searchRawResponse 搜索API原始响应结构
type searchRawResponse struct {
	Code int `json:"code"`
	Data struct {
		Resources []struct {
			ResourceType string `json:"resourceType"`
			BaseInfo     struct {
				SimpleSongData struct {
					ID          int64   `json:"id"`
					Name        string  `json:"name"`
					Fee         int     `json:"fee"`
					Pop         float64 `json:"pop"`
					Dt          int64   `json:"dt"`
					PublishTime int64   `json:"publishTime"`
					Al          struct {
						ID     int64  `json:"id"`
						Name   string `json:"name"`
						PicURL string `json:"picUrl"`
					} `json:"al"`
					Ar []struct {
						ID   int64  `json:"id"`
						Name string `json:"name"`
					} `json:"ar"`
					Alia []string `json:"alia"`
				} `json:"simpleSongData"`
			} `json:"baseInfo"`
			ExtInfo struct {
				StarCount int `json:"starCount"`
			} `json:"extInfo"`
		} `json:"resources"`
	} `json:"data"`
}

func (s *SearchAPI) formatResponse(raw *searchRawResponse) *models.SearchResponse {
	result := &models.SearchResponse{
		Code: raw.Code,
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}

	if raw.Code == 200 {
		for _, resource := range raw.Data.Resources {
			if resource.ResourceType != "song" {
				continue
			}

			songData := resource.BaseInfo.SimpleSongData

			// 构建艺术家信息
			var artists []models.Artist
			for _, ar := range songData.Ar {
				artists = append(artists, models.Artist{ID: ar.ID, Name: ar.Name})
			}

			// 构建专辑信息
			album := models.Album{
				ID:     songData.Al.ID,
				Name:   songData.Al.Name,
				PicURL: songData.Al.PicURL,
			}
			if songData.PublishTime > 0 {
				album.PublishTime = time.Unix(songData.PublishTime/1000, 0).Format("2006-01-02")
			}

			// 构建别名
			alias := ""
			if len(songData.Alia) > 0 {
				alias = songData.Alia[0]
			}

			result.Data = append(result.Data, models.Song{
				ID:        songData.ID,
				Name:      songData.Name,
				Fee:       songData.Fee,
				Pop:       songData.Pop,
				StarCount: resource.ExtInfo.StarCount,
				PicURL:    songData.Al.PicURL,
				Artists:   artists,
				Album:     album,
				Duration:  utils.FormatDuration(songData.Dt),
				Alias:     alias,
			})
		}
	}

	return result
}

// clamp 限制数值范围
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
