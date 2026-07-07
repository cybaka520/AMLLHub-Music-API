// Package utils 提供工具函数
package utils

import (
	_ "embed"
	"math/rand"
	"net"
	"strings"
)

//go:embed data/china_ip.txt
var chinaIPData string

//go:embed data/UserAgent.txt
var userAgentData string

var cidrs []string
var uas []string

func init() {
	// 解析中国IP CIDR
	for _, line := range strings.Split(chinaIPData, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cidrs = append(cidrs, line)
	}

	// 解析 User-Agent
	for _, line := range strings.Split(userAgentData, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		uas = append(uas, line)
	}
}

// RandomIP 从内置的中国IP CIDR列表中随机生成一个IP地址
func RandomIP() string {
	if len(cidrs) == 0 {
		return "116.25.146.37"
	}

	cidr := cidrs[rand.Intn(len(cidrs))]
	parts := strings.SplitN(cidr, "/", 2)
	if len(parts) != 2 {
		return "116.25.146.37"
	}

	base := parts[0]
	mask := 0
	for _, c := range parts[1] {
		mask = mask*10 + int(c-'0')
	}

	ip := net.ParseIP(base).To4()
	if ip == nil {
		return "116.25.146.37"
	}

	ipUint := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	bits := 32 - mask
	if bits <= 0 {
		return base
	}

	minIP := ipUint + 1
	maxIP := ipUint + (1 << bits) - 2
	if maxIP <= minIP {
		return base
	}

	result := minIP + uint32(rand.Intn(int(maxIP-minIP)))
	return net.IPv4(byte(result>>24), byte(result>>16), byte(result>>8), byte(result)).String()
}

// RandomUA 从内置的User-Agent列表中随机返回一个
func RandomUA() string {
	if len(uas) == 0 {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	}
	return uas[rand.Intn(len(uas))]
}
