package es_service

import (
	"StarDreamerCyberNook/global"
	"context"

	"github.com/sirupsen/logrus"
)

// TODO:es升级到v8
func Update(index, mapping string) { //如果索引存在就删除,然后进行创建或更新
	if ExistsIndex(index) {
		DeleteIndex(index)
	}
	CreateIndex(index, mapping)

}

// CreateIndex 创建索引
// 注意：如果索引已存在，直接执行会报错。建议先调用 ExistsIndex 判断。
func CreateIndex(index, mapping string) {
	_, err := global.ES.CreateIndex(index).BodyString(mapping).Do(context.Background())
	if err != nil {
		logrus.Errorf("ES创建索引失败:%s - %s", index, err.Error())
		return
	}
	logrus.Infof("ES创建索引成功:%s\n\t%s", index, mapping)
}

// ExistsIndex 判断索引是否存在
func ExistsIndex(index string) bool {
	exists, err := global.ES.IndexExists(index).Do(context.Background())
	if err != nil {
		logrus.Errorf("检查索引 %s 存在性时出错: %s", index, err)
		return false
	}
	return exists
}

// DeleteIndex 删除索引
func DeleteIndex(index string) {
	_, err := global.ES.
		DeleteIndex(index).
		Do(context.Background())

	if err != nil {
		logrus.Errorf("删除索引失败:%s - %s", index, err)
		return
	}
	logrus.Infof("索引 %s 删除成功", index)
}

// =======================
// 文档操作 (Document Operations)
// =======================

// // DocCreate 添加单个文档
// func DocCreate() {
// 	user := models.UserModel{
// 		ID:        12,
// 		UserName:  "lisi",
// 		Age:       23,
// 		NickName:  "夜空中最亮的lisi",
// 		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
// 		Title:     "今天天气很不错",
// 	}

// 	// 如果 Mapping 中没有定义的字段，ES 默认会自动创建动态映射 (Dynamic Mapping)
// 	indexResponse, err := global.ESClient.
// 		Index().
// 		Index(user.Index()). // 假设 UserModel 有 Index() 方法返回索引名，否则直接写字符串 "user_index"
// 		BodyJson(user).
// 		Do(context.Background())

// 	if err != nil {
// 		fmt.Println("添加文档失败:", err)
// 		return
// 	}

// 	fmt.Printf("文档添加成功: %#v\n", indexResponse)
// }

// // DocCreateBatch 批量添加文档
// func DocCreateBatch() {
// 	list := []models.UserModel{
// 		{
// 			ID:        12,
// 			UserName:  "fengfeng",
// 			NickName:  "夜空中最亮的枫枫",
// 			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
// 		},
// 		{
// 			ID:        13,
// 			UserName:  "lisa",
// 			NickName:  "夜空中最亮的丽萨",
// 			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
// 		},
// 	}

// 	// 初始化 Bulk 请求
// 	// Refresh("true") 表示立即刷新索引，使文档立即可见（测试用，生产环境慎用，影响性能）
// 	bulk := global.ESClient.Bulk().Index(models.UserModel{}.Index()).Refresh("true")

// 	for _, model := range list {
// 		// 使用 NewBulkCreateRequest 还是 NewBulkIndexRequest 取决于你是否需要指定 ID
// 		// 如果模型中有 ID 且希望由该 ID 决定文档唯一性，通常使用 Index 请求覆盖，或 Create 请求（若存在则失败）
// 		// 这里沿用你提供的 Create 逻辑
// 		req := elastic.NewBulkCreateRequest().Doc(model)

// 		// 如果 UserModel 结构体里有 ID 字段且希望作为 ES 的 _id，通常需要显式设置
// 		// req.Id(fmt.Sprintf("%d", model.ID))
// 		// 如果 BodyJson 序列化后包含 id 字段但不作为 _id，则按默认处理

// 		bulk.Add(req)
// 	}

// 	res, err := bulk.Do(context.Background())
// 	if err != nil {
// 		fmt.Println("批量添加失败:", err)
// 		return
// 	}

// 	// 打印执行结果
// 	fmt.Printf("批量操作完成。成功数量: %d, 失败数量: %d\n", len(res.Succeeded()), len(res.Failed()))
// 	if len(res.Failed()) > 0 {
// 		fmt.Println("失败详情:", res.Failed())
// 	}
// }

// // DocDelete 根据 ID 删除单个文档
// func DocDelete(docID string) {
// 	// 注意：实际使用中 docID 应该作为参数传入，这里硬编码是为了演示，建议修改函数签名
// 	// 原代码中是硬编码的，这里改为参数传递更灵活，如果必须硬编码可改回
// 	targetID := docID
// 	if targetID == "" {
// 		targetID = "tmcqfYkBWS69Op6Q4Z0t" // 默认示例 ID
// 	}

// 	deleteResponse, err := global.ESClient.Delete().
// 		Index(models.UserModel{}.Index()).
// 		Id(targetID).
// 		Refresh("true").
// 		Do(context.Background())

// 	if err != nil {
// 		// 如果文档不存在，err 通常包含 404 信息
// 		fmt.Println("删除文档失败 (可能文档不存在):", err)
// 		return
// 	}

// 	fmt.Printf("文档删除成功: %#v\n", deleteResponse)
// }

// // DocDeleteBatch 根据 ID 批量删除文档
// func DocDeleteBatch() {
// 	idList := []string{
// 		"tGcofYkBWS69Op6QHJ2g",
// 		"tWcpfYkBWS69Op6Q050w",
// 	}

// 	bulk := global.ESClient.Bulk().Index(models.UserModel{}.Index()).Refresh("true")

// 	for _, s := range idList {
// 		req := elastic.NewBulkDeleteRequest().Id(s)
// 		bulk.Add(req)
// 	}

// 	res, err := bulk.Do(context.Background())
// 	if err != nil {
// 		fmt.Println("批量删除请求失败:", err)
// 		return
// 	}

// 	// 注意：即使某些文档不存在，bulk.Do() 也不会返回整体错误
// 	// 需要检查 res.Failed() 或 res.Succeeded()
// 	fmt.Printf("批量删除完成。成功删除: %d 个\n", len(res.Succeeded()))
// 	if len(res.Failed()) > 0 {
// 		fmt.Println("删除失败的项目:", res.Failed())
// 	}
// }

// // =======================
// // 文档查询 (Document Query)
// // =======================

// // DocFind 列表查询 (分页 + 全量)
// func DocFind(page, limit int) {
// 	if page < 1 {
// 		page = 1
// 	}
// 	if limit < 1 {
// 		limit = 10
// 	}

// 	from := (page - 1) * limit

// 	// 构建查询条件，这里是空查询 (MatchAll)
// 	query := elastic.NewBoolQuery()

// 	res, err := global.ESClient.
// 		Search(models.UserModel{}.Index()).
// 		Query(query).
// 		From(from).
// 		Size(limit).
// 		Do(context.Background())

// 	if err != nil {
// 		fmt.Println("查询失败:", err)
// 		return
// 	}

// 	// ES 7.x+ TotalHits 是一个对象，需要取 Value
// 	count := res.Hits.TotalHits.Value
// 	fmt.Printf("总记录数: %d, 当前页: %d, 每页条数: %d\n", count, page, limit)

// 	for _, hit := range res.Hits.Hits {
// 		fmt.Println(string(hit.Source))
// 		// 如果需要反序列化到结构体:
// 		// var u models.UserModel
// 		// json.Unmarshal(hit.Source, &u)
// 	}
// }

// // DocFindExact 精确匹配 (针对 keyword 类型字段)
// // 例如：user_name 在 mapping 中定义为 keyword
// func DocFindExact(userName string) {
// 	// TermQuery 不会分词，完全匹配
// 	query := elastic.NewTermQuery("user_name", userName)

// 	res, err := global.ESClient.
// 		Search(models.UserModel{}.Index()).
// 		Query(query).
// 		Do(context.Background())

// 	if err != nil {
// 		fmt.Println("精确查询失败:", err)
// 		return
// 	}

// 	fmt.Printf("精确匹配 '%s' 的结果数: %d\n", userName, res.Hits.TotalHits.Value)
// 	for _, hit := range res.Hits.Hits {
// 		fmt.Println(string(hit.Source))
// 	}
// }

// // DocFindFuzzy 模糊/全文匹配 (针对 text 类型字段，也可用于 keyword 但需完整匹配)
// func DocFindFuzzy(nickName string) {
// 	// MatchQuery 会对输入内容进行分词，然后去匹配倒排索引
// 	// 如果字段是 text 类型：搜 "夜空中" 也能搜出 "夜空中最亮的枫枫"
// 	// 如果字段是 keyword 类型：通常需要输入完整的 "夜空中最亮的枫枫" 才能匹配（除非使用了 ngram 等分词器）
// 	query := elastic.NewMatchQuery("nick_name", nickName)

// 	res, err := global.ESClient.
// 		Search(models.UserModel{}.Index()).
// 		Query(query).
// 		Do(context.Background())

// 	if err != nil {
// 		fmt.Println("模糊查询失败:", err)
// 		return
// 	}

// 	fmt.Printf("模糊匹配 '%s' 的结果数: %d\n", nickName, res.Hits.TotalHits.Value)
// 	for _, hit := range res.Hits.Hits {
// 		fmt.Println(string(hit.Source))
// 	}
// }
