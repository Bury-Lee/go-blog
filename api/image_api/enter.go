package image_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/log_service"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//注:以后图片的操作都应该使用ID来查改,谁这么神人用哈希值来辨认

// ImageApi 图片管理API结构体
type ImageApi struct{}

// ImageListResponse 图片列表响应结构体
// 说明: 包含图片模型和Web访问路径
type ImageListResponse struct {
	models.ImageModel
	WebPath string `json:"webPath"` // Web访问路径
}

// ImageList 获取图片列表
// 参数: c - gin上下文
// 说明: 分页查询图片列表,支持文件名模糊搜索
func (ImageApi) ImageList(c *gin.Context) {
	var req common.PageInfo
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	// 查询图片列表
	_list, cout, err := common.ListQuery[models.ImageModel](models.ImageModel{}, common.Options{
		PageInfo: req,
		Likes:    []string{"filename"}, // 支持文件名模糊搜索
	})
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	// 构建响应数据
	var list = make([]ImageListResponse, 0)
	for _, model := range _list {
		list = append(list, ImageListResponse{
			ImageModel: model,
			WebPath:    model.WebPath(), // 获取Web访问路径
		})
	}
	response.OkWithList(list, cout, c)
}

// RemoveRequest 图片删除请求结构体
type RemoveRequest struct {
	IDlist []uint `json:"IDlist" binding:"required"` // 要删除的图片ID列表
}

// ImageRemoveView 批量删除图片
// 参数: c - gin上下文
// 说明: 根据ID列表批量删除图片,记录操作日志
func (ImageApi) ImageRemoveView(c *gin.Context) {
	var req RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(err, c)
		return
	}

	// 记录操作日志
	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	// 查询要删除的图片
	var list []models.ImageModel
	global.DB.Find(&list, "id IN ?", req.IDlist)

	// 批量删除图片
	if len(list) > 0 {
		err := global.DB.Delete(&list).Error
		if err != nil {
			logrus.Error(fmt.Sprintf("删除失败:%s", err))
		}
	}
	//TODO:考虑加入返回操作成功失败的个数
	response.OkWithMsg(fmt.Sprintf("图片删除成功,共删除%d张", len(list)), c)
}

// GetImage 获取图片文件
// 参数: c - gin上下文
// 说明: 根据URL参数返回图片文件,支持文件下载
// 路由: GET /api/image?url=xxx
func (ImageApi) GetImage(c *gin.Context) {
	id := c.Query("id")

	var img models.ImageModel
	// 构建图片路径 TODO:改为从配置中获取并且与图片上传的路径同步
	// 查询图片是否存在
	if global.DB.Take(&img, "id = ?", id).Error != nil {
		response.FailWithMsg("图片已被删除", c)
		return
	}

	// 返回图片文件
	c.File(img.Path)
}
