package utils

import (
	"StarDreamerCyberNook/global"

	"encoding/base64"
	"io"
	"os"
	"strings"
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

func ImageSuffixJudge(filename string) (string, bool) {
	_list := strings.Split(filename, ".")
	var suffix string
	if len(_list) == 1 {
		return suffix, false
	}
	suffix = _list[len(_list)-1]
	if !InList(suffix, global.Config.Upload.WhiteList) {
		return suffix, false
	}
	return suffix, true
}

func GetContentType(suffix string) string {
	switch strings.ToLower(suffix) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
