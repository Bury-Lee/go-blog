// Licensed to the Apache Software Foundation (ASF) under one
// ... (许可证头省略，保持原样) ...

package main

import (
	"fmt"
	"log"
	"time"

	// 引入 canal-go 客户端库，用于连接 Canal Server
	"github.com/withlin/canal-go/client"
	ES "github.com/withlin/canal-go/samples/ES_pkg"
	// 注意：实际处理 Entry 数据时通常还需要引入 protocol 相关包来解析具体字段
	// "github.com/withlin/canal-go/protocol"
)

func main() {
	if err := ES.ConnectES(); err != nil {
		log.Fatalf("连接 Elasticsearch 失败: %v", err)
	}

	// ------------------------------------------------------------------
	// 1. 初始化并建立与 Canal Server 的连接
	// ------------------------------------------------------------------
	// 参数说明:
	// - "127.0.0.1", 11111: Canal Server 的地址和端口
	// - "", "": 用户名和密码 (如果未开启鉴权则留空)
	// - "example": 订阅的 Canal instance 名称 (需与 server 端配置一致)
	// - 60000: 连接超时时间 (毫秒)
	// - 60*60*1000: 空闲超时时间 (毫秒)，用于长连接心跳检测
	connector := client.NewSimpleCanalConnector("127.0.0.1", 11111, "", "", "example", 60000, 60*60*1000)

	// 发起网络连接
	err := connector.Connect()
	if err != nil {
		log.Fatalf("连接 Canal Server 失败: %v", err)
	}

	// ------------------------------------------------------------------
	// 2. 订阅特定的数据库表变更
	// ------------------------------------------------------------------
	// 使用正则表达式过滤需要监听的表
	// 示例: ".*\\.article_models" 表示监听所有库下的 article_models 表
	// 若要监听整个库，可使用 ".*\\..*"
	err = connector.Subscribe(".*\\.article_models")
	if err != nil {
		log.Fatalf("订阅表结构失败: %v", err)
	}

	// 提示：生产环境中，请在此处初始化 Elasticsearch、Kafka 或其他下游客户端
	// esClient := initESClient()

	fmt.Println("=== 开始监听数据库变更并执行数据转换 ===")

	// ------------------------------------------------------------------
	// 3. 主循环：拉取消息、处理业务逻辑、确认消费
	// ------------------------------------------------------------------
	for {
		// 获取批次消息
		// 参数: batchSize=100, timeout=nil, unit=nil (使用默认超时)
		message, err := connector.Get(100, nil, nil)
		if err != nil {
			log.Printf("获取消息批次出错: %v", err)
			// 发生网络异常时短暂休眠，避免频繁重试导致资源耗尽
			time.Sleep(1 * time.Second)
			continue
		}

		batchId := message.Id

		// 如果没有拉取到有效数据 (batchId==-1 或 条目为空)，进入短时休眠等待
		if batchId == -1 || len(message.Entries) <= 0 {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		// 【核心业务】处理并转换当前批次的 Entry 数据
		// 建议在此函数内部实现具体的 ETL 逻辑或发送到消息队列
		ES.ProcessAndConvertEntry(message.Entries)

		// ----------------------------------------------------------------
		// 重要：确认消费 (Ack)
		// ----------------------------------------------------------------
		// 只有当数据被成功处理（如写入 ES 成功）后，才应调用 Ack。
		// 如果处理失败，不应调用 Ack，以便下次重新拉取该批次数据（需配合 Canal 服务端配置）。
		// 当前示例已注释，调试通过后请务必取消注释，否则会导致数据重复消费或无法推进位点。
		// ----------------------------------------------------------------
		// if err := connector.Ack(batchId); err != nil {
		//     log.Printf("确认消息失败: %v", err)
		// }
	}
}
