# 网易云音乐API Go语言重构文档（极简版）

**项目地址**: https://github.com/cybaka520/AMLLHub-Music-API

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术栈](#2-技术栈)
3. [项目结构](#3-项目结构)
4. [核心API设计](#4-核心api设计)
5. [重构步骤](#5-重构步骤)
6. [使用示例](#6-使用示例)
7. [性能优化](#7-性能优化)
8. [测试方案](#8-测试方案)

---

## 1. 项目概述

### 定位
**纯Go解析库**，极致简洁，专注核心功能。

### 核心功能（仅3个）
- ✅ 音乐搜索
- ✅ 单曲解析
- ✅ 歌单解析

### 明确排除的功能
- ❌ 缓存功能（由调用方实现）
- ❌ HTTP API服务（纯库设计）
- ❌ 管理后台
- ❌ 公告系统
- ❌ 统计数据
- ❌ Token池管理（单一Token）

### 优势
- 🚀 性能提升5-10倍（相比PHP）
- 📦 编译后 < 5MB
- ⚡ 零外部依赖（无Redis、无数据库）
- 🎯 极简API（仅需3个方法）

---

## 2. 技术栈

### 核心技术（仅标准库）
- **Go 1.21+**
- `net/http` - HTTP客户端
- `crypto/aes` - AES加密
- `crypto/md5` - MD5哈希
- `encoding/json` - JSON处理

### 测试依赖
- `github.com/stretchr/testify` - 单元测试

### 明确排除
- ❌ Redis
- ❌ 数据库
- ❌ 配置文件库
- ❌ 第三方HTTP库
- ❌ 日志库

---

## 3. 项目结构

```
netease-music-api/
├── pkg/
│   ├── api/
│   │   ├── music.go       # 单曲解析
│   │   ├── search.go      # 音乐搜索
│   │   └ playlist.go    # 歌单解析
│   │
│   ├── client/
│   │   ├── http.go        # HTTP客户端
│   │   └ crypto.go      # 加密处理
│   │
│   ├── crypto/
│   │   ├── aes.go         # AES加密
│   │   └ md5.go           # MD5哈希
│   │
│   ├── models/
│   │   ├── song.go        # 歌曲模型
│   │   ├── playlist.go    # 歌单模型
│   │   ├── search.go      # 搜索结果
│   │   └ response.go    # API响应
│   │
│   ├── utils/
│   │   ├── fake.go        # IP/UA伪造
│   │   ├── format.go      # 格式化
│   │   └ parser.go      # ID解析
│   │
│   └── netease.go         # 主入口
│
├── examples/
│   ├── basic.go           # 基础示例
│   ├── batch.go           # 批量解析
│   └ playlist.go        # 歌单示例
│
├── tests/
│   ├── music_test.go
│   ├── search_test.go
│   ├── playlist_test.go
│   └ bench_test.go
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└ LICENSE
```

**总文件数：约20个**（不含测试）

---

## 4. 核心API设计

### 主入口

```go
package netease

// Client - 客户端（极简设计）
type Client struct {
    token     string
    httpClient *HTTPClient
}

// NewClient - 创建客户端（仅需Token）
func NewClient(token string) (*Client, error) {
    return &Client{
        token:     token,
        httpClient: NewHTTPClient(token),
    }, nil
}
```

### 核心方法（仅3个）

```go
// Search - 搜索音乐
func (c *Client) Search(keywords string, limit int) (*SearchResponse, error)

// ParseMusic - 解析单曲
func (c *Client) ParseMusic(songID string, level string) (*MusicResponse, error)

// ParsePlaylist - 解析歌单
func (c *Client) ParsePlaylist(playlistID string) (*PlaylistResponse, error)
```

### 数据模型

```go
// SearchResponse - 搜索响应
type SearchResponse struct {
    Code int        `json:"code"`
    Data []Song     `json:"data"`
    Error string    `json:"error,omitempty"`
}

// MusicResponse - 音乐解析响应
type MusicResponse struct {
    Code       int    `json:"code"`
    Name       string `json:"name"`
    ArtistName string `json:"ar_name"`
    AlbumName  string `json:"al_name"`
    Level      string `json:"level"`
    Size       string `json:"size"`
    URL        string `json:"url"`
    Pic        string `json:"pic"`
    Lyric      string `json:"lyric"`
    Error      string `json:"error,omitempty"`
}

// PlaylistResponse - 歌单响应
type PlaylistResponse struct {
    Code     int            `json:"code"`
    Playlist PlaylistDetail `json:"playlist"`
    Error    string         `json:"error,omitempty"`
}

// Song - 歌曲模型
type Song struct {
    ID      int64    `json:"id"`
    Name    string   `json:"name"`
    Artists []Artist `json:"artists"`
    Album   Album    `json:"album"`
}

// Artist - 歌手
type Artist struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

// Album - 专辑
type Album struct {
    ID     int64  `json:"id"`
    Name   string `json:"name"`
    PicURL string `json:"picUrl"`
}
```

### 工具函数

```go
package utils

// ExtractSongID - 提取歌曲ID
func ExtractSongID(input string) (string, error)

// ExtractPlaylistID - 提取歌单ID
func ExtractPlaylistID(input string) (string, error)

// FormatFileSize - 格式化文件大小
func FormatFileSize(bytes int64) string

// FormatDuration - 格式化时长
func FormatDuration(milliseconds int) string

// FormatLevel - 格式化音质
func FormatLevel(level string) string

// RandomIP - 随机IP
func RandomIP() string

// RandomUA - 随机User-Agent
func RandomUA() string
```

---

## 5. 重构步骤（5天计划）

### 第1天：HTTP客户端 + 加密模块

**任务**：
1. 实现 `pkg/client/http.go` - HTTP请求封装
2. 实现 `pkg/crypto/aes.go` - AES加密
3. 实现 `pkg/crypto/md5.go` - MD5哈希

**关键代码**：

```go
// pkg/client/http.go
type HTTPClient struct {
    client    *http.Client
    token     string
    userAgent string
    fakeIP    string
}

func NewHTTPClient(token string) *HTTPClient {
    transport := &http.Transport{
        MaxIdleConns:        50,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     30 * time.Second,
    }
    
    return &HTTPClient{
        client: &http.Client{
            Transport: transport,
            Timeout:   10 * time.Second,
        },
        token:     token,
        userAgent: utils.RandomUA(),
        fakeIP:    utils.RandomIP(),
    }
}

func (h *HTTPClient) Post(url string, data map[string]string) ([]byte, error) {
    // 构建请求
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return nil, err
    }
    
    // 设置请求头
    req.Header.Set("User-Agent", h.userAgent)
    req.Header.Set("Cookie", h.token)
    req.Header.Set("X-Forwarded-For", h.fakeIP)
    
    // 发送请求
    resp, err := h.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return ioutil.ReadAll(resp.Body)
}
```

```go
// pkg/crypto/aes.go
const AESKey = "e82ckenh8dichen8"

func AESEncrypt(data string) (string, error) {
    block, _ := aes.NewCipher([]byte(AESKey))
    
    // PKCS7填充
    padded := pkcs7Pad([]byte(data), 16)
    
    // ECB加密
    encrypted := make([]byte, len(padded))
    for i := 0; i < len(padded); i += 16 {
        block.Encrypt(encrypted[i:i+16], padded[i:i+16])
    }
    
    return hex.EncodeToString(encrypted), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
    padding := blockSize - len(data)%blockSize
    return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}
```

### 第2天：数据模型 + 工具函数

**任务**：
1. 定义 `pkg/models/` 下所有数据结构
2. 实现 `pkg/utils/fake.go` - IP/UA伪造
3. 实现 `pkg/utils/format.go` - 格式化工具
4. 实现 `pkg/utils/parser.go` - ID解析

**关键代码**：

```go
// pkg/utils/fake.go
func RandomIP() string {
    // 中国IP范围
    ranges := []string{
        "116.0.0.0", "119.255.255.255",
        "120.0.0.0", "123.255.255.255",
    }
    
    rangeIdx := rand.Intn(len(ranges))
    start := parseIP(ranges[rangeIdx*2])
    end := parseIP(ranges[rangeIdx*2+1])
    
    ip := make(net.IP, 4)
    for i := 0; i < 4; i++ {
        ip[i] = byte(start[i] + rand.Intn(int(end[i]-start[i])))
    }
    
    return ip.String()
}

func RandomUA() string {
    uas := []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
    }
    return uas[rand.Intn(len(uas))]
}
```

```go
// pkg/utils/parser.go
func ExtractSongID(input string) (string, error) {
    // 纯数字
    if regexp.MustCompile(`^\d+$`).MatchString(input) {
        return input, nil
    }
    
    // 处理URL
    if strings.Contains(input, "music.163.com") {
        u, err := url.Parse(input)
        if err == nil && u.Query().Get("id") != "" {
            return u.Query().Get("id"), nil
        }
        
        // 字符串查找id=
        if idx := strings.Index(input, "id="); idx != -1 {
            return input[idx+3:], nil
        }
    }
    
    return "", fmt.Errorf("无法解析歌曲ID")
}
```

### 第3天：搜索API + 单曲解析API

**任务**：
1. 实现 `pkg/api/search.go` - 搜索功能
2. 实现 `pkg/api/music.go` - 单曲解析

**关键代码**：

```go
// pkg/api/search.go
func (c *Client) Search(keywords string, limit int) (*SearchResponse, error) {
    // 参数验证
    if keywords == "" {
        return nil, errors.New("关键词不能为空")
    }
    limit = clamp(limit, 1, 100)
    
    // 构建请求
    url := "https://interface3.music.163.com/api/search/song/list/page"
    params := map[string]string{
        "keyword": keywords,
        "limit":   fmt.Sprintf("%d", limit),
    }
    
    // 发送请求
    body, err := c.httpClient.Post(url, params)
    if err != nil {
        return nil, err
    }
    
    // 解析响应
    var result SearchResponse
    json.Unmarshal(body, &result)
    
    return &result, nil
}
```

```go
// pkg/api/music.go
func (c *Client) ParseMusic(songID string, level string) (*MusicResponse, error) {
    // 参数验证
    if songID == "" {
        return nil, errors.New("歌曲ID不能为空")
    }
    
    // 提取ID
    songID, err := utils.ExtractSongID(songID)
    if err != nil {
        return nil, err
    }
    
    // 获取歌曲URL
    urlData, err := c.getSongURL(songID, level)
    if err != nil {
        return nil, err
    }
    
    // 获取歌曲详情
    songDetail, err := c.getSongDetail(songID)
    if err != nil {
        return nil, err
    }
    
    // 获取歌词
    lyric, _ := c.getLyric(songID)
    
    // 构建响应
    return &MusicResponse{
        Code:       200,
        Name:       songDetail.Name,
        ArtistName: songDetail.Artists[0].Name,
        AlbumName:  songDetail.Album.Name,
        Level:      utils.FormatLevel(urlData.Level),
        Size:       utils.FormatFileSize(urlData.Size),
        URL:        urlData.URL,
        Pic:        songDetail.Album.PicURL,
        Lyric:      lyric,
    }, nil
}

func (c *Client) getSongURL(songID, level string) (*URLData, error) {
    // 构建加密请求
    params := map[string]interface{}{
        "ids":       []string{songID},
        "level":     level,
        "encodeType": "flac",
    }
    
    encryptedParams, _ := crypto.EncryptAPIRequest("/api/song/enhance/player/url/v1", params)
    
    // 发送请求到网易云音乐API
    url := "https://interface3.music.163.com/eapi/song/enhance/player/url/v1"
    body, _ := c.httpClient.Post(url, map[string]string{"params": encryptedParams})
    
    // 解析响应
    var result URLResponse
    json.Unmarshal(body, &result)
    
    return result.Data[0], nil
}
```

### 第4天：歌单解析API + 主入口

**任务**：
1. 实现 `pkg/api/playlist.go` - 歌单解析
2. 实现 `pkg/netease.go` - 主入口整合

**关键代码**：

```go
// pkg/api/playlist.go
func (c *Client) ParsePlaylist(playlistID string) (*PlaylistResponse, error) {
    // 提取ID
    playlistID, err := utils.ExtractPlaylistID(playlistID)
    if err != nil {
        return nil, err
    }
    
    // 获取歌单详情
    url := fmt.Sprintf("https://interface3.music.163.com/api/v6/playlist/detail?id=%s", playlistID)
    body, _ := c.httpClient.Get(url)
    
    // 解析响应
    var result PlaylistResponse
    json.Unmarshal(body, &result)
    
    return &result, nil
}
```

```go
// pkg/netease.go
package netease

import (
    "github.com/cybaka520/AMLLHub-Music-API/pkg/client"
)

type Client struct {
    token     string
    httpClient *client.HTTPClient
}

// NewClient - 创建客户端
func NewClient(token string) (*Client, error) {
    return &Client{
        token:     token,
        httpClient: client.NewHTTPClient(token),
    }, nil
}

// 导入子包
func (c *Client) Search(keywords string, limit int) (*SearchResponse, error) {
    return searchAPI.Search(c.httpClient, keywords, limit)
}

func (c *Client) ParseMusic(songID string, level string) (*MusicResponse, error) {
    return musicAPI.Parse(c.httpClient, songID, level)
}

func (c *Client) ParsePlaylist(playlistID string) (*PlaylistResponse, error) {
    return playlistAPI.Parse(c.httpClient, playlistID)
}
```

### 第5天：测试 + 示例 + 文档

**任务**：
1. 编写单元测试（3个）
2. 编写使用示例（3个）
3. 编写README.md

---

## 6. 使用示例

### 示例1：基础使用

```go
package main

import (
    "fmt"
    "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
    // 创建客户端（仅需Token）
    token := "MUSIC_U=xxxxx; NMTID=xxxxx"
    client, _ := netease.NewClient(token)
    
    // 搜索音乐
    result, _ := client.Search("周杰伦", 10)
    for _, song := range result.Data {
        fmt.Printf("%s - %s\n", song.Name, song.Artists[0].Name)
    }
    
    // 解析单曲
    music, _ := client.ParseMusic("123456", "lossless")
    fmt.Printf("歌曲: %s\n", music.Name)
    fmt.Printf("URL: %s\n", music.URL)
    fmt.Printf("音质: %s\n", music.Level)
}
```

### 示例2：批量解析（并发）

```go
package main

import (
    "fmt"
    "sync"
    "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
    client, _ := netease.NewClient("your_token")
    
    // 搜索歌曲列表
    result, _ := client.Search("测试", 50)
    
    // 并发解析（Go优势）
    var wg sync.WaitGroup
    for i, song := range result.Data {
        wg.Add(1)
        go func(idx int, songID string) {
            defer wg.Done()
            
            music, err := client.ParseMusic(songID, "standard")
            if err == nil {
                fmt.Printf("[%d] %s -> %s\n", idx, songID, music.URL)
            }
        }(i, fmt.Sprintf("%d", song.ID))
    }
    wg.Wait()
}
```

### 示例3：歌单解析

```go
package main

import (
    "fmt"
    "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
    client, _ := netease.NewClient("your_token")
    
    // 解析歌单
    playlist, _ := client.ParsePlaylist("歌单ID")
    
    fmt.Printf("歌单: %s\n", playlist.Playlist.Name)
    fmt.Printf("歌曲数: %d\n", playlist.Playlist.TrackCount)
    
    // 显示歌曲
    for _, track := range playlist.Playlist.Tracks {
        fmt.Printf("- %s (%s)\n", track.Name, track.Artists[0].Name)
    }
}
```

### 示例4：使用者自己实现缓存

```go
package main

import (
    "sync"
    "time"
    "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

// 用户自己实现缓存
type MyCache struct {
    data map[string]*CacheItem
    mutex sync.RWMutex
}

type CacheItem struct {
    value     interface{}
    expire    time.Time
}

func (c *MyCache) Get(key string) interface{} {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if item, ok := c.data[key]; ok && time.Now().Before(item.expire) {
        return item.value
    }
    return nil
}

func (c *MyCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.data[key] = &CacheItem{
        value:  value,
        expire: time.Now().Add(ttl),
    }
}

func main() {
    client, _ := netease.NewClient("your_token")
    cache := &MyCache{data: make(map[string]*CacheItem)}
    
    // 搜索并缓存
    cacheKey := "search:周杰伦:10"
    
    // 先查缓存
    if cached := cache.Get(cacheKey); cached != nil {
        result := cached.(*netease.SearchResponse)
        fmt.Println("命中缓存")
        // 使用缓存数据
    } else {
        // 调用API
        result, _ := client.Search("周杰伦", 10)
        // 存入缓存
        cache.Set(cacheKey, result, 20*time.Minute)
    }
}
```

---

## 7. 性能优化

### HTTP连接池（核心优化）

```go
// pkg/client/http.go
func NewHTTPClient(token string) *HTTPClient {
    // 使用连接池（复用连接）
    transport := &http.Transport{
        MaxIdleConns:        50,              // 最大空闲连接
        MaxIdleConnsPerHost: 10,              // 每主机最大空闲连接
        IdleConnTimeout:     30 * time.Second, // 空闲超时
        DisableCompression:  false,           // 启用压缩
    }
    
    return &HTTPClient{
        client: &http.Client{
            Transport: transport,
            Timeout:   10 * time.Second,
        },
        token: token,
    }
}
```

### 并发控制

```go
// 使用goroutine池（避免无限制并发）
type WorkerPool struct {
    workers int
    tasks   chan func()
}

func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        workers: workers,
        tasks:   make(chan func(), workers*2),
    }
    
    for i := 0; i < workers; i++ {
        go func() {
            for task := range pool.tasks {
                task()
            }
        }()
    }
    
    return pool
}
```

### 性能指标

| 指标 | 目标值 |
|------|--------|
| 单次响应 | < 500ms |
| 并发能力 | > 500 QPS |
| 内存占用 | < 20MB |
| 编译体积 | < 5MB |

---

## 8. 测试方案

### 单元测试

```go
// tests/music_test.go
func TestParseMusic(t *testing.T) {
    client, _ := netease.NewClient("test_token")
    
    result, err := client.ParseMusic("123456", "standard")
    
    assert.NoError(t, err)
    assert.Equal(t, 200, result.Code)
    assert.NotEmpty(t, result.URL)
}

// tests/search_test.go
func TestSearch(t *testing.T) {
    client, _ := netease.NewClient("test_token")
    
    result, err := client.Search("周杰伦", 10)
    
    assert.NoError(t, err)
    assert.Greater(t, len(result.Data), 0)
}
```

### 性能测试

```go
// tests/bench_test.go
func BenchmarkParseMusic(b *testing.B) {
    client, _ := netease.NewClient("test_token")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.ParseMusic("123456", "standard")
    }
}

func BenchmarkSearch(b *testing.B) {
    client, _ := netease.NewClient("test_token")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.Search("测试", 10)
    }
}
```

### 运行测试

```bash
# 单元测试
go test ./tests

# 性能测试
go test -bench=. ./tests

# 覆盖率
go test -cover ./tests
```

---

## 9. README.md

```markdown
# 网易云音乐API解析库

纯Go实现的网易云音乐解析库，极致简洁。

## 特性
- 🚀 高性能：利用Go并发，提升5-10倍
- 📦 极简：仅3个核心方法
- ⚡ 零依赖：无Redis、无数据库
- 🎯 纯库：无HTTP服务、无缓存（使用者自己实现）

## 安装
```bash
go get github.com/cybaka520/AMLLHub-Music-API
```

## 快速开始
```go
package main

import "github.com/cybaka520/AMLLHub-Music-API/pkg"

func main() {
    client, _ := netease.NewClient("MUSIC_U=xxxxx")
    
    // 搜索
    result, _ := client.Search("周杰伦", 10)
    
    // 解析
    music, _ := client.ParseMusic("123456", "lossless")
    fmt.Println(music.URL)
}
```

## API
- `Search(keywords, limit)` - 搜索音乐
- `ParseMusic(songID, level)` - 解析单曲
- `ParsePlaylist(playlistID)` - 解析歌单

## 音质等级
- `standard` - 标准
- `exhigh` - 极高
- `lossless` - 无损
- `hires` - 高解析度
- `dolby` - 杜比全景声

## License
MIT
```

---

## 10. 构建脚本

```makefile
# Makefile
.PHONY: test build clean

test:
	go test ./tests -v

bench:
	go test -bench=. ./tests

build:
	go build -o netease-api.a

clean:
	rm -f netease-api.a
```

---

## 总结

### 工作量（5天）
- 第1天：HTTP客户端 + 加密（2天任务）
- 第2天：数据模型 + 工具（1天）
- 第3天：搜索 + 单曲解析（1天）
- 第4天：歌单解析 + 主入口（1天）
- 第5天：测试 + 示例 + 文档（1天）

### 最终成果
- 20个Go文件（不含测试）
- 3个核心API方法
- 3个使用示例
- 完整测试覆盖
- 简洁README

### 性能提升
- 相比PHP：5-10倍性能提升
- 编译体积：< 5MB
- 内存占用：< 20MB
- 并发能力：> 500 QPS

**项目地址**: https://github.com/cybaka520/AMLLHub-Music-API