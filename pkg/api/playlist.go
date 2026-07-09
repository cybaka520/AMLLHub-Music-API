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

// PlaylistAPI 歌单解析API
type PlaylistAPI struct {
	httpClient *client.HTTPClient
}

// NewPlaylistAPI 创建歌单解析API实例
func NewPlaylistAPI(httpClient *client.HTTPClient) *PlaylistAPI {
	return &PlaylistAPI{httpClient: httpClient}
}

// Parse 解析歌单
func (p *PlaylistAPI) Parse(ctx context.Context, input string) (*models.PlaylistResponse, error) {
	// 参数验证
	if input == "" {
		return nil, models.NewAPIError(400, "缺少歌单ID参数")
	}

	// 提取歌单ID
	playlistID, err := utils.ExtractPlaylistID(input)
	if err != nil || playlistID == "" {
		return nil, models.NewAPIError(400, "无法从输入中提取有效的歌单ID")
	}

	// 获取歌单数据
	playlistData, err := p.fetchPlaylistData(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("获取歌单数据失败: %w", err)
	}

	if playlistData.Code != 200 {
		return nil, models.NewAPIError(playlistData.Code, "歌单解析失败")
	}

	// 提取所有 trackIds
	var trackIDs []int64
	if playlistData.Playlist.TrackIds != nil {
		for _, track := range playlistData.Playlist.TrackIds {
			trackIDs = append(trackIDs, track.ID)
		}
	}

	// 如果有 trackIds，获取所有歌曲的详细信息
	var fetchedTracks []models.Track
	if len(trackIDs) > 0 {
		fetchedTracks, err = p.fetchSongsData(ctx, trackIDs)
		if err == nil && len(fetchedTracks) > 0 {
			playlistData.Playlist.TrackCount = len(fetchedTracks)
		}
	}

	// 简化响应数据
	var result *models.PlaylistResponse
	if len(fetchedTracks) > 0 {
		result = p.simplifyPlaylistDataWithTracks(playlistData, fetchedTracks)
	} else {
		result = p.simplifyPlaylistData(playlistData)
	}
	result.Time = time.Now().Format("2006-01-02 15:04:05")

	return result, nil
}

// playlistRawResponse 歌单API原始响应
type playlistRawResponse struct {
	Code     int `json:"code"`
	Playlist struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		CoverImgURL string `json:"coverImgUrl"`
		UserID      int64  `json:"userId"`
		CreateTime  int64  `json:"createTime"`
		TrackCount  int    `json:"trackCount"`
		Creator     struct {
			UserID    int64  `json:"userId"`
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatarUrl"`
		} `json:"creator"`
		TrackIds []struct {
			ID int64 `json:"id"`
		} `json:"trackIds"`
		Tracks []rawTrack `json:"tracks"`
	} `json:"playlist"`
}

// rawTrack 原始歌曲数据
type rawTrack struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
	Ar   []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"ar"`
	Al struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		PicURL string `json:"picUrl"`
	} `json:"al"`
	Dt  int64 `json:"dt"`
	Fee int   `json:"fee"`
}

// fetchPlaylistData 获取歌单数据
func (p *PlaylistAPI) fetchPlaylistData(ctx context.Context, playlistID string) (*playlistRawResponse, error) {
	// 使用 url.Values 对 playlistID 做编码，避免特殊字符破坏 URL
	q := url.Values{}
	q.Set("id", playlistID)
	apiURL := "https://interface3.music.163.com/api/v6/playlist/detail?" + q.Encode()

	body, err := p.httpClient.Get(ctx, apiURL)
	if err != nil {
		return nil, err
	}

	var result playlistRawResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析歌单响应失败: %w", err)
	}

	return &result, nil
}

// songDetailItem 用于批量构造 c 参数的 JSON（ID 为数值）
type songDetailItem struct {
	ID int64 `json:"id"`
	V  int   `json:"v"`
}

// fetchSongsData 获取多首歌曲的详细信息
// 返回成功获取的歌曲列表；若全部 chunk 失败则返回 error。
func (p *PlaylistAPI) fetchSongsData(ctx context.Context, trackIDs []int64) ([]models.Track, error) {
	if len(trackIDs) == 0 {
		return nil, nil
	}

	var tracks []models.Track
	var chunkErrors []error

	// 网易云API限制每次最多获取1000首歌曲详情
	chunkSize := 1000
	for i := 0; i < len(trackIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(trackIDs) {
			end = len(trackIDs)
		}
		chunk := trackIDs[i:end]

		// 用 json.Marshal 安全构造 c 参数，避免注入
		criteria := make([]songDetailItem, 0, len(chunk))
		for _, id := range chunk {
			criteria = append(criteria, songDetailItem{ID: id, V: 0})
		}
		cJSON, err := json.Marshal(criteria)
		if err != nil {
			chunkErrors = append(chunkErrors, fmt.Errorf("序列化 chunk 失败: %w", err))
			continue
		}

		apiURL := "https://music.163.com/api/v3/song/detail"
		data := url.Values{}
		data.Set("c", string(cJSON))

		body, err := p.httpClient.PostForm(ctx, apiURL, data, nil)
		if err != nil {
			chunkErrors = append(chunkErrors, fmt.Errorf("chunk[%d-%d] 请求失败: %w", i, end-1, err))
			continue
		}

		var result struct {
			Code  int        `json:"code"`
			Songs []rawTrack `json:"songs"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			chunkErrors = append(chunkErrors, fmt.Errorf("chunk[%d-%d] 解析失败: %w", i, end-1, err))
			continue
		}

		if result.Code == 200 {
			for _, song := range result.Songs {
				track := models.Track{
					Name: song.Name,
					ID:   song.ID,
					Al:   models.Album{ID: song.Al.ID, Name: song.Al.Name, PicURL: song.Al.PicURL},
					Dt:   song.Dt,
					Fee:  song.Fee,
				}
				for _, ar := range song.Ar {
					track.Ar = append(track.Ar, models.Artist{ID: ar.ID, Name: ar.Name})
				}
				tracks = append(tracks, track)
			}
		}
	}

	// 全部 chunk 失败时返回错误
	if len(tracks) == 0 && len(chunkErrors) > 0 {
		return nil, fmt.Errorf("所有歌曲详情请求失败: %v", chunkErrors)
	}

	return tracks, nil
}

// simplifyPlaylistData 简化歌单数据
func (p *PlaylistAPI) simplifyPlaylistData(data *playlistRawResponse) *models.PlaylistResponse {
	result := &models.PlaylistResponse{
		Code: data.Code,
	}

	result.Playlist = models.PlaylistDetail{
		ID:          data.Playlist.ID,
		Name:        data.Playlist.Name,
		CoverImgURL: data.Playlist.CoverImgURL,
		UserID:      data.Playlist.UserID,
		CreateTime:  data.Playlist.CreateTime,
		TrackCount:  data.Playlist.TrackCount,
		Creator: models.Creator{
			UserID:    data.Playlist.Creator.UserID,
			Nickname:  data.Playlist.Creator.Nickname,
			AvatarURL: data.Playlist.Creator.AvatarURL,
		},
	}

	for _, track := range data.Playlist.Tracks {
		t := models.Track{
			Name: track.Name,
			ID:   track.ID,
			Al:   models.Album{ID: track.Al.ID, Name: track.Al.Name, PicURL: track.Al.PicURL},
			Dt:   track.Dt,
			Fee:  track.Fee,
		}
		for _, ar := range track.Ar {
			t.Ar = append(t.Ar, models.Artist{ID: ar.ID, Name: ar.Name})
		}
		result.Playlist.Tracks = append(result.Playlist.Tracks, t)
	}

	return result
}

// simplifyPlaylistDataWithTracks 简化歌单数据（使用已获取的歌曲列表）
func (p *PlaylistAPI) simplifyPlaylistDataWithTracks(data *playlistRawResponse, tracks []models.Track) *models.PlaylistResponse {
	result := &models.PlaylistResponse{
		Code: data.Code,
	}

	result.Playlist = models.PlaylistDetail{
		ID:          data.Playlist.ID,
		Name:        data.Playlist.Name,
		CoverImgURL: data.Playlist.CoverImgURL,
		UserID:      data.Playlist.UserID,
		CreateTime:  data.Playlist.CreateTime,
		TrackCount:  len(tracks),
		Creator: models.Creator{
			UserID:    data.Playlist.Creator.UserID,
			Nickname:  data.Playlist.Creator.Nickname,
			AvatarURL: data.Playlist.Creator.AvatarURL,
		},
		Tracks: tracks,
	}

	return result
}
