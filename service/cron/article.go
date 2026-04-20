package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 异步双写实现计数同步
// SyncArticle 同步Redis中的文章增量统计到数据库
// 逻辑流程：拉取ID -> 读取数据 -> 拼接SQL -> 执行更新 -> 确认回写
// 此处一次更新多个字段,对内存有更高的要求,如果内存不足可以牺牲性能,分多几批更新,一次弹出一个字段的更新
func SyncArticle() {
	//检查是否有锁,如果有则跳过本次启动
	batchSize := 500
	total := 0
	batchIndex := 0

	// 循环拉取批次
	for {
		//每次写入数据库前，先获取锁，确保只有一个节点在写入
		ctx := context.Background()
		Lock := global.RedisTimeCache.Get(ctx, "cron_lock").Val()
		if Lock != global.Config.System.Addr() {
			logrus.Info("定时任务锁已被占用，跳过本次任务")
			break
		}
		batchIndex++
		// 逻辑点A: 从Redis弹出一批脏ID
		ids := redis_count.PopDirtyArticleIDs(batchSize)
		if len(ids) == 0 {
			break
		}

		// 根据ID批量获取具体的计数增量
		// 逻辑点B: 这里假设Redis中存储的是每个ID的增量值
		collectMap := redis_count.GetAllCacheCollect(ids)
		diggMap := redis_count.GetAllCacheDigg(ids)
		lookMap := redis_count.GetAllCacheLook(ids)
		comMap := redis_count.GetAllCacheComment(ids)

		// 准备SQL构建数据
		flushIDs := make([]uint, 0, len(ids))
		lookCases := make([]string, 0, len(ids))
		diggCases := make([]string, 0, len(ids))
		collCases := make([]string, 0, len(ids))
		comCases := make([]string, 0, len(ids))

		// 数据清洗与SQL片段构建
		for _, id := range ids {
			look := lookMap[id]
			digg := diggMap[id]
			coll := collectMap[id]
			com := comMap[id]
			if look == 0 && digg == 0 && coll == 0 && com == 0 {
				continue
			}
			flushIDs = append(flushIDs, id)
			// 构建CASE WHEN SQL，实现单条SQL批量更新不同值
			lookCases = append(lookCases, fmt.Sprintf("WHEN %d THEN look_count + %d", id, look))
			diggCases = append(diggCases, fmt.Sprintf("WHEN %d THEN digg_count + %d", id, digg))
			collCases = append(collCases, fmt.Sprintf("WHEN %d THEN collect_count + %d", id, coll))
			comCases = append(comCases, fmt.Sprintf("WHEN %d THEN comment_count + %d", id, com))
		}

		if len(flushIDs) == 0 {
			continue
		}

		// 5. 拼接最终SQL语句
		diggSQL := "CASE id " + strings.Join(diggCases, " ") + " ELSE digg_count END"
		lookSQL := "CASE id " + strings.Join(lookCases, " ") + " ELSE look_count END"
		collSQL := "CASE id " + strings.Join(collCases, " ") + " ELSE collect_count END"
		comSQL := "CASE id " + strings.Join(comCases, " ") + " ELSE comment_count END"

		// 6. 执行数据库更新
		err := global.DB.Model(&models.ArticleModel{}).
			Where("id IN ?", flushIDs).
			Updates(map[string]any{
				"digg_count":    gorm.Expr(diggSQL),
				"look_count":    gorm.Expr(lookSQL),
				"collect_count": gorm.Expr(collSQL),
				"comment_count": gorm.Expr(comSQL),
			}).Error

		if err != nil {
			logrus.Errorf("批量更新失败 [批次 %d, 数量 %d]: %v", batchIndex, len(flushIDs), err)
			// 逻辑点C: 将失败的ID重新放回Redis
			redis_count.RequeueDirtyArticleIDs(flushIDs)
			continue
		}

		// 数据库更新成功，清理Redis缓存
		redis_count.AckArticleSync(flushIDs, lookMap, diggMap, collectMap, comMap)
		total += len(flushIDs)
		logrus.Infof("批量更新成功 [批次 %d, 数量 %d]", batchIndex, len(flushIDs))
	}

	if total == 0 {
		logrus.Info("没有需要同步的文章数据")
		return
	}
	logrus.Infof("文章数据同步完成，本次同步总数：%d", total)
}
