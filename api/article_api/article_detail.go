package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	"context"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleDetailResponse struct {
	models.ArticleModel
	UserName   string `json:"username"`
	NickName   string `json:"nickname"`
	UserAvatar string `json:"userAvatar"`
}

func (ArticleApi) ArticleDetailView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	//redis缓存
	idStr := strconv.FormatUint(uint64(req.ID), 10)
	ctx := context.Background()
	res, exit := global.RedisHotPool.Get(ctx, "ArticleID"+idStr).Result()
	if exit == nil {
		// 缓存命中
		var cached ArticleDetailResponse
		err := json.Unmarshal([]byte(res), &cached)
		if err != nil {
			response.FailWithMsg("缓存数据解析错误", c)
			return
		}
		//权限检查
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil {
			if cached.Status != models.StatusPublished {
				response.FailWithMsg("文章不存在", c)
				return
			}
		} else if claims.Role == enum.UserRole && claims.UserID != cached.UserID {
			if cached.Status != models.StatusPublished {
				response.FailWithMsg("文章不存在", c)
				return
			}
		}

		// 计数只在响应阶段叠加,不写回详情缓存
		result := cached
		collectCount := redis_count.GetCacheCollect(result.ID)
		lookCount := redis_count.GetCacheLook(result.ID)
		diggCount := redis_count.GetCacheDigg(result.ID)
		result.CollectCount += collectCount
		result.LookCount += lookCount
		result.DiggCount += diggCount

		response.OkWithData(result, c)
		return
	}

	// 未登录的用户，只能看到发布成功的文章
	// 登录用户，能看到自己的所有文章
	// 管理员，能看到全部的文章

	var article models.ArticleModel
	err := global.DB.Preload("UserModel").Take(&article, req.ID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		if article.Status != models.StatusPublished {
			response.FailWithMsg("文章不存在", c)
			return
		}
	} else if claims.Role == enum.UserRole && claims.UserID != article.UserID {
		if article.Status != models.StatusPublished {
			response.FailWithMsg("文章不存在", c)
			return
		}
	}

	cached := ArticleDetailResponse{
		ArticleModel: article,
		UserName:     article.UserModel.UserName,
		NickName:     article.UserModel.NickName,
		UserAvatar:   article.UserModel.Avatar,
	}

	// 计数只在响应阶段叠加,不写回详情缓存
	result := cached
	collectCount := redis_count.GetCacheCollect(result.ID)
	lookCount := redis_count.GetCacheLook(result.ID)
	diggCount := redis_count.GetCacheDigg(result.ID)
	result.CollectCount += collectCount
	result.LookCount += lookCount
	result.DiggCount += diggCount

	response.OkWithData(result, c)

	//把数据加入缓存
	// logrus.Debug("缓存文章详情", idStr, idStr, cached) //debug
	jsonData, err := json.Marshal(cached)
	if err != nil {
		logrus.Error("缓存数据序列化错误:", err)
		return
	}
	global.RedisHotPool.Set(ctx, "ArticleID"+idStr, string(jsonData), 0)
}
