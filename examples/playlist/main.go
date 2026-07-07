package main

import (
	"fmt"

	netease "github.com/cybaka520/AMLLHub-Music-API/pkg"
)

func main() {
	token := "MUSIC_U=xxxxx"
	client := netease.NewClient(token)

	// 解析歌单
	playlist, err := client.ParsePlaylist("歌单ID")
	if err != nil {
		fmt.Println("解析失败:", err)
		return
	}

	fmt.Printf("歌单: %s\n", playlist.Playlist.Name)
	fmt.Printf("创建者: %s\n", playlist.Playlist.Creator.Nickname)
	fmt.Printf("歌曲数: %d\n", playlist.Playlist.TrackCount)

	// 显示歌曲列表
	for i, track := range playlist.Playlist.Tracks {
		artistName := ""
		if len(track.Ar) > 0 {
			artistName = track.Ar[0].Name
		}
		fmt.Printf("%d. %s - %s\n", i+1, track.Name, artistName)
	}
}
