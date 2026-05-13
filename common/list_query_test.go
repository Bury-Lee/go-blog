package common

import (
	"StarDreamerCyberNook/global"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestModel 测试用的模型
// 参数:无
// 返回:无
// 说明:用于数据库相关单元测试
type TestModel struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"size:100"`
	Age     int
	Profile TestProfile `gorm:"foreignKey:TestModelID"`
}

// TestProfile 测试用的关联模型
type TestProfile struct {
	ID          uint `gorm:"primaryKey"`
	TestModelID uint
	Description string
}

func setupTestDB() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&TestModel{}, &TestProfile{})
	global.DB = db.Debug() //开启调试

	// 插入测试数据
	db.Create(&TestModel{Name: "Alice", Age: 20, Profile: TestProfile{Description: "Alice Profile"}})
	db.Create(&TestModel{Name: "Bob", Age: 25, Profile: TestProfile{Description: "Bob Profile"}})
	db.Create(&TestModel{Name: "Charlie", Age: 30, Profile: TestProfile{Description: "Charlie Profile"}})
	db.Create(&TestModel{Name: "David", Age: 35, Profile: TestProfile{Description: "David Profile"}})
	db.Create(&TestModel{Name: "Eve", Age: 40, Profile: TestProfile{Description: "Eve Profile"}})
}

// TestPageInfo_GetLimit 测试 GetLimit 方法
// 参数:t - testing.T 测试上下文
// 返回:无
// 说明:测试各种边界条件和正常情况下的限制数量
func TestPageInfo_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		pageInfo PageInfo
		expected int
	}{
		{"正常范围", PageInfo{Limit: 20}, 20},
		{"小于1", PageInfo{Limit: 0}, 10},
		{"大于40", PageInfo{Limit: 50}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pageInfo.GetLimit(); got != tt.expected {
				t.Errorf("PageInfo.GetLimit() = %v, 期望 %v", got, tt.expected)
			}
		})
	}
}

// TestPageInfo_GetPage 测试 GetPage 方法
// 参数:t - testing.T 测试上下文
// 返回:无
// 说明:测试页码的有效范围和越界情况
func TestPageInfo_GetPage(t *testing.T) {
	tests := []struct {
		name     string
		pageInfo PageInfo
		expected int
	}{
		{"正常范围", PageInfo{Page: 2}, 2},
		{"小于1", PageInfo{Page: 0}, 1},
		{"大于30", PageInfo{Page: 40}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pageInfo.GetPage(); got != tt.expected {
				t.Errorf("PageInfo.GetPage() = %v, 期望 %v", got, tt.expected)
			}
		})
	}
}

// TestPageInfo_GetOffset 测试 GetOffset 方法
// 参数:t - testing.T 测试上下文
// 返回:无
// 说明:测试根据页码和限制计算出的偏移量
func TestPageInfo_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		pageInfo PageInfo
		expected int
	}{
		{"第1页_10条", PageInfo{Page: 1, Limit: 10}, 0},
		{"第2页_10条", PageInfo{Page: 2, Limit: 10}, 10},
		{"第3页_15条", PageInfo{Page: 3, Limit: 15}, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pageInfo.GetOffset(); got != tt.expected {
				t.Errorf("PageInfo.GetOffset() = %v, 期望 %v", got, tt.expected)
			}
		})
	}
}

// TestListQuery 测试通用分页查询函数
// 参数:t - testing.T 测试上下文
// 返回:无
// 说明:测试基础查询、模糊匹配、排序、游标分页、预加载等功能
func TestListQuery(t *testing.T) {
	setupTestDB()

	t.Run("基础查询", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 2},
		}
		list, count, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count != 5 {
			t.Errorf("总数 = %v, 期望 5", count)
		}
		if len(list) != 2 {
			t.Errorf("列表长度 = %v, 期望 2", len(list))
		}
		// 默认排序为 id DESC
		if len(list) > 0 && list[0].Name != "Eve" {
			t.Errorf("默认排序错误, 第一个应该是 Eve, 实际是 %v", list[0].Name)
		}
	})

	t.Run("基础查询-带有Model条件", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 10},
		}
		list, count, err := ListQuery(&TestModel{Age: 25}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count != 1 {
			t.Errorf("带有Model条件的查询总数 = %v, 期望 1", count)
		}
		if len(list) != 1 || list[0].Name != "Bob" {
			t.Errorf("带有Model条件的查询结果错误")
		}
	})

	t.Run("模糊匹配-单字段", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 10, Key: "li"},
			Likes:    []string{"name"},
		}
		list, count, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count != 2 { // Alice 和 Charlie 包含 'li'
			t.Errorf("模糊匹配总数 = %v, 期望 2", count)
		}
		if len(list) != 2 {
			t.Errorf("模糊匹配列表长度 = %v, 期望 2", len(list))
		}
	})

	t.Run("模糊匹配-多字段", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 10, Key: "a"},
			Likes:    []string{"name", "age"}, // Sqlite会将age转为字符串匹配
		}
		list, count, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count < 1 {
			t.Errorf("多字段模糊匹配失败, 期望有结果, 实际 0")
		}
		if len(list) < 1 {
			t.Errorf("多字段模糊匹配列表为空")
		}
	})

	t.Run("排序测试-传入Order", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 10, Order: "age asc"},
		}
		list, count, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count < 5 {
			t.Errorf("总数 = %v, 期望 >= 5", count)
		}
		if list[0].Name != "Alice" { // Age 20 是最小
			t.Errorf("排序结果错误, 第一个应该是 Alice, 实际是 %v", list[0].Name)
		}
	})

	t.Run("排序测试-DefaultOrder", func(t *testing.T) {
		options := Options{
			PageInfo:     PageInfo{Page: 1, Limit: 10},
			DefaultOrder: "age asc",
		}
		list, _, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if list[0].Name != "Alice" {
			t.Errorf("默认排序结果错误, 第一个应该是 Alice, 实际是 %v", list[0].Name)
		}
	})

	t.Run("定制化查询", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 10},
			Where:    global.DB.Where("age > ?", 25),
		}
		list, count, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count != 3 { // Charlie, David, Eve
			t.Errorf("定制化查询总数 = %v, 期望 3", count)
		}
		if len(list) != 3 {
			t.Errorf("定制化查询列表长度 = %v, 期望 3", len(list))
		}
	})

	t.Run("预加载查询", func(t *testing.T) {
		options := Options{
			PageInfo: PageInfo{Page: 1, Limit: 10},
			Preloads: []string{"Profile"},
		}
		list, count, err := ListQuery(TestModel{}, options)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count < 5 {
			t.Errorf("预加载查询总数 = %v, 期望 >= 5", count)
		}
		if len(list) > 0 && list[0].Profile.Description == "" {
			t.Errorf("预加载失败, Profile为空")
		}
	})

	t.Run("游标分页模式", func(t *testing.T) {
		// 先查出最后一个作为游标
		options1 := Options{
			PageInfo: PageInfo{Page: 1, Limit: 2},
		}
		list1, _, _ := ListQuery(TestModel{}, options1)

		// 游标模式，应该返回小于这个id的数据
		options2 := Options{
			PageInfo: PageInfo{EndId: list1[len(list1)-1].ID, Limit: 2},
		}
		list2, count2, err := ListQuery(TestModel{}, options2)
		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}
		if count2 < 5 {
			t.Errorf("游标分页查询总数应包含所有记录总数 = %v, 期望 >= 5", count2)
		}
		if len(list2) == 0 {
			t.Errorf("游标分页返回列表为空")
		}
		if list2[0].ID >= list1[len(list1)-1].ID {
			t.Errorf("游标分页未生效, 返回的ID没有小于游标ID")
		}
	})
}
