package sql

import (
	"StarDreamerCyberNook/global"
	"fmt"
	"strconv"
	"strings"
)

func ConvertSliceOrderSql(orderList []uint) string {
	if len(orderList) == 0 {
		return ""
	}

	var dbType = global.Config.DB[0].DBName

	switch strings.ToLower(dbType) {
	case "mysql", "mariadb":
		// MySQL: FIELD(id, 1, 2, 3)
		var idStrs []string
		for _, id := range orderList {
			idStrs = append(idStrs, strconv.FormatUint(uint64(id), 10))
		}
		return fmt.Sprintf("FIELD(id, %s)", strings.Join(idStrs, ","))

	case "postgres", "postgresql":
		// PostgreSQL: CASE id WHEN 1 THEN 0 WHEN 2 THEN 1 ELSE 2 END
		var sb strings.Builder
		sb.WriteString("CASE id ")
		for i, id := range orderList {
			sb.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, i))
		}
		sb.WriteString(fmt.Sprintf("ELSE %d END", len(orderList)))
		return sb.String()

	case "sqlite", "sqlite3":
		// SQLite: 使用 CASE 语法（与 PostgreSQL 相同）
		var sb strings.Builder
		sb.WriteString("CASE id ")
		for i, id := range orderList {
			sb.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, i))
		}
		sb.WriteString(fmt.Sprintf("ELSE %d END", len(orderList)))
		return sb.String()

	default:
		// 默认使用 CASE 语法（通用）
		var sb strings.Builder
		sb.WriteString("CASE id ")
		for i, id := range orderList {
			sb.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, i))
		}
		sb.WriteString(fmt.Sprintf("ELSE %d END", len(orderList)))
		return sb.String()
	}
}
