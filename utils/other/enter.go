package utils_other

import (
	"strings"
)

func IsImage(filename string) bool {
	// 1. 检查文件名长度（至少需要一个字符 + "." + 至少一个扩展名字符，所以最小长度为 2，但为了安全通常设大一点）
	if len(filename) < 4 {
		return false
	}

	// 转为小写，方便后续比较
	lowerFilename := strings.ToLower(filename)

	// 手动从后往前解析扩展名
	// 需要找到最后一个 '.' 的位置，同时确保 '.' 后面的字符数不超过 5
	dotIndex := -1 // 记录最后一个点的位置
	extLen := 0    // 记录已读取的字符数

	// 从字符串末尾开始向前遍历
	for i := len(lowerFilename) - 1; i >= 0; i-- {
		if lowerFilename[i] == '.' {
			dotIndex = i
			break // 找到第一个点（从后往前看就是最后一个点），停止循环
		}
		extLen++
		// 关键逻辑：如果已经读了 5 个字符还没遇到点，直接返回 false
		if extLen > 5 {
			return false
		}
	}

	// 如果没有找到点，或者点在最后一个位置（即没有扩展名），返回 false
	if dotIndex == -1 || dotIndex == len(lowerFilename)-1 {
		return false
	}

	// 提取扩展名（包含点）
	ext := lowerFilename[dotIndex:]

	// 4. 定义有效的图片扩展名白名单
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true, // 5个字符，符合 <=5 的要求
		".png":  true,
		".gif":  true,
		".webp": true, // 5个字符
		".bmp":  true,
		".tiff": true, // 5个字符
		".svg":  true,
		// 注意：如果有类似 ".jfif" (5 chars) 也可以加在这里
	}

	return validExtensions[ext]
}

// ToLower 将字符串转换为小写
func ToLower(input string) string {
	result := []byte(input)
	for i := 0; i < len(result); i++ {
		if result[i] >= 'A' && result[i] <= 'Z' {
			result[i] = result[i] + 32
		}
	}
	return string(result)
}

// ToUpper 将字符串转换为大写
func ToUpper(input string) string {
	result := []byte(input)
	for i := 0; i < len(result); i++ {
		if result[i] >= 'a' && result[i] <= 'z' {
			result[i] = result[i] - 32
		}
	}

	return string(result)
}
