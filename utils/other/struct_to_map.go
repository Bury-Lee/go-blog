package utils_other

import (
	"encoding/json"
	"reflect"
)

// StructToMap 将结构体转换为 map[string]any
// 该函数利用反射机制，提取结构体字段值并映射到 Map 中。
//
// 参数:
//   - data: 需要转换的结构体实例 (必须是结构体类型)
//   - t:    指定用于作为 Map Key 的结构体标签名称 (例如: "json" 或 "form")
//
// 返回:
//   - map[string]any: 转换后的映射表
func StructToMap(data any, t string) (mp map[string]any) {
	mp = make(map[string]any)

	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		tag := fieldType.Tag.Get(t)
		if tag == "" || tag == "-" {
			continue
		}

		value, ok := normalizeFieldValue(field)
		if !ok {
			continue
		}
		mp[tag] = value
	}
	return
}

func normalizeFieldValue(v reflect.Value) (any, bool) {
	for v.IsValid() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil, false
	}

	if shouldMarshalAsJSON(v) {
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return nil, false
		}
		return string(b), true
	}
	return v.Interface(), true
}

func shouldMarshalAsJSON(v reflect.Value) bool {
	t := v.Type()
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return false
	}
	if v.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return false
	}

	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		return true
	default:
		return false
	}
}
