package banner_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"fmt"

	"github.com/gin-gonic/gin"
)

type BannerApi struct{}

type BannerCreateRequest struct {
	Cover  string `json:"cover"`
	Href   string `json:"href"`
	IsShow bool   `json:"isShow"`
}

func (BannerApi) BannerCreateView(c *gin.Context) { //TODO:如果图片不存在,可以当场上传一张
	var req BannerCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
	}
	var empty BannerCreateRequest
	if req == empty {
		response.FailWithMsg("参数错误", c)
		return
	}
	err := global.DB.Create(&models.BannerModel{
		Cover:  req.Cover,
		Href:   req.Href,
		IsShow: req.IsShow,
	}).Error
	if err != nil {
		response.FailWithMsg("添加失败", c)
		return
	}
	response.OkWithMsg("上传成功", c)
}
func (BannerApi) BannerListView(c *gin.Context) {
	var req common.PageInfo
	c.ShouldBind(&req)

	list, count, _ := common.ListQuery[models.BannerModel](&models.BannerModel{
		IsShow: true,
	}, common.Options{
		PageInfo: req,
	})
	response.OkWithList(list, count, c)
}

func (BannerApi) BannerRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
	}
	var list = []models.BannerModel{}
	global.DB.Find(&list, "id in ?", req.IDList)
	if len(list) > 0 {
		global.DB.Delete(&list)
	}
	response.OkWithMsg(fmt.Sprintf("成功删除%d个", len(list)), c)
}

func (BannerApi) BannerUpdateView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	var model models.BannerModel
	err := global.DB.Take(&model, req.ID).Error
	if err != nil {
		response.FailWithMsg("未找到记录", c)
	}
	var data BannerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	if err := global.DB.Model(&model).Updates(map[string]any{
		"cover":  data.Cover,
		"href":   data.Href,
		"isShow": data.IsShow,
	}).Error; err != nil {
		response.FailWithMsg("更新失败", c)
	} else {
		response.OkWithMsg("更新成功", c)
	}
}
