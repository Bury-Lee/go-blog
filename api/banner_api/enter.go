package banner_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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
	//更新缓存
	ctx := context.Background()
	global.RedisHotPool.Del(ctx, "banner_list")
	response.OkWithMsg("上传成功", c)
}

func (BannerApi) BannerListView(c *gin.Context) {
	//放缓存里,也在缓存查询
	ctx := context.Background()
	List, err := global.RedisHotPool.Get(ctx, "banner_list").Result()
	if err != nil && err != redis.Nil { //查询出错
		logrus.Errorf("查询缓存失败:%v,内容:%s", err, List)

	} else if err == nil { //查询到了,返回
		var cached []models.BannerModel
		err := json.Unmarshal([]byte(List), &cached)
		if err != nil {
			logrus.Errorf("缓存数据解析错误:%v,内容:%s", err, List)
			response.FailWithMsg("缓存数据解析错误", c)
			return
		}
		response.OkWithList(cached, len(cached), c)
		return
	}
	var req common.PageInfo
	c.ShouldBind(&req)

	list, count, _ := common.ListQuery(models.BannerModel{
		IsShow: true,
	}, common.Options{
		PageInfo: req,
	})
	jsonData, err := json.Marshal(list)
	if err != nil {
		logrus.Error("缓存数据序列化错误:", err)
	}
	//把数据加入缓存
	global.RedisHotPool.Set(ctx, "banner_list", string(jsonData), 0)
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
	//更新缓存
	ctx := context.Background()
	global.RedisHotPool.Del(ctx, "banner_list")
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
	if err := c.ShouldBindJSON(&data); err != nil {
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
	//更新缓存
	ctx := context.Background()
	global.RedisHotPool.Del(ctx, "banner_list")
}
