package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/message_service"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (ArticleApi) ArticleRemoveUserView(c *gin.Context) {
	var req models.RemoveRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)

	// 验证ID列表不能为空
	if len(req.IDList) == 0 {
		response.FailWithMsg("ID列表不能为空", c)
		return
	}

	// 检查这些文章是否都属于当前用户
	var existingCount int64
	err = global.DB.Model(&models.ArticleModel{}).
		Where("user_id = ? AND id IN ?", claims.UserID, req.IDList).
		Count(&existingCount).Error

	if err != nil {
		response.FailWithMsg("查询文章失败", c)
		return
	}

	// 如果找到的文章数量与请求删除的数量不一致，说明有些文章不属于当前用户或不存在
	if int(existingCount) != len(req.IDList) {
		response.FailWithMsg("部分文章不存在或不属于当前用户", c)
		return
	}

	// 批量删除
	var list []models.ArticleModel
	global.DB.Find(&list, "id in ?", req.IDList)

	if len(list) > 0 {
		err := global.DB.Delete(&list).Error //也可以把status设置为offline
		if err != nil {
			response.FailWithMsg("删除失败", c)
			return
		}
	}

	response.OkWithMsg(fmt.Sprintf("成功删除 %d 篇文章", len(req.IDList)), c)
}

func (ArticleApi) ArticleRemoveView(c *gin.Context) { //管理员删除
	var req models.RemoveRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	var list []models.ArticleModel
	global.DB.Find(&list, "id in ?", req.IDList)

	if len(list) > 0 {
		err := global.DB.Delete(&list).Error //也可以把status设置为offline
		if err != nil {
			response.FailWithMsg("删除失败", c)
			return
		}
	}
	//查看Redis是否存在该文章缓存
	//redis删除缓存
	delList := make([]string, 0)
	for _, id := range req.IDList {
		delList = append(delList, "ArticleID"+strconv.FormatUint(uint64(id), 10))
	}
	global.RedisHotPool.Del(c, delList...) //由于Redis对于不存在的键,删除操作是无害的,所以这里直接删除

	titleList := []byte{}
	for _, article := range list {
		titleList = append(titleList, []byte(article.Title+" 、")...)
		titleList = append(titleList, []byte("\n")...)
	}
	titleListStr := string(titleList)
	message_service.InsertSystemMessage(models.MessageModel{
		RevUserID:          0,
		ActionUserID:       0,
		ActionUserNickname: "系统",
		Title:              "文章删除通知",
		Content:            fmt.Sprintf("您的文章 %s等已被管理员删除", titleListStr),
	})
	response.OkWithMsg(fmt.Sprintf("删除成功 成功删除%d条", len(list)), c)
}
