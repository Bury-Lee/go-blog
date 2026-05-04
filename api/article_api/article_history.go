package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/redis_service/redis_count"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticleLookRequest struct {
	ArticleID  uint `json:"articleID" binding:"required"`
	TimeSecond int  `json:"timeSecond"` // 读文章一共用了多久,预备字段,也许以后可以用于做喜好分析
}

// 创建历史记录
func (ArticleApi) ArticleLookView(c *gin.Context) {
	//这里浏览量的增加任务已经交给Redis了,由Redis在后台记录增量并定期刷新,由于浏览量的增加不需要特别强的一致性,所以在这里直接返回成功,真正的增加浏览量的任务交给Redis去做,这样可以大大提高接口的响应速度,并且可以防止刷浏览量的攻击,因为Redis天然支持去重,所以同一个用户在同一天内多次请求这个接口,只有第一次会增加浏览量,后续的请求都会被Redis拦截掉,不会增加浏览量
	//TODO:由于redis天然支持消息订阅,后台挂一个协程,攒够50~100条或时间到了之后的时候再一次性写入数据库,去重查询(防止有人刷浏览量)就用这个
	var req ArticleLookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.OkWithMsg("未登录", c) //未登录不给加浏览量
		return
	}

	// 引入缓存
	// 当天这个用户请求这个文章之后，将用户id和文章id作为key存入缓存，在这里进行判断，如果存在就直接返回
	if redis_count.GetUserArticleHistoryCache(req.ArticleID, claims.UserID) {
		response.OkWithMsg("成功", c)
		return
	}
	var article models.ArticleModel
	err = global.DB.Take(&article, "status = ? and id = ?", models.StatusPublished, req.ArticleID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	// 查这个文章今天有没有在足迹里面
	var history models.UserArticleHistoryModel
	err = global.DB.Take(&history,
		"user_id = ? and article_id = ? and created_at < ? and created_at > ?",
		claims.UserID, req.ArticleID,
		time.Now().Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02")+" 00:00:00",
	).Error
	if err == nil {
		response.OkWithMsg("成功", c)
		return
	}

	redis_count.SetCacheLook(req.ArticleID, true)

	err = global.DB.Create(&models.UserArticleHistoryModel{
		UserID:      claims.UserID,
		ArticleName: article.Title,
		ArticleID:   article.ID,
	}).Error
	if err != nil {
		response.FailWithMsg("失败", c)
		return
	}

	// 增加文章的浏览量
	response.OkWithMsg("成功", c)
}

type ArticleLookListRequest struct {
	common.PageInfo
	UserID uint `form:"userID"` //为0时是查询自己的浏览记录,否则是指定用户的浏览记录
}

type ArticleLookListResponse struct {
	ID        uint      `json:"id"`       // 浏览记录的id
	LookDate  time.Time `json:"lookDate"` // 浏览的时间
	Title     string    `json:"title"`
	Cover     string    `json:"cover"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	UserID    uint      `json:"userID"`
	ArticleID uint      `json:"articleID"`
}

func (ArticleApi) ArticleLookListView(c *gin.Context) { //除了可以记录浏览量之外,还可以根据用户id查询用户的浏览记录
	var req ArticleLookListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	claims, _ := jwts.ParseTokenByGin(c)
	switch req.UserID {
	case 0: //一会做一下适配
		if claims == nil {
			response.FailWithMsg("未登录", c)
			return
		}
		req.UserID = claims.UserID
	default:
		var user models.UserConfModel
		if global.DB.Take(&user, "user_id = ?", req.UserID).Error != nil { //检查这个用户是否存在
			response.FailWithMsg("用户不存在", c)
			return
		}
		if user.OpenHistory != true {
			response.FailWithMsg("用户未公开浏览记录", c)
			return
		}
	}

	_list, count, _ := common.ListQuery(models.UserArticleHistoryModel{
		UserID: req.UserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Likes:    []string{"article_name"},
		Preloads: []string{"UserModel", "ArticleModel"},
	})

	var list = make([]ArticleLookListResponse, 0)
	for _, model := range _list {
		list = append(list, ArticleLookListResponse{
			ID:        model.ID,
			LookDate:  model.CreatedAt,
			Title:     model.ArticleModel.Title,
			Cover:     model.ArticleModel.Cover,
			Nickname:  model.UserModel.NickName,
			Avatar:    model.UserModel.Avatar,
			UserID:    model.UserID,
			ArticleID: model.ArticleID,
		})
	}

	response.OkWithList(list, count, c)

}

func (ArticleApi) ArticleLookRemoveView(c *gin.Context) { //TODO:写一个定时任务?如果数据库里超过1个月的浏览记录,就删除
	var req models.RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	claims := jwts.GetClaims(c)
	var list []models.UserArticleHistoryModel
	global.DB.Find(&list, "user_id = ? and id in ?", claims.UserID, req.IDList) //TODO:可以在这里进行时间判断,如果超过1个月,就删除

	if len(list) > 0 {
		err := global.DB.Delete(&list).Error
		if err != nil {
			response.FailWithMsg("历史记录删除失败", c)
			return
		}
	}

	response.OkWithMsg(fmt.Sprintf("删除历史记录成功 共删除%d条", len(list)), c)
}
