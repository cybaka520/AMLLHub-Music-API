package main

import (
	"context"
	"fmt"

	netease "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
	ctx := context.Background()

	// 创建客户端（仅需Token）
	token := "MUSIC_U=xxxxx; NMTID=xxxxx"
	client := netease.NewClient(token)

	// 搜索音乐
	result, err := client.Search(ctx, "周杰伦", 10)
	if err != nil {
		fmt.Println("搜索失败:", err)
		return
	}

	fmt.Printf("找到 %d 首歌曲\n", len(result.Data))
	for _, song := range result.Data {
		artists := ""
		if len(song.Artists) > 0 {
			artists = song.Artists[0].Name
		}
		fmt.Printf("- %s (%s) [%s]\n", song.Name, artists, song.Duration)
	}

	// 解析单曲
	if len(result.Data) > 0 {
		songID := fmt.Sprintf("%d", result.Data[0].ID)
		music, err := client.ParseMusic(ctx, songID, "lossless")
		if err != nil {
			fmt.Println("解析失败:", err)
			return
		}

		fmt.Println("\n=== 解析结果 ===")
		fmt.Printf("歌曲: %s\n", music.Name)
		fmt.Printf("歌手: %s\n", music.ArtistName)
		fmt.Printf("专辑: %s\n", music.AlbumName)
		fmt.Printf("音质: %s\n", music.Level)
		fmt.Printf("大小: %s\n", music.Size)
		fmt.Printf("URL:  %s\n", music.URL)
	}
}
