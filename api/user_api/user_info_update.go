package user_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	jwts "StarDreamerCyberNook/utils/jwts"
	utils_other "StarDreamerCyberNook/utils/other"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type UserInfoUpdateRequest struct { //TODO:这个到时候要做测试
	Avatar      *string            `json:"avatar" s-u:"avatar"`
	Abstract    *string            `json:"abstract" s-u:"abstract"`
	LikeTags    *[]string          `json:"likeTags" s-u:"like_tags"`
	NickName    *string            `json:"nickName" s-u:"nick_name"`       // 昵称
	Age         *int               `json:"Age" s-u:"age"`                  // 年龄
	ContactInfo *map[string]string `json:"contactInfo" s-u:"contact_info"` // 联系方式，JSON格式存储
	// Email       *string            `json:"email" s-u:"email"`              // 邮箱，唯一索引//这个和登录账号相关,暂时不放在这里更新

	OpenCollect *bool `json:"openCollect" s-u-c:"open_collect"`  // 公开我的收藏
	OpenFollow  *bool `json:"openFollow" s-u-c:"open_follow"`    // 公开我的关注
	OpenFans    *bool `json:"openFans" s-u-c:"open_fans"`        // 公开我的粉丝
	HomeStyleID *uint `json:"homeStyleID" s-u-c:"home_style_id"` // 主页样式的id
}

func (UserApi) UserInfoUpdateView(c *gin.Context) {
	var req UserInfoUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMsg(err.Error(), c)
		return
	}

	if req.LikeTags != nil && len(*req.LikeTags) > 36 {
		response.FailWithMsg("喜欢的标签数量太多啦", c)
		return
	}
	//ai审核环节
	if global.Config.AI.Enable { //启用ai审核
		ctx := context.Background()

		// 构建完整的消息列表
		var messages []openai.ChatCompletionMessage

		// 添加系统提示词作为第一条消息

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: global.SystemPromptUser.String(),
		})

		// 添加增量更新的值
		if req.Abstract != nil {
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "用户简介:" + *req.Abstract,
				})
		}
		if req.NickName != nil {
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "用户昵称:" + *req.NickName,
				})
		}
		if req.LikeTags != nil {
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "用户喜欢的标签:" + fmt.Sprintf("%v", *req.LikeTags),
				})
		}
		// 创建非流式请求
		res, err := global.LocalAIClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    global.Config.AI.Model,
				Messages: messages,
			},
		)
		if err != nil {
			logrus.Errorf("ai审核失败: %s", err.Error())
			//出错自动降级为非ai流程
		}
		switch res.Choices[0].Message.Content { //TODO:这里无论成功还是失败都应该插入消息,告知原因
		case "通过":
			//通过,更新用户信息
		case "拒绝":
			//拒绝,返回错误
			response.FailWithMsg("存在违规信息,用户信息未更新", c)
			return
		default:
			logrus.Errorf("ai审核结果未知: %s,用户:%+v", res.Choices[0].Message.Content, req)
			response.FailWithMsg("审核结果未知,用户信息未更新", c) //也许也可以考虑放行?
			return
		}
	}

	userMap := utils_other.StructToMap(req, "s-u")
	userConfMap := utils_other.StructToMap(req, "s-u-c")
	// fmt.Println("userMap", userMap)
	// fmt.Println("userConfMap", userConfMap)

	claims := jwts.GetClaims(c)

	if len(userMap) > 0 {
		var userModel models.UserModel
		err = global.DB.Preload("UserConfModel").Take(&userModel, claims.UserID).Error
		if err != nil {
			response.FailWithMsg("用户不存在", c)
			return
		}
		// //TODO:修改用户名的逻辑,以后移植到专门的路由
		// if cr.username != nil {
		// 	var userCount int64
		// 	global.DB.Debug().Model(models.UserModel{}).
		// 		Where("username = ? and id <> ?", *cr.Username, claims.UserID).
		// 		Count(&userCount)
		// 	fmt.Println(*cr.Username, userCount)
		// 	if userCount > 0 {
		// 		res.FailWithMsg("该用户名被使用", c)
		// 		return
		// 	}
		// 	if *cr.Username != userModel.Username {
		// 		// 如果和我的用户名是一样的
		// 		var uud = userModel.UserConfModel.UpdateUsernameDate
		// 		if uud != nil {
		// 			if time.Now().Sub(*uud).Hours() < 720 {
		// 				res.FailWithMsg("用户名30天内只能修改一次", c)
		// 				return
		// 			}
		// 		}
		// 		userConfMap["update_username_date"] = time.Now()
		// 	}
		// }

		if req.NickName != nil || req.Avatar != nil {
		}

		err = global.DB.Model(&userModel).Updates(userMap).Error
		if err != nil {
			response.FailWithMsg("用户信息修改失败", c)
			return
		}
	}
	if len(userConfMap) > 0 {
		var userConfModel models.UserConfModel
		err = global.DB.Take(&userConfModel, "user_id = ?", claims.UserID).Error
		if err != nil {
			response.FailWithMsg("用户配置信息不存在", c)
			return
		}
		err = global.DB.Model(&userConfModel).Updates(userConfMap).Error
		if err != nil {
			response.FailWithMsg("用户信息修改失败", c)
			return
		}
	}
	response.OkWithMsg("用户信息修改成功", c)
}
