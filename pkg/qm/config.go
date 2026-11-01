// Package qm 提供音乐平台（QM）的解析能力：搜索、歌曲详情、播放直链、歌单、歌词。
//
// 快速开始：
//
//	client := qm.NewClient("qm_keyst=xxx; uin=123456")
//	result, _ := client.Search(context.Background(), "关键词", 10)
//	music, _ := client.ParseMusic(context.Background(), "0039MnYb0qxYhV", "hq")
//
// 本包不缓存任何数据，调用方自行实现业务级缓存。
package qm

import "time"

// 上游端点。地址属于协议的一部分，改动会导致接口失效。
const (
	hostMusicu   = "https://u.y.qq.com/cgi-bin/musicu.fcg"
	hostMusics   = "https://u.y.qq.com/cgi-bin/musics.fcg"
	hostPlaylist = "https://c.y.qq.com/qzone/fcg-bin/fcg_ucc_getcdinfo_byids_cp.fcg"
)

// 两个 UA 不能统一：移动端 UA 是 fcg 免签的前提，Web UA 是 c.y.qq.com 的要求。
const (
	uaFCG      = "QQMusic 14090008(android 15)"
	uaWeb      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	refererWeb = "https://y.qq.com/"
)

// 超时、重试与体积上限。
const (
	defaultTimeout = 30 * time.Second
	callTimeout    = 8 * time.Second
	maxRespBody    = 10 << 20 // 10MB，防异常大包
	maxRetries     = 2
	retryBackoff   = 300 * time.Millisecond
	sessionTTL     = time.Hour
	maxPageSize    = 100
)

// comm 基础值（伪装 Android 客户端）。均为逆向得到的固定值，不可改。
const (
	commCT        = 11
	commCV        = 14090008
	commChid      = "10003505"
	commOSVer     = "15"
	commPhoneType = "24122RKC7C"
	commAppID     = "qqmusic"
	commNetType   = "NETWORK_WIFI"
	commZero      = "0"
)

// 搜索接口的 comm 覆盖值，类型必须是字符串。
const (
	searchCV    = "1003006"
	searchAppID = "qqmusiclight"
)

// 封面模板与默认 CDN 前缀。
const (
	coverURLTemplate = "https://y.gtimg.cn/music/photo_new/T002R%dx%dM000%s.jpg"
	defaultSIP       = "https://isure.stream.qqmusic.qq.com/"
)

// 匿名请求必须带这个 Cookie 头，缺失会被上游判为 2001。
const anonymousCookie = "tmeLoginType=-1;"

// proxyHost 与搜索网关共用地址，单独列出便于排查。
const searchGateway = hostMusics

// levelSpec 描述一个音质档位与直链文件名的对应关系。
type levelSpec struct {
	name   string
	prefix string
	ext    string
}

// levelTable 按音质从高到低排列，降级时按此顺序向后取候选。
//
// 注意：sq 是 128kbps、lq 是 AAC，与常见的“标准/极高”语义相反。
var levelTable = []levelSpec{
	{name: "hi-res", prefix: "AI00", ext: ".flac"},
	{name: "lossless", prefix: "F000", ext: ".flac"},
	{name: "hq", prefix: "M800", ext: ".mp3"},
	{name: "sq", prefix: "M500", ext: ".mp3"},
	{name: "lq", prefix: "C400", ext: ".m4a"},
}

// defaultLevel 未指定音质时的默认档位。
const defaultLevel = "hq"

// levelIndex 返回档位下标，未识别返回 -1。
func levelIndex(name string) int {
	for i, l := range levelTable {
		if l.name == name {
			return i
		}
	}
	return -1
}
