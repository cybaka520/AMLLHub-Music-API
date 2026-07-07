# AMLLHub-Music-API

纯 Go 实现的网易云音乐解析库。

## 安装

```bash
go get github.com/cybaka520/AMLLHub-Music-API
```

## 快速开始

```go
package main

import (
    "fmt"
    netease "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
    client := netease.NewClient("MUSIC_U=xxxxx")

    // 搜索音乐
    result, _ := client.Search("周杰伦", 10)
    for _, song := range result.Data {
        fmt.Printf("%s - %s\n", song.Name, song.Artists[0].Name)
    }

    // 解析单曲
    music, _ := client.ParseMusic("123456", "lossless")
    fmt.Printf("URL: %s\n", music.URL)
}
```

## API

### `NewClient(token string) *Client`

创建客户端，需一个 Token。

### `Search(keywords string, limit int) (*SearchResponse, error)`

搜索音乐。

| 参数       | 说明    | 范围    |
| -------- | ----- | ----- |
| keywords | 搜索关键词 | 非空    |
| limit    | 返回数量  | 1-100 |

### `ParseMusic(songID string, level string) (*MusicResponse, error)`

解析单曲，返回播放链接、歌词等信息。

| 参数     | 说明       |
| ------ | -------- |
| songID | 歌曲ID或URL |
| level  | 音质等级     |

### `ParsePlaylist(playlistID string) (*PlaylistResponse, error)`

解析歌单，返回歌单详情和完整歌曲列表。

| 参数         | 说明       |
| ---------- | -------- |
| playlistID | 歌单ID或URL |

## 音质等级

| 代码         | 说明                     |
| ---------- | ---------------------- |
| `standard` | 标准                     |
| `exhigh`   | 极高 (HQ)                |
| `lossless` | 无损 (SQ)                |
| `hires`    | 高解析度无损 (Hi-Res)        |
| `jyeffect` | 高清臻音 (Spatial Audio)   |
| `jymaster` | 超清母带 (Master)          |
| `sky`      | 沉浸环绕声 (Surround Audio) |
| `dolby`    | 杜比全景声 (Dolby Atmos)    |

## 缓存说明

本库不提供缓存功能，调用方可自行实现。推荐使用 `sync.Map` 或 Redis：

```go
client := netease.NewClient(token)

// 自定义缓存逻辑
cacheKey := fmt.Sprintf("search:%s:%d", keywords, limit)
if cached := myCache.Get(cacheKey); cached != nil {
    return cached.(*netease.SearchResponse)
}

result, _ := client.Search(keywords, limit)
myCache.Set(cacheKey, result, 20*time.Minute)
```

## 项目结构

```
pkg/
├── api/           # 核心API实现
│   ├── music.go       # 单曲解析
│   ├── search.go      # 音乐搜索
│   └── playlist.go    # 歌单解析
├── client/        # HTTP客户端
├── crypto/        # 加密工具（AES/MD5）
├── models/        # 数据模型
├── utils/         # 工具函数（IP/UA伪造、格式化、解析）
└── netease.go     # 主入口
```

## 开发

```bash
# 构建
make build

# 运行测试
make test

# 性能测试
make bench

# 代码检查
make vet
```

