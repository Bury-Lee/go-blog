package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const (
	Digits  = "0123456789"                  //数字
	Lower   = "abcdefghijklmnopqrstuvwxyz"  //小写字母
	Upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"  //大写字母
	Symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?/" //特殊字符

	// 组合使用
	AlphaNum = Digits + Lower + Upper
	All      = AlphaNum + Symbols
)

func InList[T comparable](key T, mapList map[T]struct{}) bool {
	if _, ok := mapList[key]; ok {
		return true
	}
	return false
}

func Md5(data []byte) string {
	md5New := md5.New()
	md5New.Write(data)
	result := hex.EncodeToString(md5New.Sum(nil))
	return result
}

// GetRandomString 生成随机字符串,生成大小写字母组合的n位字符串
func GetRandomString(Len int, sorce string) string {
	// 定义字符集：a-z, A-Z

	result := make([]byte, Len)
	charsetLen := big.NewInt(int64(len(sorce)))

	for i := 0; i < Len; i++ {
		// crypto/rand.Int 返回一个 [0, max) 之间的随机数
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return ""
		}
		result[i] = sorce[num.Int64()]
	}
	return string(result)
}

func GetRandomInt(n int) (int64, error) {
	if n <= 1 {
		return 0, errors.New("随机数长度太短")
	}
	Len := 10
	for i := 1; i < n; i++ {
		Len *= 10
	}
	max := big.NewInt(int64(Len))
	result, err := rand.Int(rand.Reader, max)
	fmt.Println("随机数:", result.Int64()) // 0-99
	return result.Int64(), err
}
