package sql

import (
	"fmt"
)

func ConvertSliceSql(orderList []uint) string {
	result := "("
	for i, v := range orderList {
		result += fmt.Sprintf("%d", v)
		if i < len(orderList)-1 {
			result += ","
		}
	}
	result += ")"
	return result
}

func ConvertSliceOrderSql(orderList []uint) string {
	result := ""
	for i, v := range orderList {
		if i == len(orderList)-1 {
			result += fmt.Sprintf("id = %d desc", v)
			break
		}
		result += fmt.Sprintf("id = %d dasc,", v)
	}
	return result
}
