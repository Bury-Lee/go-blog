package chat_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

//TODO:这里迟点再检查了,怕有bug

type ChatSendRequest struct {
	Msg       models.ChatMsg `json:"msg"`
	RevUserID uint           `json:"revUserID"` // 接收用户ID
}

func (ChatApi) ChatSendView(c *gin.Context) { //查询会话是否存在,没有就创建一个,有的话就更新会话信息
	var req ChatSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("消息发送失败", c)
		return
	}
	claim := jwts.GetClaims(c)
	// 加入检查是否有好友关系
	var revUser models.UserFollowModel
	global.DB.Where("(user_id = ? AND focus_user_id = ?) or (user_id = ? AND focus_user_id = ?)", claim.UserID, req.RevUserID, req.RevUserID, claim.UserID).First(&revUser)
	if !revUser.Friend {
		response.FailWithMsg("您不是好友，不能发送消息", c)
		return
	}
	//TODO:以后加入拉黑机制
	var msgType models.ChatMessageType
	if req.Msg.ImageMsg != nil {
		msgType = msgType | models.ImageMsgType
	}
	if req.Msg.MarkdownMsg != nil {
		msgType = msgType | models.MarkdownMsgType
	}
	if req.Msg.TextMsg != nil {
		msgType = msgType | models.TextMsgType
	}

	if msgType == 0 {
		response.FailWithMsg("不能发送空消息", c)
		return
	}

	var chatModel = models.ChatModel{
		SendUserID: claim.UserID,
		RevUserID:  req.RevUserID,
		Msg:        req.Msg,
		MsgType:    msgType,
	}
	global.DB.Create(&chatModel)

	//我给别人发了消息,那改变的就应该是别人的会话状态
	var session = models.SessionModel{
		UniqueID: fmt.Sprintf("%d%d", req.RevUserID, claim.UserID),
	}

	err := global.DB.Where("unique_id = ?", session.UniqueID).First(&session)
	if session.ID == 0 {
		response.FailWithMsg("?", c)
		return
	}
	if err != nil {
		session = models.SessionModel{
			UniqueID:        fmt.Sprintf("%d%d", req.RevUserID, claim.UserID),
			UserID:          req.RevUserID,
			LastMessage:     chatModel,
			LastMessageType: msgType,
			LastMessageTime: time.Now(),
			IsRead:          false,
			UnreadCount:     1,
		}
		global.DB.Create(&session)
		return
	} else { //更新会话信息
		session.IsRead = false
		session.UnreadCount += 1
		session.LastMessage = chatModel
		session.LastMessageType = msgType
		session.LastMessageTime = time.Now()
		global.DB.Save(&session)
	}
}

type ChatListRequest struct {
	common.PageInfo
	UserID uint `form:"userID" binding:"required"` // 查我和他的聊天记录
}

type ChatListResponse struct {
	models.ChatModel
	SendUserNickname string `json:"sendUserNickname"`
	SendUserAvatar   string `json:"sendUserAvatar"`
	RevUserNickname  string `json:"revUserNickname"`
	RevUserAvatar    string `json:"revUserAvatar"`
	IsMe             bool   `json:"isMe"` // 是自己发的吗
}

// TODO:撤回和删除消息功能

func (ChatApi) ChatListView(c *gin.Context) { // 查自己和指定用户的聊天记录
	var req ChatListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)
	query := global.DB.Where("(send_user_id = ? and rev_user_id = ?) or(send_user_id = ? and rev_user_id = ?) ",
		req.UserID, claims.UserID, claims.UserID, req.UserID,
	)

	req.Order = "created_at desc"
	_list, count, err := common.ListQuery[models.ChatModel](models.ChatModel{}, common.Options{
		PageInfo: req.PageInfo,
		Preloads: []string{"SendUserModel", "RevUserModel"},
		Where:    query,
	})
	if err != nil {
		response.FailWithMsg("查询聊天记录失败", c)
		return
	}

	// 提取所有消息ID
	var chatIDs []uint
	for _, v := range _list {
		chatIDs = append(chatIDs, v.ID)
	}

	// 如果没有消息ID，直接返回空列表
	if len(chatIDs) == 0 {
		var list []ChatListResponse
		response.OkWithList(list, count, c)
		return
	}

	// // 根据消息ID查询用户操作记录，获取删除状态
	// var userActions []models.UserChatActionModel
	// if err := global.DB.Where("chat_id IN ? and user_id = ? and is_delete = ?", chatIDs, claims.UserID, true).
	// 	Find(&userActions).Error; err != nil {
	// 	response.FailWithMsg("查询用户操作记录失败", c)
	// 	return
	// } //一会检查一下,看看是这个字段吗

	// // 创建一个map来快速查找某个消息是否被删除
	// deletedMap := make(map[uint]bool)
	// for _, action := range userActions {
	// 	if action.IsDelete {
	// 		deletedMap[action.ChatID] = true
	// 	}
	// }

	var list = make([]ChatListResponse, 0)
	for _, model := range _list {
		// // 检查该消息是否被当前用户删除
		// if deletedMap[model.ID] {
		// 	continue // 跳过已删除的消息
		// }

		//屏蔽掉配置字段
		model.SendUserModel.Password = ""
		model.SendUserModel.UserName = ""
		model.SendUserModel.Email = ""
		model.SendUserModel.OpenID = ""
		model.SendUserModel.Role = enum.UserRole
		model.SendUserModel.UserConfModel = nil

		model.RevUserModel.Password = ""
		model.RevUserModel.UserName = ""
		model.RevUserModel.Email = ""
		model.RevUserModel.OpenID = ""
		model.RevUserModel.Role = enum.UserRole
		model.RevUserModel.UserConfModel = nil

		item := ChatListResponse{
			ChatModel:        model,
			SendUserNickname: model.SendUserModel.NickName,
			SendUserAvatar:   model.SendUserModel.Avatar,
			RevUserNickname:  model.RevUserModel.NickName,
			RevUserAvatar:    model.RevUserModel.Avatar,
		}
		if model.SendUserID == claims.UserID {
			item.IsMe = true
		}
		list = append(list, item)
	}
	response.OkWithList(list, count, c)

	//更新会话信息
	//TODO:没有对话就要创建一个会话
	//查询了就是把消息读取了
	var session = models.SessionModel{
		UniqueID: fmt.Sprintf("%d%d", claims.UserID, req.UserID),
	}

	err = global.DB.Where("unique_id = ?", session.UniqueID).First(&session).Error
	if session.ID == 0 {
		response.FailWithMsg("没有对话记录", c)
		return
	}
	if err != nil {
		session = models.SessionModel{
			UniqueID:        fmt.Sprintf("%d%d", claims.UserID, req.UserID),
			UserID:          req.UserID,
			LastMessageTime: time.Now(),
			IsRead:          true,
			UnreadCount:     0,
		}
		global.DB.Create(&session)
	}
	//更新会话信息
	session.IsRead = true
	session.UnreadCount = 0
	global.DB.Save(&session)
}

type SessionListRequest struct {
	common.PageInfo
}

func (ChatApi) SessionListView(c *gin.Context) { // 查我的会话列表
	var req SessionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	claims := jwts.GetClaims(c)

	list, count, err := common.ListQuery[models.SessionModel](
		models.SessionModel{},
		common.Options{
			PageInfo:     req.PageInfo,
			Where:        global.DB.Where("unique_id LIKE ?", fmt.Sprintf("%d%%", claims.UserID)),
			Preloads:     []string{"UserModel"},
			DefaultOrder: "last_message_time desc",
		},
	)
	if err != nil {
		response.FailWithMsg("查询会话列表失败", c)
		return
	}

	response.OkWithList(list, count, c)
}
