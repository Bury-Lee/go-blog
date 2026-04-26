package utils_other

import (
	"reflect"
	"testing"
)

// TestStructToMap 测试结构体转换为Map的功能
// 说明: 验证普通字段、指针字段、忽略字段（-）以及JSON嵌套结构的处理
func TestStructToMap(t *testing.T) {
	type Address struct {
		City   string `json:"city"`
		Street string `json:"street"`
	}

	type User struct {
		ID       int      `json:"id"`
		Name     string   `json:"name"`
		Age      *int     `json:"age"`
		Password string   `json:"-"`
		Addr     Address  `json:"address"`
	}

	age := 25
	user := User{
		ID:       1,
		Name:     "张三",
		Age:      &age,
		Password: "secret_password",
		Addr: Address{
			City:   "北京",
			Street: "朝阳区",
		},
	}

	// 1. 测试正常结构体转换
	result := StructToMap(user, "json")
	
	// 验证 ID
	if result["id"] != 1 {
		t.Errorf("预期 id=1, 实际得到: %v", result["id"])
	}

	// 验证 Name
	if result["name"] != "张三" {
		t.Errorf("预期 name=张三, 实际得到: %v", result["name"])
	}

	// 验证 Age (应该解引用)
	if result["age"] != 25 {
		t.Errorf("预期 age=25, 实际得到: %v", result["age"])
	}

	// 验证 Password (应该被忽略)
	if _, exists := result["-"]; exists {
		t.Errorf("预期忽略标签为 '-' 的字段，但仍然存在")
	}

	// 验证 Address (应该被序列化为JSON字符串)
	expectedAddrJSON := `{"city":"北京","street":"朝阳区"}`
	if result["address"] != expectedAddrJSON {
		t.Errorf("预期 address=%s, 实际得到: %v", expectedAddrJSON, result["address"])
	}

	// 2. 测试指针结构体
	resultPtr := StructToMap(&user, "json")
	if !reflect.DeepEqual(result, resultPtr) {
		t.Errorf("传入指针和值的结果不一致")
	}

	// 3. 测试非结构体类型 (应返回空或忽略)
	resultNonStruct := StructToMap("just_a_string", "json")
	if len(resultNonStruct) != 0 {
		t.Errorf("预期非结构体返回空map, 实际返回长度: %d", len(resultNonStruct))
	}
}
