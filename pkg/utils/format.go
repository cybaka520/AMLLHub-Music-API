package utils

import "fmt"

// FormatFileSize 格式化文件大小（字节转为人类可读格式）
func FormatFileSize(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	size := float64(bytes)

	for i, unit := range units {
		if size < 1024 || i == len(units)-1 {
			if i == 0 {
				return fmt.Sprintf("%.0f%s", size, unit)
			}
			return fmt.Sprintf("%.2f%s", size, unit)
		}
		size /= 1024
	}

	return fmt.Sprintf("%.2f%s", size, units[len(units)-1])
}

// FormatDuration 格式化时长（毫秒转为 mm:ss 格式）
func FormatDuration(milliseconds int64) string {
	seconds := milliseconds / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// FormatLevel 将音质代码转换为可读的中文描述
func FormatLevel(level string) string {
	levels := map[string]string{
		"standard":  "标准",
		"exhigh":    "极高(HQ)",
		"lossless":  "无损(SQ)",
		"hires":     "高解析度无损(Hi-Res)",
		"jyeffect":  "高清臻音(Spatial Audio)",
		"jymaster":  "超清母带(Master)",
		"sky":       "沉浸环绕声(Surround Audio)",
		"dolby":     "杜比全景声(Dolby Atmos)",
	}

	if formatted, ok := levels[level]; ok {
		return formatted
	}
	return "未知音质"
}
