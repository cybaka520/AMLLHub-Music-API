package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cybaka520/AMLLHub-Music-API/pkg/client"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/crypto"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/utils"
)

// MusicAPI 单曲解析API
type MusicAPI struct {
	httpClient *client.HTTPClient
}

// NewMusicAPI 创建单曲解析API实例
func NewMusicAPI(httpClient *client.HTTPClient) *MusicAPI {
	return &MusicAPI{httpClient: httpClient}
}

// Parse 解析单曲
func (m *MusicAPI) Parse(songInput string, level string) (*models.MusicResponse, error) {
	// 参数验证
	if songInput == "" {
		return &models.MusicResponse{Code: 400, Error: "必须提供 ids 或 url 参数"}, nil
	}
	if level == "" {
		return &models.MusicResponse{Code: 400, Error: "level参数为空"}, nil
	}

	// 提取歌曲ID
	songID, err := utils.ExtractSongID(songInput)
	if err != nil || songID == "" {
		return &models.MusicResponse{Code: 400, Error: "无法从输入解析有效的歌曲ID"}, nil
	}

	// 获取歌曲URL
	urlData, err := m.getSongURL(songID, level)
	if err != nil {
		return &models.MusicResponse{Code: 500, Error: err.Error()}, nil
	}

	if urlData == nil || len(urlData.Data) == 0 || urlData.Data[0].URL == "" {
		return &models.MusicResponse{Code: 400, Error: "获取歌曲链接信息失败"}, nil
	}

	urlInfo := urlData.Data[0]
	songIDFromURL := fmt.Sprintf("%d", urlInfo.ID)

	// 获取歌曲详情
	songDetail, err := m.getSongDetail(songIDFromURL)
	if err != nil || songDetail == nil || len(songDetail.Songs) == 0 {
		return &models.MusicResponse{Code: 400, Error: "获取歌曲名称等详细信息失败"}, nil
	}

	songInfo := songDetail.Songs[0]

	// 获取歌词
	lyricData, _ := m.getLyric(songIDFromURL)

	// 构建响应
	resp := &models.MusicResponse{
		Code:       200,
		Name:       songInfo.Name,
		Pic:        songInfo.Al.PicURL,
		AlbumName:  songInfo.Al.Name,
		Level:      utils.FormatLevel(urlInfo.Level),
		Size:       utils.FormatFileSize(urlInfo.Size),
		URL:        replaceHTTPtoHTTPS(urlInfo.URL),
	}

	// 构建歌手名
	var artistNames []string
	for _, ar := range songInfo.Ar {
		artistNames = append(artistNames, ar.Name)
	}
	resp.ArtistName = joinStrings(artistNames, "/")

	// 歌词
	if lyricData != nil {
		resp.Lyric = lyricData.Lrc.Lyric
		resp.TLyric = lyricData.Tlyric.Lyric
	}

	return resp, nil
}

// urlV1Response 歌曲URL API响应
type urlV1Response struct {
	Code int `json:"code"`
	Data []struct {
		ID    int64  `json:"id"`
		URL   string `json:"url"`
		Level string `json:"level"`
		Size  int64  `json:"size"`
	} `json:"data"`
}

// getSongURL 获取歌曲播放链接（eapi加密接口）
func (m *MusicAPI) getSongURL(songID string, level string) (*urlV1Response, error) {
	apiURL := "https://interface3.music.163.com/eapi/song/enhance/player/url/v1"
	apiPath := "/api/song/enhance/player/url/v1"

	config := map[string]interface{}{
		"os":        "android",
		"appver":    "9.3.90",
		"osver":     "",
		"deviceId":  "pyncm!",
		"requestId": client.RandomRequestID(),
	}

	configJSON, _ := json.Marshal(config)

	payload := map[string]interface{}{
		"ids":         []string{songID},
		"level":       level,
		"encodeType":  "flac",
		"header":      string(configJSON),
	}

	if level == "sky" {
		payload["immerseType"] = "c51"
	}

	// 加密请求参数
	encryptedParams, err := crypto.EncryptAPIRequest(apiPath, payload)
	if err != nil {
		return nil, err
	}

	// 发送请求
	body, err := m.httpClient.PostEAPI(apiURL, encryptedParams, nil)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result urlV1Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// songDetailResponse 歌曲详情API响应
type songDetailResponse struct {
	Code  int `json:"code"`
	Songs []struct {
		Name string `json:"name"`
		Al   struct {
			PicURL string `json:"picUrl"`
			Name   string `json:"name"`
		} `json:"al"`
		Ar []struct {
			Name string `json:"name"`
		} `json:"ar"`
	} `json:"songs"`
}

// getSongDetail 获取歌曲详细信息（带重试）
func (m *MusicAPI) getSongDetail(songID string) (*songDetailResponse, error) {
	apiURL := "https://interface3.music.163.com/api/v3/song/detail"
	data := url.Values{}
	data.Set("c", fmt.Sprintf(`[{"id":"%s","v":0}]`, songID))

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		body, err := m.httpClient.PostForm(apiURL, data, nil)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		var result songDetailResponse
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		if result.Code == 200 && len(result.Songs) > 0 {
			return &result, nil
		}

		lastErr = fmt.Errorf("歌曲详情为空")
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	return nil, lastErr
}

// lyricResponse 歌词API响应
type lyricResponse struct {
	Code   int `json:"code"`
	Lrc    struct {
		Lyric string `json:"lyric"`
	} `json:"lrc"`
	Tlyric struct {
		Lyric string `json:"lyric"`
	} `json:"tlyric"`
}

// getLyric 获取歌词
func (m *MusicAPI) getLyric(songID string) (*lyricResponse, error) {
	apiURL := "https://interface3.music.163.com/api/song/lyric"
	data := url.Values{}
	data.Set("id", songID)
	data.Set("cp", "false")
	data.Set("tv", "0")
	data.Set("lv", "0")
	data.Set("rv", "0")
	data.Set("kv", "0")
	data.Set("yv", "0")
	data.Set("ytv", "0")
	data.Set("yrv", "0")

	body, err := m.httpClient.PostForm(apiURL, data, nil)
	if err != nil {
		return nil, err
	}

	var result lyricResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// joinStrings 用分隔符连接字符串切片
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
