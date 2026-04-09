package ES

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/olivere/elastic/v7"
	pbe "github.com/withlin/canal-go/protocol/entry"
	"google.golang.org/protobuf/proto"
)

var ESClient *elastic.Client

func ConnectES() error {
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
		elastic.SetBasicAuth("elastic", "es"),
	)
	if err != nil {
		return err
	}
	ESClient = client
	return nil
}

// ArticleDocument 对应 Elasticsearch 的 Mapping 结构
// 确保字段名和 json tag 与你的 ES 定义完全一致
// ... (前面的 import 保持不变)

// ArticleDocument
// 【重要】这里的 json tag 必须严格对应 Elasticsearch 的 mapping (下划线风格)
// 即使你的 GORM 结构体用的是驼峰，这里发给 ES 时必须转回下划线
type ArticleDocument struct {
	ID           int      `json:"id"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	Title        string   `json:"title"`
	Abstract     string   `json:"abstract"`
	Content      string   `json:"content"`
	CategoryID   int      `json:"category_id"` // 匹配 ES: category_id
	TagList      []string `json:"tag_list"`    // 匹配 ES: tag_list
	Cover        string   `json:"cover"`
	UserID       int      `json:"user_id"`       // 匹配 ES: user_id
	LookCount    int      `json:"look_count"`    // 匹配 ES: look_count
	DiggCount    int      `json:"digg_count"`    // 匹配 ES: digg_count
	CommentCount int      `json:"comment_count"` // 匹配 ES: comment_count
	CollectCount int      `json:"collect_count"` // 匹配 ES: collect_count
	Status       int      `json:"status"`
	OpenComment  bool     `json:"open_comment"` // 匹配 ES: open_comment
}

// getColumnValueByName 辅助函数：从列列表中根据名称获取值
func getColumnValueByName(columns []*pbe.Column, name string) string {
	for _, col := range columns {
		if col.GetName() == name {
			return col.GetValue()
		}
	}
	return ""
}

// checkError 辅助函数
func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "发生致命错误：%s\n", err.Error())
		os.Exit(1)
	}
}

// processAndConvertEntry 解析数据并转换为 ES 所需的格式
func ProcessAndConvertEntry(entrys []pbe.Entry) {
	for _, entry := range entrys {
		if entry.GetEntryType() == pbe.EntryType_TRANSACTIONBEGIN || entry.GetEntryType() == pbe.EntryType_TRANSACTIONEND {
			continue
		}

		rowChange := new(pbe.RowChange)
		err := proto.Unmarshal(entry.GetStoreValue(), rowChange)
		if err != nil {
			log.Printf("Unmarshal error: %v", err)
			continue
		}

		if rowChange == nil {
			continue
		}

		eventType := rowChange.GetEventType()
		header := entry.GetHeader()

		// 简单日志：显示操作类型
		action := ""
		if eventType == pbe.EventType_INSERT {
			action = "INDEX (插入/更新)"
		} else if eventType == pbe.EventType_UPDATE {
			action = "UPDATE (部分更新)"
		} else if eventType == pbe.EventType_DELETE {
			action = "DELETE (删除)"
		}

		fmt.Printf("\n[%s] 表：%s.%s\n", action, header.GetSchemaName(), header.GetTableName())

		for _, rowData := range rowChange.GetRowDatas() {
			var columns []*pbe.Column
			var docID string

			// 1. 确定数据源 (删除用 Before，其他用 After)
			if eventType == pbe.EventType_DELETE {
				columns = rowData.GetBeforeColumns()
				// 删除操作只需要找到 ID
				docID = getColumnValueByName(columns, "id")
				if docID == "" {
					log.Println("警告：删除操作中未找到 ID 字段")
					continue
				}
				fmt.Printf("-> 准备删除 ES 文档 ID: %s\n", docID)
				DeleteByID(docID)
				// 在这里调用 esClient.Delete(...)
				continue
			} else {
				// INSERT 和 UPDATE
				columns = rowData.GetAfterColumns()

				// 2. 【核心功能】自动转换并构建结构体
				doc, err := ConvertToArticleDoc(columns)
				if err != nil {
					log.Printf("数据转换失败: %v", err)
					continue
				}

				// 3. 序列化为 JSON (这就是可以直接发给 ES 的数据)
				jsonData, err := json.Marshal(doc)
				if err != nil {
					log.Printf("JSON 序列化失败: %v", err)
					continue
				}

				docID = strconv.Itoa(doc.ID)
				CreateFromJSON(docID, string(jsonData))
				fmt.Printf("-> 转换成功 (ES Doc ID: %s):\n%s\n", docID, string(jsonData))

				// 在这里调用 esClient.Index(...) 传入 jsonData
			}
		}
	}
}

// convertToArticleDoc 将 Canal 获取的 Column 列表转换为 ArticleDocument
func ConvertToArticleDoc(columns []*pbe.Column) (ArticleDocument, error) {
	var doc ArticleDocument

	// 构建列名 -> 值的映射，方便查找
	// 数据库列名通常是下划线风格 (e.g., category_id)
	colMap := make(map[string]string)
	for _, col := range columns {
		colMap[col.GetName()] = col.GetValue()
	}

	// 辅助函数：获取字符串
	getStr := func(key string) string {
		if v, ok := colMap[key]; ok {
			return v
		}
		return ""
	}

	// 辅助函数：安全获取整数
	getInt := func(key string) int {
		v := getStr(key)
		if v == "" {
			return 0
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			// 记录日志但不中断，返回 0
			// log.Printf("警告：字段 %s 的值 '%s' 无法转换为整数", key, v)
			return 0
		}
		return i
	}

	// 辅助函数：安全获取布尔值
	// 数据库可能存 "0"/"1", "true"/"false", "t"/"f"
	getBool := func(key string) bool {
		v := strings.ToLower(getStr(key))
		return v == "1" || v == "true" || v == "t"
	}

	// --- 开始映射 (键名使用数据库列名：下划线风格) ---

	// 1. 主键
	doc.ID = getInt("id")
	if doc.ID == 0 && getStr("id") == "" {
		return doc, fmt.Errorf("缺少主键 id，无法构建文档")
	}

	// 2. 时间字段 (直接作为字符串传入 ES，ES 会根据 mapping 格式解析)
	doc.CreatedAt = getStr("created_at")
	doc.UpdatedAt = getStr("updated_at")

	doc.CreatedAt = doc.CreatedAt[:19]
	doc.UpdatedAt = doc.UpdatedAt[:19]

	// 3. 文本字段
	doc.Title = getStr("title")
	doc.Abstract = getStr("abstract")
	doc.Content = getStr("content")
	doc.Cover = getStr("cover")

	// 4. 数值字段 (注意：这里读取的是数据库列名，即下划线风格)
	doc.CategoryID = getInt("category_id")
	doc.UserID = getInt("user_id")
	doc.LookCount = getInt("look_count")
	doc.DiggCount = getInt("digg_count")
	doc.CommentCount = getInt("comment_count")
	doc.CollectCount = getInt("collect_count")
	doc.Status = getInt("status")

	// 5. 布尔字段
	doc.OpenComment = getBool("open_comment")

	// 6. 特殊处理：TagList
	// GORM 的 serializer:json 会将 []string 存为 JSON 字符串，例如：["go","es"]
	tagStr := getStr("tag_list")
	if tagStr != "" && tagStr != "null" {
		var tags []string
		// 尝试标准 JSON 反序列化
		if err := json.Unmarshal([]byte(tagStr), &tags); err != nil {
			// 降级策略：如果不是合法 JSON，尝试按逗号分割（兼容旧数据或错误数据）
			log.Printf("警告：tag_list 不是合法 JSON，尝试按逗号分割: %s", tagStr)
			rawTags := strings.Split(tagStr, ",")
			tags = make([]string, 0, len(rawTags))
			for _, t := range rawTags {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
		}
		doc.TagList = tags
	} else {
		// 空值初始化为空切片，避免 ES 中为 null
		doc.TagList = []string{}
	}

	return doc, nil
}
