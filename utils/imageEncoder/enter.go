package utils

import (
	"encoding/base64"
	"io"
	"os"
)

// encodeImageToBase64 将图片文件编码为base64字符串
func EncodeImageToBase64(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 5*1024*1024) // 5MB buffer
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(buf[:n])
	return encoded, nil
}
