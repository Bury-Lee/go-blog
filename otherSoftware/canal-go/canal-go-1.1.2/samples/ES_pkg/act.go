package ES

import (
	"context"
	"fmt"
)

// CreateFromJSON 演示如何直接将现有的 JSON 字符串写入 Elasticsearch
func CreateFromJSON(docID, data string) {
	// 假设这是你已经准备好的 JSON 字符串
	// 注意：在实际代码中，这个字符串可能来自数据库查询结果、消息队列或其他服务

	// 可选：从 JSON 中提取 ID 作为 ES 的文档 _id
	// 如果不提取，ES 会自动生成一个随机 UUID，那么 JSON 里的 "id":2 就只是一个普通字段
	// 这里简单演示如何硬编码或通过解析获取。
	// 生产环境建议使用 json.Unmarshal 解析 data 到一个临时 struct 或 map 来获取 id，以保证健壮性。
	// 为了演示清晰，这里假设我们知道 ID 是 2，或者你手动指定它。

	// 构建并执行请求
	indexResponse, err := ESClient.
		Index().
		Index("article_index").
		// 如果索引不存在，确保 ES 配置了自动创建，或者提前手动创建好映射 (Mapping)

		// 【核心修改点 1】指定文档 ID
		// 这一步非常关键。如果不加这行，ES 会忽略 JSON 里的 "id" 字段，自己生成一个随机 _id。
		// 加上这行，ES 的 _id 就是 "2"，后续可以用这个 ID 进行更新 (Update) 或删除 (Delete)。
		Id(docID).

		// 【核心修改点 2】直接传入 JSON 字符串
		// BodyString 告诉客户端："这是一个已经序列化好的 JSON 字符串，不要再尝试序列化了，直接发过去"
		// 也可以使用 BodyBytes([]byte(data)) 效果一样
		BodyString(data).

		// 执行请求
		Do(context.Background())

	if err != nil {
		// 错误处理
		// 常见错误原因：
		// 1. 索引名称错误且未开启自动创建
		// 2. JSON 格式不合法（比如少了个逗号或引号）
		// 3. 字段类型与索引中已存在的 Mapping 冲突（例如 title 之前定义为 int，现在传了 string）
		fmt.Println("Index failed:", err)
		return
	}

	// 打印成功响应
	// 检查 indexResponse.Result，如果是 "created" 表示新建，"updated" 表示覆盖了旧数据
	fmt.Printf("Index success: %#v\n", indexResponse)

	// 额外提示：你可以检查响应中的 ID 是否与你预期的一致
	fmt.Printf("Document ID in ES: %s\n", indexResponse.Id)
}

// DocDeleteByID 根据传入的 documentID 删除指定文档
// 参数: docID - 要删除的文档唯一标识符
func DeleteByID(docID string) {
	// 简单的校验，防止传入空字符串
	if docID == "" {
		fmt.Println("错误：文档 ID 不能为空")
		return
	}

	deleteResponse, err := ESClient.Delete().
		Index("article_index"). // 指定索引
		Id(docID).              // 【关键】使用传入的变量作为 ID
		Refresh("wait_for").    // 强制刷新（测试用，生产环境可去掉或设为 "wait_for"）
		Do(context.Background())

	if err != nil {
		// 注意：如果文档不存在，某些版本的客户端或配置可能会返回 404 错误
		// 你可以根据需要判断是否是 "not_found" 错误
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	// 检查响应结果，确认是否真的删除了
	// deleteResponse.Result 的值可能是 "deleted" (成功删除) 或 "not_found" (文档原本就不存在)
	fmt.Printf("删除操作完成。结果状态: %s\n", deleteResponse.Result)

	// 如果需要更详细的判断：
	if deleteResponse.Result == "not_found" {
		fmt.Println("提示：该文档不存在，无需删除。")
	} else if deleteResponse.Result == "deleted" {
		fmt.Println("成功：文档已删除。")
	}
}
