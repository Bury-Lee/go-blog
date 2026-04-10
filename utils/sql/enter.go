package sql

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertSliceOrderSql(orderList []uint) string {
	if len(orderList) == 0 {
		return ""
	}

	var idStrs []string
	for _, id := range orderList {
		idStrs = append(idStrs, strconv.FormatUint(uint64(id), 10))
	}

	return fmt.Sprintf("FIELD(id, %s)", strings.Join(idStrs, ","))
}
