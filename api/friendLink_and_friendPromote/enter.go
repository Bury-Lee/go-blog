package friendlink_and_friendpromote

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"fmt"

	"github.com/gin-gonic/gin"
)

type FriendApi struct{}

type FriendLinkCreateRequest struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Logo      string `json:"logo"`
	IsShow    bool   `json:"is_show"`
	SortOrder int    `json:"sort_order"`
	Remark    string `json:"remark"`
}

type FriendPromotionCreateRequest struct {
	Title         string   `json:"title"`
	FriendName    string   `json:"friend_name"`
	Avatar        string   `json:"avatar"`
	Category      string   `json:"category"`
	Description   string   `json:"description"`
	PreviewImages string   `json:"preview_images"`
	ContactInfo   []string `json:"contact_info"`
	IsShow        bool     `json:"is_show"`
	SortOrder     int      `json:"sort_order"`
	Position      string   `json:"position"`
	Remark        string   `json:"remark"`
}

func (FriendApi) FriendLinkCreateView(c *gin.Context) {
	var req FriendLinkCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}
	err := global.DB.Create(&models.FriendLink{
		Name:      req.Name,
		URL:       req.URL,
		Logo:      req.Logo,
		IsShow:    req.IsShow,
		SortOrder: req.SortOrder,
		Remark:    req.Remark,
	}).Error
	if err != nil {
		response.FailWithMsg("添加失败", c)
		return
	}
	response.OkWithMsg("添加成功", c)
}

func (FriendApi) FriendLinkListView(c *gin.Context) {
	var req common.PageInfo
	c.ShouldBind(&req)

	list, count, _ := common.ListQuery(models.FriendLink{
		IsShow: true,
	}, common.Options{
		PageInfo: req,
	})
	response.OkWithList(list, count, c)
}

func (FriendApi) FriendLinkRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return // 添加返回语句
	}
	// 执行删除操作
	result := global.DB.Where("id IN ?", req.IDList).Delete(&models.FriendLink{})
	if result.Error != nil {
		response.FailWithMsg("删除失败", c)
		return
	}
	response.OkWithMsg(fmt.Sprintf("成功删除%d个", result.RowsAffected), c)
}
func (FriendApi) FriendLinkUpdateView(c *gin.Context) {
	var req models.IDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	var model models.FriendLink
	err := global.DB.Take(&model, req.ID).Error
	if err != nil {
		response.FailWithMsg("未找到记录", c)
		return
	}
	var data FriendLinkCreateRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	if err := global.DB.Model(&model).Updates(map[string]any{
		"name":       data.Name,
		"url":        data.URL,
		"logo":       data.Logo,
		"is_show":    data.IsShow,
		"sort_order": data.SortOrder,
		"remark":     data.Remark,
	}).Error; err != nil {
		response.FailWithMsg("更新失败", c)
	} else {
		response.OkWithMsg("更新成功", c)
	}
}

func (FriendApi) FriendPromotionListView(c *gin.Context) {
	var req common.PageInfo
	c.ShouldBind(&req)

	list, count, _ := common.ListQuery(models.FriendPromotion{
		IsShow: true,
	}, common.Options{
		PageInfo: req,
	})
	response.OkWithList(list, count, c)
}

func (FriendApi) FriendPromotionRemoveView(c *gin.Context) {
	var req models.RemoveRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}
	// 执行删除操作
	result := global.DB.Where("id IN ?", req.IDList).Delete(&models.FriendPromotion{})
	if result.Error != nil {
		response.FailWithMsg("删除失败", c)
		return
	}
	response.OkWithMsg(fmt.Sprintf("成功删除%d个", result.RowsAffected), c)
}

func (FriendApi) FriendPromotionUpdateView(c *gin.Context) {
	var idReq models.IDRequest
	if err := c.ShouldBindUri(&idReq); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	var friendPromotion models.FriendPromotion
	err := global.DB.Take(&friendPromotion, idReq.ID).Error
	if err != nil {
		response.FailWithMsg("未找到记录", c)
		return
	}
	var updateReq FriendPromotionCreateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.FailWithMsg("绑定参数失败", c)
		return
	}
	if err := global.DB.Model(&friendPromotion).Updates(map[string]any{
		"title":          updateReq.Title,
		"friend_name":    updateReq.FriendName,
		"avatar":         updateReq.Avatar,
		"category":       updateReq.Category,
		"description":    updateReq.Description,
		"preview_images": updateReq.PreviewImages,
		"contact_info":   updateReq.ContactInfo,
		"is_show":        updateReq.IsShow,
		"sort_order":     updateReq.SortOrder,
		"position":       updateReq.Position,
		"remark":         updateReq.Remark,
	}).Error; err != nil {
		response.FailWithMsg("更新失败", c)
	} else {
		response.OkWithMsg("更新成功", c)
	}
}

func (FriendApi) FriendPromotionCreateView(c *gin.Context) {
	var req FriendPromotionCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数绑定失败", c)
		return
	}
	err := global.DB.Create(&models.FriendPromotion{
		Title:         req.Title,
		FriendName:    req.FriendName,
		Avatar:        req.Avatar,
		Category:      req.Category,
		Description:   req.Description,
		PreviewImages: req.PreviewImages,
		ContactInfo:   req.ContactInfo,
		IsShow:        req.IsShow,
		SortOrder:     req.SortOrder,
		Position:      req.Position,
		Remark:        req.Remark,
	}).Error
	if err != nil {
		response.FailWithMsg("添加失败", c)
		return
	}
	response.OkWithMsg("添加成功", c)
}
