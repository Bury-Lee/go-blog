package utils_other

import (
	"testing"
)

// TestIsImage 测试图片格式判断功能
// 说明:测试有效图片、无效图片、无后缀名、短文件名等边界情况
func TestIsImage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"Valid JPG", "test.jpg", true},
		{"Valid JPEG", "photo.jpeg", true},
		{"Valid PNG uppercase", "IMAGE.PNG", true},
		{"Valid WEBP", "pic.webp", true},
		{"Invalid extension", "test.txt", false},
		{"No extension", "testimage", false},
		{"Dot at end", "test.", false},
		{"Too short", "a.b", false},
		{"Extension too long", "test.abcdef", false},
		{"Multiple dots", "archive.tar.gz", false},
		{"Multiple dots valid", "my.photo.jpg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsImage(tt.filename); got != tt.want {
				t.Errorf("IsImage(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// TestToLower 测试字符串转小写功能
// 说明:测试包含大写、小写、数字和特殊字符的字符串
func TestToLower(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"All uppercase", "HELLO", "hello"},
		{"Mixed case", "HeLlO", "hello"},
		{"All lowercase", "hello", "hello"},
		{"With numbers", "HELLO123", "hello123"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLower(tt.input); got != tt.want {
				t.Errorf("ToLower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestToUpper 测试字符串转大写功能
// 说明:测试包含大写、小写、数字和特殊字符的字符串
func TestToUpper(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"All lowercase", "hello", "HELLO"},
		{"Mixed case", "HeLlO", "HELLO"},
		{"All uppercase", "HELLO", "HELLO"},
		{"With numbers", "hello123", "HELLO123"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUpper(tt.input); got != tt.want {
				t.Errorf("ToUpper(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
