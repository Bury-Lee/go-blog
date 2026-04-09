package Hash

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 生成密码哈希
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10，可根据性能需求调整（4-31）
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
