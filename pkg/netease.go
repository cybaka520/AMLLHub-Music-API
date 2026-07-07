// Package netease 是网易云音乐API解析库的主入口
//
// 快速开始:
//
//	client, _ := netease.NewClient("MUSIC_U=xxxxx")
//	result, _ := client.Search("周杰伦", 10)
//	music, _ := client.ParseMusic("123456", "lossless")
package netease

import (
	"github.com/cybaka520/AMLLHub-Music-API/pkg/api"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/client"
	"github.com/cybaka520/AMLLHub-Music-API/pkg/models"
)

// Client 网易云音乐API客户端
type Client struct {
	token       string
	httpClient  *client.HTTPClient
	searchAPI   *api.SearchAPI
	musicAPI    *api.MusicAPI
	playlistAPI *api.PlaylistAPI
}

// NewClient 创建客户端（仅需Token）
func NewClient(token string) *Client {
	httpClient := client.NewHTTPClient(token)
	return &Client{
		token:       token,
		httpClient:  httpClient,
		searchAPI:   api.NewSearchAPI(httpClient),
		musicAPI:    api.NewMusicAPI(httpClient),
		playlistAPI: api.NewPlaylistAPI(httpClient),
	}
}

// Search 搜索音乐
// keywords: 搜索关键词
// limit: 返回结果数量（1-100）
func (c *Client) Search(keywords string, limit int) (*models.SearchResponse, error) {
	return c.searchAPI.Search(keywords, limit)
}

// ParseMusic 解析单曲
// songID: 歌曲ID或URL
// level: 音质等级（standard/exhigh/lossless/hires/jyeffect/jymaster/sky/dolby）
func (c *Client) ParseMusic(songID string, level string) (*models.MusicResponse, error) {
	return c.musicAPI.Parse(songID, level)
}

// ParsePlaylist 解析歌单
// playlistID: 歌单ID或URL
func (c *Client) ParsePlaylist(playlistID string) (*models.PlaylistResponse, error) {
	return c.playlistAPI.Parse(playlistID)
}
