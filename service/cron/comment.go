package cron_service

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SyncComment 同步Redis中的评论增量统计到数据库
// 逻辑流程：拉取ID -> 读取数据 -> 拼接SQL -> 执行更新 -> 确认回写
func SyncComment() {
	batchSize := 500
	total := 0
	batchIndex := 0

	for {
		batchIndex++
		ids := redis_count.PopDirtyCommentIDs(batchSize)
		if len(ids) == 0 {
			break
		}

		diggMap := redis_count.GetAllCacheCommentDigg(ids)

		frashIDs := make([]uint, 0, len(ids))
		diggCases := make([]string, 0, len(ids))

		for _, id := range ids {
			digg := diggMap[id]
			if digg == 0 {
				continue
			}
			frashIDs = append(frashIDs, id)
			diggCases = append(diggCases, fmt.Sprintf("WHEN %d THEN digg_count + %d", id, digg))
		}

		if len(frashIDs) == 0 {
			continue
		}

		diggSQL := "CASE id " + strings.Join(diggCases, " ") + " ELSE digg_count END"

		err := global.DB.Model(&models.CommentModel{}).
			Where("id IN ?", frashIDs).
			Updates(map[string]any{
				"digg_count": gorm.Expr(diggSQL),
			}).Error

		if err != nil {
			logrus.Errorf("评论批量更新失败 [批次 %d, 数量 %d]: %v", batchIndex, len(frashIDs), err)
			redis_count.RequeueDirtyCommentIDs(frashIDs)
			continue
		}

		redis_count.AckCommentSync(frashIDs, diggMap)
		total += len(frashIDs)
		logrus.Infof("评论批量更新成功 [批次 %d, 数量 %d]", batchIndex, len(frashIDs))
	}

	if total == 0 {
		logrus.Info("没有需要同步的评论数据")
		return
	}
	logrus.Infof("评论数据同步完成，本次同步总数：%d", total)
}
