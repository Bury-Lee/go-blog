package Hash

import (
	"testing"
)

// TestHashPasswordAndCheck 测试密码哈希生成和验证
// 说明: 验证生成的哈希能够正确匹配原密码，并且不匹配错误密码
func TestHashPasswordAndCheck(t *testing.T) {
	password := "my_secret_password_123"
	
	// 1. 生成哈希
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("生成哈希失败: %v", err)
	}
	
	// 哈希值不应该为空且不能等于原密码
	if hash == "" || hash == password {
		t.Errorf("生成的哈希值不合法: %s", hash)
	}
	
	// 2. 验证正确的密码
	isValid := CheckPassword(password, hash)
	if !isValid {
		t.Errorf("预期验证成功，但实际验证失败")
	}
	
	// 3. 验证错误的密码
	wrongPassword := "wrong_password"
	isInvalid := CheckPassword(wrongPassword, hash)
	if isInvalid {
		t.Errorf("预期验证失败，但实际验证成功")
	}
}
