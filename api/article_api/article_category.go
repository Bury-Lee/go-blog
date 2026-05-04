package article_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CategoryCreateRequest struct {
	ID    uint   `json:"id"` //id为0时为创建，否则为修改
	Title string `json:"title" binding:"required,max=32"`
}

func (ArticleApi) CategoryCreateView(c *gin.Context) {
	var req CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	var model models.CategoryModel
	if req.ID == 0 {
		// 创建
		err := global.DB.Take(&model, "user_id = ? and title = ?", claims.UserID, req.Title).Error
		if err == nil {
			response.FailWithMsg("分类名称重复", c)
			return
		}

		err = global.DB.Create(&models.CategoryModel{
			Title:  req.Title,
			UserID: claims.UserID,
		}).Error
		if err != nil {
			response.FailWithMsg("创建分类错误", c)
			return
		}

		response.OkWithMsg("创建分类成功", c)
		return
	}

	err := global.DB.Take(&model, "user_id = ? and id = ?", claims.UserID, req.ID).Error
	if err != nil {
		response.FailWithMsg("分类不存在", c)
		return
	}

	err = global.DB.Model(&model).Update("title", req.Title).Error

	if err != nil {
		response.FailWithMsg("更新分类错误", c)
		return
	}

	response.OkWithMsg("更新分类成功", c)
}

type CategoryListRequest struct {
	common.PageInfo
	UserID uint   `form:"userID"`
	Type   string `form:"type" binding:"required"` // self 查自己 other 查别人 admin 后台
}

type CategoryListResponse struct {
	models.CategoryModel
	ArticleCount int    `json:"articleCount"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
}

func (ArticleApi) CategoryListView(c *gin.Context) {
	var req CategoryListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	var preload = []string{"ArticleList"}
	switch req.Type {
	case "self":
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil {
			response.FailWithMsg("未登录", c)
			return
		}
		req.UserID = claims.UserID
	case "other":
		if req.UserID == 0 {
			response.FailWithMsg("用户ID不能为空", c)
			return
		}
	case "admin":
		claims, err := jwts.ParseTokenByGin(c)
		if err != nil {
			response.FailWithMsg("未登录", c)
			return
		}
		if claims.Role != enum.AdminRole {
			response.FailWithMsg("权限错误", c)
			return
		}
		preload = append(preload, "UserModel")
	default:
		response.FailWithMsg("类型错误", c)
		return
	}

	_list, count, _ := common.ListQuery(models.CategoryModel{
		UserID: req.UserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Likes:    []string{"title"},
		Preloads: preload,
	})

	var list = make([]CategoryListResponse, 0)
	for _, i2 := range _list {
		list = append(list, CategoryListResponse{
			CategoryModel: i2,
			ArticleCount:  len(i2.ArticleList),
			Nickname:      i2.UserModel.NickName,
			Avatar:        i2.UserModel.Avatar,
		})
	}

	response.OkWithList(list, count, c)
}

func (ArticleApi) CategoryRemoveView(c *gin.Context) {
	var req = models.RemoveRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("请求参数错误", c)
		return
	}

	var list []models.CategoryModel
	query := global.DB.Where("id in ?", req.IDList)
	claims := jwts.GetClaims(c)
	if claims.Role != enum.AdminRole {
		query.Where("user_id = ?", claims.UserID)
	}

	global.DB.Where(query).Find(&list)

	if len(list) > 0 {
		err := global.DB.Delete(&list).Error
		if err != nil {
			logrus.Error("删除分类失败", err)
			response.FailWithMsg("删除分类失败", c)
			return
		}
	}

	msg := fmt.Sprintf("删除分类成功 共删除%d条", len(list))

	response.OkWithMsg(msg, c)
}

func (ArticleApi) CategoryOptionsView(c *gin.Context) { //是文章分类选项列表接口
	claims := jwts.GetClaims(c)

	var list []models.OptionsResponse[uint]
	global.DB.Model(models.CategoryModel{}).Where("user_id = ?", claims.UserID).
		Select("id as value", "title as label").Scan(&list)

	response.OkWithData(list, c)

}
