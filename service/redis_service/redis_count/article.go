package redis_count

import (
	"StarDreamerCyberNook/global"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func GetCacheLook(articleID uint) int {
	return getDirtArticle(articleCacheLook, articleID)
}

func GetCacheDigg(articleID uint) int {
	return getDirtArticle(articleCacheDigg, articleID)
}

func GetCacheCollect(articleID uint) int {
	return getDirtArticle(articleCacheCollect, articleID)
}

func GetCacheComment(articleID uint) int { //获取当前文章的评论增量
	return getDirtArticle(articleCacheComment, articleID)
}

// SetCacheLook 设置文章浏览增量+1
func SetCacheLook(articleID uint, increase bool) {
	setDirtArticle(articleCacheLook, articleID, increase)
}

// SetCacheDigg 设置文章点赞增量+1
func SetCacheDigg(articleID uint, increase bool) {
	setDirtArticle(articleCacheDigg, articleID, increase)
}

// SetCacheCollect 设置文章收藏增量+1
func SetCacheCollect(articleID uint, increase bool) {
	setDirtArticle(articleCacheCollect, articleID, increase)
}

// SetCacheComment 设置文章评论增量+1
func SetCacheComment(articleID uint, increase bool) {
	setDirtArticle(articleCacheComment, articleID, increase)
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
	if err := global.RedisTimeCache.SAdd(ctx, DirtyArticleSetKey, fields...).Err(); err != nil {
		logrus.Errorf("回退脏文章ID失败, count: %d, err: %v", len(ids), err)
	}
}

func AckArticleSync(ids []uint, lookMap, diggMap, collectMap, commentMap map[uint]int) {
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
		if comment := commentMap[id]; comment != 0 {
			pipe.HIncrBy(ctx, string(articleCacheComment), field, int64(-comment))
		}
		// 同时从脏数据集合中移除ID
	}
	if _, err := pipe.Exec(ctx); err != nil {
		logrus.Errorf("确认文章同步失败, count: %d, err: %v", len(ids), err)
		// Ack失败时回补脏ID,避免出现哈希有增量但集合无ID的悬挂状态
		RequeueDirtyArticleIDs(ids)
	}
}

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

// ClearArticleCache 清空文章统计相关缓存
// 说明:删除浏览点赞收藏哈希和脏文章集合
func ClearArticleCache() {
	ctx := context.Background()
	err := global.RedisTimeCache.Del(ctx, string(articleCacheLook), string(articleCacheDigg), string(articleCacheCollect), DirtyArticleSetKey).Err()
	if err != nil {
		logrus.Error(err)
	}
}

// PopDirtyArticleIDs 从Redis集合中批量弹出脏文章ID
// 参数limit表示最多弹出的ID数量
// 返回值为弹出的ID切片
func PopDirtyArticleIDs(limit int) (ids []uint) {
	// 创建上下文对象，用于控制请求的生命周期
	ctx := context.Background()

	// 如果限制数量小于等于0，直接返回空切片
	if limit <= 0 {
		return
	}

	// 初始化结果切片，预分配容量以提高性能
	ids = make([]uint, 0, limit)

	// 循环弹出指定数量的ID或直到集合为空
	for i := 0; i < limit; i++ {
		// 从Redis集合(DirtyArticleSetKey)中随机弹出一个元素(文章ID)
		// SPop命令会随机移除并返回集合中的一个元素
		idStr, err := global.RedisTimeCache.SPop(ctx, DirtyArticleSetKey).Result()
		if err != nil {
			// 如果错误是redis.Nil，说明集合已空，跳出循环
			if err == redis.Nil {
				break
			}
			// 记录其他错误日志
			logrus.Errorf("弹出脏文章ID失败, err: %v", err)
			break
		}

		// 将字符串类型的ID转换为整型
		id, err := strconv.Atoi(idStr)
		if err != nil {
			// 如果转换失败，跳过此次循环继续下一次
			continue
		}

		// 将转换后的ID添加到结果切片中
		ids = append(ids, uint(id))
	}

	// 返回弹出的ID切片
	return
}

func GetDirtyArticleIDs() (ids []uint) { //获取"脏文章ID"集合
	context := context.Background()
	res, err := global.RedisTimeCache.SMembers(context, DirtyArticleSetKey).Result()
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

// 批量获取增量数据,返回map[文章ID]增量值
func GetAllCacheLook(articleIDs []uint) (mps map[uint]int) {
	return GetAllDirtArticleID(articleCacheLook, articleIDs)
}

func GetAllCacheDigg(articleIDs []uint) (mps map[uint]int) {
	return GetAllDirtArticleID(articleCacheDigg, articleIDs)
}

func GetAllCacheCollect(articleIDs []uint) (mps map[uint]int) {
	return GetAllDirtArticleID(articleCacheCollect, articleIDs)
}

func GetAllCacheComment(articleIDs []uint) (mps map[uint]int) {
	return GetAllDirtArticleID(articleCacheComment, articleIDs)
}

// setDirtArticle 更新计数缓存并标记脏文章
// 参数:t - 缓存类型键
// 参数:articleID - 文章ID
// 参数:increase - true表示递增,false表示递减
// 说明:先读取当前值后增减,再回写哈希并把ID加入脏集合
func setDirtArticle(t CacheType, articleID uint, increase bool) {
	field := strconv.Itoa(int(articleID))
	var delta int64
	context := context.Background()

	global.RedisTimeCache.HIncrBy(context, string(t), field, delta) //先进行一次0增加操作,避免初始化为0导致数值增减失败
	if increase {
		delta = 1
	} else {
		delta = -1
	}
	// 先执行 HIncrBy 操作
	global.RedisTimeCache.HIncrBy(context, string(t), field, delta)
	// 设置该哈希键的过期时间
	//这里要和cron的同步时间要一致,比其多一些
	global.RedisTimeCache.Expire(context, string(t), cacheTTL)

	// 添加到脏数据集合
	global.RedisTimeCache.SAdd(context, DirtyArticleSetKey, field)

	// 同时给脏数据集合也设置过期时间
	global.RedisTimeCache.Expire(context, DirtyArticleSetKey, cacheTTL)
}

// func setDirtArticle(t CacheType, articleID uint, increase bool) {//这个是一次增加但是每次第一次写入时增量变成实际增量-1的版本,因为初始化问题
// 	field := strconv.Itoa(int(articleID))
// 	var delta int64 = 1
// 	if !increase {
// 		delta = -1
// 	}
// 	context := context.Background()

// 	// 先执行 HIncrBy 操作
// 	global.RedisTimeCache.HIncrBy(context, string(t), field, delta)

// 	// 设置该哈希键的过期时间
// 	//这里要和cron的同步时间要一致,比其多一些
// 	global.RedisTimeCache.Expire(context, string(t), cacheTTL)

// 	// 添加到脏数据集合
// 	global.RedisTimeCache.SAdd(context, DirtyArticleSetKey, field)

//		// 同时给脏数据集合也设置过期时间
//		global.RedisTimeCache.Expire(context, DirtyArticleSetKey, cacheTTL)
//	}

func getDirtArticle(t CacheType, ID uint) int { //获取单个文章的增量值
	context := context.Background()
	num, err := global.RedisTimeCache.HGet(context, string(t), strconv.Itoa(int(ID))).Int()
	if err != nil && err != redis.Nil {
		logrus.Errorf("获取缓存计数失败, err: %v", err)
	}
	return num
}

func GetAllDirtArticleID(t CacheType, articleIDs []uint) (mps map[uint]int) {
	/*
		创建一个map来存储结果（文章ID到数量的映射）
		如果输入的文章ID列表为空，则直接返回空map
		将文章ID转换为字符串格式，准备用于Redis查询
		使用Redis的HMGET命令批量从哈希表中获取多个字段的值
		遍历返回的结果，将非空且有效的值转换为整数，存入结果map中
		返回包含文章ID及其对应数量的map
	*/
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
