package utils_other

import (
	"reflect"
	"testing"
)

// TestReverseArray 测试数组反转功能
// 说明:测试不同类型和长度的数组反转，包括空数组、奇数个元素、偶数个元素
func TestReverseArray(t *testing.T) {
	t.Run("Reverse Int Array", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		want := []int{5, 4, 3, 2, 1}
		got := ReverseArray(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ReverseArray() = %v, want %v", got, want)
		}
	})

	t.Run("Reverse String Array Even", func(t *testing.T) {
		input := []string{"a", "b", "c", "d"}
		want := []string{"d", "c", "b", "a"}
		got := ReverseArray(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ReverseArray() = %v, want %v", got, want)
		}
	})

	t.Run("Reverse Empty Array", func(t *testing.T) {
		var input []int
		var want []int
		got := ReverseArray(input)
		if !reflect.DeepEqual(got, want) && !(len(got) == 0 && len(want) == 0) {
			t.Errorf("ReverseArray() = %v, want %v", got, want)
		}
	})

	t.Run("Reverse Single Element Array", func(t *testing.T) {
		input := []float64{3.14}
		want := []float64{3.14}
		got := ReverseArray(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ReverseArray() = %v, want %v", got, want)
		}
	})
}
