package redis_count

import (
	"StarDreamerCyberNook/global"
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func SetCacheCommentDigg(commentID uint, increase bool) {
	setDirtComment(commontCacheDigg, commentID, increase)
}

// RequeueDirtyCommentIDs 将一批脏评论ID重新加入到Redis集合中
// 这个函数通常用于将之前取出但未处理成功的脏评论ID重新放回待处理队列
// 参数ids是一个包含需要重新入队的脏评论ID的切片
func RequeueDirtyCommentIDs(ids []uint) {
	// 检查输入参数，如果ID列表为空则直接返回
	if len(ids) == 0 {
		return
	}

	// 创建一个空接口切片，用于存储转换后的ID字符串
	// 容量设置为与输入ID数量相同以优化内存分配
	fields := make([]any, 0, len(ids))

	// 遍历所有输入的ID，将其转换为字符串格式
	for _, id := range ids {
		// 将uint类型的ID转换为int再转为字符串格式
		fields = append(fields, strconv.Itoa(int(id)))
	}

	// 创建上下文对象用于Redis操作
	ctx := context.Background()

	// 使用SAdd命令将所有ID批量添加到Redis集合DirtyCommentSetKey中
	if err := global.RedisTimeCache.SAdd(ctx, DirtyCommentSetKey, fields...).Err(); err != nil {
		// 如果添加操作失败，记录错误日志，包括失败的数量和具体错误信息
		logrus.Errorf("回退脏评论ID失败, count: %d, err: %v", len(ids), err)
	}
}

// AckCommentSync 确认评论同步操作，更新评论点赞数缓存并清理已处理的脏评论ID
// 参数ids为需要确认同步的评论ID列表
// 参数diggMap为评论ID与其对应点赞数变化量的映射表
func AckCommentSync(ids []uint, diggMap map[uint]int) {
	// 如果没有需要处理的ID，则直接返回
	if len(ids) == 0 {
		return
	}

	// 创建上下文对象用于Redis操作
	ctx := context.Background()

	// 创建Redis事务管道，用于批量执行多个Redis命令
	pipe := global.RedisTimeCache.TxPipeline()

	// 遍历所有需要确认的评论ID
	for _, id := range ids {
		// 将评论ID转换为字符串格式作为Redis哈希表的字段名
		field := strconv.Itoa(int(id))

		// 从diggMap中获取该评论的点赞数变化量
		if digg := diggMap[id]; digg != 0 {
			// 在Redis哈希表commontCacheDigg中减少对应的点赞数
			// 使用负值是因为这里是要撤销之前增加的点赞数
			pipe.HIncrBy(ctx, string(commontCacheDigg), field, int64(-digg))
		}
	}

	// 执行事务管道中的所有命令
	if _, err := pipe.Exec(ctx); err != nil {
		// 如果执行失败，记录错误日志
		logrus.Errorf("确认评论同步失败, count: %d, err: %v", len(ids), err)
		// 将这些ID重新加入到脏评论队列中，以便后续重试
		RequeueDirtyCommentIDs(ids)
	}
}

// PopDirtyCommentIDs 从Redis集合中弹出指定数量的脏评论ID
// 参数limit表示要弹出的ID数量上限
// 返回一个包含脏评论ID的切片
func PopDirtyCommentIDs(limit int) (ids []uint) {
	ctx := context.Background()
	// 如果限制数量小于等于0，则直接返回空切片
	if limit <= 0 {
		return
	}
	// 初始化结果切片，容量设为limit以提高性能
	ids = make([]uint, 0, limit)

	// 循环尝试弹出limit个元素
	for i := 0; i < limit; i++ {
		// 从Redis集合DirtyCommentSetKey中随机弹出一个元素（SPop命令）
		idStr, err := global.RedisTimeCache.SPop(ctx, DirtyCommentSetKey).Result()
		if err != nil {
			// 如果Redis集合为空（redis.Nil错误），则跳出循环
			if err == redis.Nil {
				break
			}
			// 记录其他类型的错误
			logrus.Errorf("弹出脏评论ID失败, err: %v", err)
			break
		}
		// 将字符串形式的ID转换为整数
		id, err := strconv.Atoi(idStr)
		if err != nil {
			// 如果转换失败，跳过这个ID，继续处理下一个
			continue
		}
		// 将转换后的ID添加到结果切片中
		ids = append(ids, uint(id))
	}
	return
}

func GetDirtyCommentIDs() (ids []uint) {
	ctx := context.Background()
	res, err := global.RedisTimeCache.SMembers(ctx, DirtyCommentSetKey).Result()
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

func GetAllCacheCommentDigg(commentIDs []uint) (mps map[uint]int) {
	return GetAllDirtArticleID(commontCacheDigg, commentIDs)
}

func setDirtComment(t CacheType, commentID uint, increase bool) {
	// 将评论ID转换为字符串作为Redis哈希表的字段名
	field := strconv.Itoa(int(commentID))

	// 设置增量，默认为1，如果increase为false则设为-1
	var delta int64 = 1
	if !increase {
		delta = -1
	}

	// 创建上下文对象
	ctx := context.Background()

	// 在Redis哈希表中对指定字段增加delta值
	global.RedisTimeCache.HIncrBy(ctx, string(t), field, delta)

	// 设置哈希表的过期时间
	global.RedisTimeCache.Expire(ctx, string(t), cacheTTL)

	// 将评论ID添加到脏数据集合中
	// 标记需要同步到数据库的数据
	global.RedisTimeCache.SAdd(ctx, DirtyCommentSetKey, field)

	// 设置脏数据集合的过期时间
	global.RedisTimeCache.Expire(ctx, DirtyCommentSetKey, cacheTTL)
}
