package redis_count

import "time"

// 业界主流方案（建议版）

// - 方案 A（主流且落地快）： Redis 计数 + MQ/Stream 异步落库 + 对账
//   - 写入：点赞/收藏/浏览先做关系表/幂等校验，再原子 INCRBY 计数，同时投递事件到 Kafka/Redis Stream 。
//   - 落库：消费者按分区顺序消费，批量 UPDATE article SET count = count + ? ，失败重试+死信。
//   - 读取：详情静态信息单独缓存；计数独立读取 Redis（或短 TTL 本地缓存）。
//   - 保障：每日离线对账（关系表反算 vs article 计数），自动修正漂移。
// - 方案 B（高并发进阶）： 事件溯源 + 近实时聚合
//   - 所有行为事件只入日志流，不直接改 article 计数。
//   - 流处理（Flink/Kafka Streams）实时聚合写回 Redis/ClickHouse，MySQL 定时固化快照。
//   - 适合超大规模互动场景，复杂度更高。

type CacheType string

// articleCacheType Redis中文章统计缓存的哈希键类型
const (
	articleCacheLook    CacheType = "article_look_key"
	articleCacheDigg    CacheType = "article_digg_key"
	articleCacheCollect CacheType = "article_collect_key"
	articleCacheComment CacheType = "article_comment_key"

	DirtyArticleSetKey string = "cache_dirty_article_ids" //脏文章ID集合键

	commontCacheDigg CacheType = "comment_digg_key"

	DirtyCommentSetKey string = "cache_dirty_comment_ids" //脏评论ID集合键
)

var cacheTTL = time.Minute * 25 //统一过期时间
