package utils_other

import (
	"fmt"
	"strings"
)

// 使用 Base62: 0-9, A-Z, a-z
// 这样既没有 '/' 也没有 '+'，非常适合做路径的一部分，且排序友好
const (
	baseChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Separator = '/'
	Base      = 62
)

var unMap = map[byte]uint{
	// 0-9
	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	// A-Z (10-35)
	'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15, 'G': 16, 'H': 17, 'I': 18, 'J': 19,
	'K': 20, 'L': 21, 'M': 22, 'N': 23, 'O': 24, 'P': 25, 'Q': 26, 'R': 27, 'S': 28, 'T': 29,
	'U': 30, 'V': 31, 'W': 32, 'X': 33, 'Y': 34, 'Z': 35,
	// a-z (36-61)
	'a': 36, 'b': 37, 'c': 38, 'd': 39, 'e': 40, 'f': 41, 'g': 42, 'h': 43, 'i': 44, 'j': 45,
	'k': 46, 'l': 47, 'm': 48, 'n': 49, 'o': 50, 'p': 51, 'q': 52, 'r': 53, 's': 54, 't': 55,
	'u': 56, 'v': 57, 'w': 58, 'x': 59, 'y': 60, 'z': 61,
}

// UintEncodePath 将 uint ID 转换为 Base62 字符串并拼接到 basePath 后
// 结果示例: /user/A1b (高位在前，符合人类阅读和字典序排序)
func EncodePath(basePath string, parentID uint) string {
	if parentID == 0 {
		// 0 的特殊处理，对应 baseChars[0] 即 '0'
		return basePath + string(Separator) + "0"
	}

	var buffer []byte

	// 标准进制转换：不断取余，得到的是低位在前
	for parentID > 0 {
		buffer = append(buffer, baseChars[parentID%Base])
		parentID /= Base
	}
	// 但这通常没问题，只要它能唯一还原且不含非法字符。

	// 反转方便阅读,可以在调试的时候用
	// for i, j := 0, len(buffer)-1; i < j; i, j = i+1, j-1 {
	// 	buffer[i], buffer[j] = buffer[j], buffer[i]
	// }

	var sb strings.Builder
	sb.Grow(len(basePath) + 1 + len(buffer))
	sb.WriteString(basePath)
	sb.WriteByte(Separator)
	sb.Write(buffer)

	return sb.String()
}

func DecodePath(path string) (uint, error) {
	// 1. 找到分隔符位置
	sepIdx := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == Separator {
			sepIdx = i
			break
		}
	}

	if sepIdx == -1 || sepIdx == len(path)-1 {
		return 0, fmt.Errorf("路径无法解码")
	}

	idx := sepIdx + 1
	var result uint

	// 小端序权重基数，从 1 (62^0) 开始，每循环一次乘以 62
	var powerOfBase uint = 1
	// 2. 遍历每一位字符 (从左到右，即从低位到高位)
	for i := idx; i < len(path); i++ {
		char := path[i]
		val, ok := unMap[char]
		if !ok {
			return 0, fmt.Errorf("invalid char: %c", char)
		}
		// 小端序核心公式：结果 += 值 * (62 的 n 次方)
		result += val * powerOfBase
		// 更新下一位的权重 (62^1, 62^2, ...)
		// 注意：如果数字非常大，这里可能会溢出，但在 uint 范围内通常够用
		powerOfBase *= Base
	}
	return result, nil
}

// DecodeRootPath 解码路径中的第一段（根节点）
// 示例: 输入 "/a/c/v/d"，提取 "a" 并按小端序解码返回 uint
func DecodeRootPath(path string) (uint, error) {
	// 1. 跳过开头的分隔符（如果有）
	start := 0
	if len(path) > 0 && path[0] == Separator {
		start = 1
	}
	if start >= len(path) {
		return 0, fmt.Errorf("空路径")
	}
	// 2. 找到第一个分隔符的位置，确定第一段的结束位置
	end := -1
	for i := start; i < len(path); i++ {
		if path[i] == Separator {
			end = i
			break
		}
	}
	// 如果没有找到分隔符，说明整个剩余部分就是第一段
	if end == -1 {
		end = len(path)
	}
	// 如果第一段为空
	if end == start {
		return 0, fmt.Errorf("根路径为空")
	}
	// 3. 按小端序逻辑解码第一段 (逻辑同原 DecodePath)
	var result uint
	var powerOfBase uint = 1
	for i := start; i < end; i++ {
		char := path[i]
		val, ok := unMap[char]
		if !ok {
			return 0, fmt.Errorf("invalid char: %c", char)
		}
		result += val * powerOfBase
		powerOfBase *= Base
	}
	return result, nil
}
