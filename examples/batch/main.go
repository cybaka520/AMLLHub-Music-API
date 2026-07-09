package main

import (
	"context"
	"fmt"
	"sync"

	netease "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
	ctx := context.Background()
	token := "MUSIC_U=xxxxx"
	client := netease.NewClient(token)

	// 搜索歌曲
	result, err := client.Search(ctx, "测试歌曲", 50)
	if err != nil {
		fmt.Println("搜索失败:", err)
		return
	}

	// 并发解析（限制并发数，避免触发网易云风控）
	const maxConcurrency = 5
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, song := range result.Data {
		songID := fmt.Sprintf("%d", song.ID)
		name := song.Name

		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func(idx int, songID, name string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			music, err := client.ParseMusic(ctx, songID, "standard")
			if err == nil {
				fmt.Printf("[%d] %s -> %s\n", idx, name, music.URL)
			}
		}(i, songID, name)
	}
	wg.Wait()

	fmt.Println("批量解析完成")
}
