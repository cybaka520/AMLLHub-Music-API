package main

import (
	"fmt"
	"sync"

	netease "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
	token := "MUSIC_U=xxxxx"
	client := netease.NewClient(token)

	// 搜索歌曲
	result, err := client.Search("测试歌曲", 50)
	if err != nil {
		fmt.Println("搜索失败:", err)
		return
	}

	// 并发解析（Go的并发优势）
	var wg sync.WaitGroup
	for i, song := range result.Data {
		wg.Add(1)
		go func(idx int, songID string, name string) {
			defer wg.Done()

			music, err := client.ParseMusic(songID, "standard")
			if err == nil {
				fmt.Printf("[%d] %s -> %s\n", idx, name, music.URL)
			}
		}(i, fmt.Sprintf("%d", song.ID), song.Name)
	}
	wg.Wait()

	fmt.Println("批量解析完成")
}
