package redis_article

import (
	"StarDreamerCyberNook/global"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

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

type articleCacheType string

// articleCacheType Redis中文章统计缓存的哈希键类型
const (
	articleCacheLook    articleCacheType = "article_look_key"
	articleCacheDigg    articleCacheType = "article_digg_key"
	articleCacheCollect articleCacheType = "article_collect_key"
	articleDirtySetKey                   = "article_cache_dirty_ids"
)

// set 更新文章计数缓存并标记脏文章
// 参数:t - 缓存类型键
// 参数:articleID - 文章ID
// 参数:increase - true表示递增,false表示递减
// 说明:先读取当前值后增减,再回写哈希并把文章ID加入脏集合
func set(t articleCacheType, articleID uint, increase bool) {
	field := strconv.Itoa(int(articleID))
	var delta int64 = 1
	if !increase {
		delta = -1
	}
	context := context.Background()

	// 先执行 HIncrBy 操作
	global.RedisTimeCache.HIncrBy(context, string(t), field, delta)

	// 设置该哈希键的过期时间
	ttl := time.Minute * 20 //这里要和cron的同步时间要一致,比其多一些
	global.RedisTimeCache.Expire(context, string(t), ttl)

	// 添加到脏数据集合
	global.RedisTimeCache.SAdd(context, articleDirtySetKey, field)

	// 同时给脏数据集合也设置过期时间
	global.RedisTimeCache.Expire(context, articleDirtySetKey, ttl)
}

// SetCacheLook 更新文章浏览数缓存
// 参数:articleID - 文章ID
// 参数:increase - true表示递增,false表示递减
// 说明:复用通用set逻辑处理浏览计数
func SetCacheLook(articleID uint, increase bool) {
	set(articleCacheLook, articleID, increase)
}

// SetCacheDigg 更新文章点赞数缓存
// 参数:articleID - 文章ID
// 参数:increase - true表示递增,false表示递减
// 说明:复用通用set逻辑处理点赞计数
func SetCacheDigg(articleID uint, increase bool) {
	set(articleCacheDigg, articleID, increase)
}

// SetCacheCollect 更新文章收藏数缓存
// 参数:articleID - 文章ID
// 参数:increase - true表示递增,false表示递减
// 说明:复用通用set逻辑处理收藏计数
func SetCacheCollect(articleID uint, increase bool) {
	set(articleCacheCollect, articleID, increase)
}

// get 读取单篇文章的指定统计缓存值
// 参数:t - 缓存类型键
// 参数:articleID - 文章ID
// 返回:num - 当前缓存计数
// 说明:读取失败时按原逻辑返回0
func get(t articleCacheType, articleID uint) int {
	context := context.Background()
	num, _ := global.RedisTimeCache.HGet(context, string(t), strconv.Itoa(int(articleID))).Int()
	return num
}

// GetCacheLook 获取文章浏览数缓存
// 参数:articleID - 文章ID
// 返回:num - 浏览数
// 说明:包装通用get读取浏览计数
func GetCacheLook(articleID uint) int {
	return get(articleCacheLook, articleID)
}

// GetCacheDigg 获取文章点赞数缓存
// 参数:articleID - 文章ID
// 返回:num - 点赞数
// 说明:包装通用get读取点赞计数
func GetCacheDigg(articleID uint) int {
	return get(articleCacheDigg, articleID)
}

// GetCacheCollect 获取文章收藏数缓存
// 参数:articleID - 文章ID
// 返回:num - 收藏数
// 说明:包装通用get读取收藏计数
func GetCacheCollect(articleID uint) int {
	return get(articleCacheCollect, articleID)
}

// GetAll 批量获取文章指定统计缓存值
// 参数:t - 缓存类型键
// 参数:articleIDs - 文章ID列表
// 返回:mps - 文章ID到缓存值的映射
// 说明:使用HMGet批量读取,忽略空值和解析失败项
func GetAll(t articleCacheType, articleIDs []uint) (mps map[uint]int) {
	mps = make(map[uint]int, len(articleIDs))
	if len(articleIDs) == 0 {
		return mps
	}

	fields := make([]string, 0, len(articleIDs))
	for _, id := range articleIDs {
		fields = append(fields, strconv.Itoa(int(id)))
	}

	context := context.Background()
	res, err := global.RedisTimeCache.HMGet(context, string(t), fields...).Result()
	if err != nil {
		return mps
	}

	for idx, val := range res {
		if val == nil {
			continue
		}
		num, err := strconv.Atoi(fmt.Sprintf("%v", val))
		if err != nil {
			continue
		}
		mps[articleIDs[idx]] = num
	}
	return mps
}

// GetAllCacheLook 批量获取文章浏览数缓存
// 参数:articleIDs - 文章ID列表
// 返回:mps - 文章ID到浏览数的映射
// 说明:包装GetAll读取浏览计数
func GetAllCacheLook(articleIDs []uint) (mps map[uint]int) {
	return GetAll(articleCacheLook, articleIDs)
}

// GetAllCacheDigg 批量获取文章点赞数缓存
// 参数:articleIDs - 文章ID列表
// 返回:mps - 文章ID到点赞数的映射
// 说明:包装GetAll读取点赞计数
func GetAllCacheDigg(articleIDs []uint) (mps map[uint]int) {
	return GetAll(articleCacheDigg, articleIDs)
}

// GetAllCacheCollect 批量获取文章收藏数缓存
// 参数:articleIDs - 文章ID列表
// 返回:mps - 文章ID到收藏数的映射
// 说明:包装GetAll读取收藏计数
func GetAllCacheCollect(articleIDs []uint) (mps map[uint]int) {
	return GetAll(articleCacheCollect, articleIDs)
}

// GetDirtyArticleIDs 获取所有脏文章ID
// 返回:ids - 需要回写数据库的文章ID列表
// 说明:从脏集合读取字符串ID并转换为uint
func GetDirtyArticleIDs() (ids []uint) {
	context := context.Background()
	res, err := global.RedisTimeCache.SMembers(context, articleDirtySetKey).Result()
	if err != nil {
		return
	}
	ids = make([]uint, 0, len(res))
	for _, idStr := range res {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, uint(id))
	}
	return
}

func PopDirtyArticleIDs(limit int) (ids []uint) {
	ctx := context.Background()
	if limit <= 0 {
		return
	}
	ids = make([]uint, 0, limit)
	for i := 0; i < limit; i++ {
		idStr, err := global.RedisTimeCache.SPop(ctx, articleDirtySetKey).Result()
		if err != nil {
			if err == redis.Nil {
				break
			}
			logrus.Errorf("弹出脏文章ID失败, err: %v", err)
			break
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, uint(id))
	}
	return
}

func RequeueDirtyArticleIDs(ids []uint) {
	if len(ids) == 0 {
		return
	}
	fields := make([]any, 0, len(ids))
	for _, id := range ids {
		fields = append(fields, strconv.Itoa(int(id)))
	}
	ctx := context.Background()
	if err := global.RedisTimeCache.SAdd(ctx, articleDirtySetKey, fields...).Err(); err != nil {
		logrus.Errorf("回退脏文章ID失败, count: %d, err: %v", len(ids), err)
	}
}

func AckArticleSync(ids []uint, lookMap, diggMap, collectMap map[uint]int) {
	if len(ids) == 0 {
		return
	}
	ctx := context.Background()
	pipe := global.RedisTimeCache.TxPipeline()
	for _, id := range ids {
		field := strconv.Itoa(int(id))
		if look := lookMap[id]; look != 0 {
			pipe.HIncrBy(ctx, string(articleCacheLook), field, int64(-look))
		}
		if digg := diggMap[id]; digg != 0 {
			pipe.HIncrBy(ctx, string(articleCacheDigg), field, int64(-digg))
		}
		if collect := collectMap[id]; collect != 0 {
			pipe.HIncrBy(ctx, string(articleCacheCollect), field, int64(-collect))
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		logrus.Errorf("确认文章同步失败, count: %d, err: %v", len(ids), err)
		// Ack失败时回补脏ID,避免出现哈希有增量但集合无ID的悬挂状态
		RequeueDirtyArticleIDs(ids)
	}
}

// SetUserArticleHistoryCache 写入用户文章阅读历史缓存
// 参数:articleID - 文章ID
// 参数:userID - 用户ID
// 说明:使用用户维度哈希记录文章字段,并设置近24小时过期
func SetUserArticleHistoryCache(articleID, userID uint) {
	key := fmt.Sprintf("history_%d", userID)
	field := fmt.Sprintf("%d", articleID)

	endTime := time.Now().Local().Add(time.Hour * 23)
	ctx := context.Background()
	err := global.RedisTimeCache.HSet(ctx, key, field, "").Err()
	if err != nil {
		logrus.Errorf("设置阅读历史缓存失败, userID: %d, articleID: %d, key: %s, error: %v", userID, articleID, key, err)
		return
	}
	err = global.RedisTimeCache.ExpireAt(ctx, key, endTime).Err()
	if err != nil {
		logrus.Errorf("设置阅读历史缓存过期时间失败, userID: %d, articleID: %d, key: %s, error: %v", userID, articleID, key, err)
		return
	}
}

// GetUserArticleHistoryCache 判断用户是否已读指定文章
// 参数:articleID - 文章ID
// 参数:userID - 用户ID
// 返回:ok - true表示缓存命中,false表示未命中
// 说明:通过HGet是否报错判断阅读记录是否存在
func GetUserArticleHistoryCache(articleID, userID uint) bool {
	key := fmt.Sprintf("history_%d", userID)
	field := fmt.Sprintf("%d", articleID)
	ctx := context.Background()
	err := global.RedisTimeCache.HGet(ctx, key, field).Err()
	if err != nil {
		return false
	}
	return true
}

// Clear 清空文章统计相关缓存
// 说明:删除浏览点赞收藏哈希和脏文章集合
func Clear() {
	ctx := context.Background()
	err := global.RedisTimeCache.Del(ctx, string(articleCacheLook), string(articleCacheDigg), string(articleCacheCollect), articleDirtySetKey).Err()
	if err != nil {
		logrus.Error(err)
	}
}
