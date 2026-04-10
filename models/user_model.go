// models/user_model.go
// 用户模型定义
// 包含用户基本信息和用户配置信息，采用垂直分表设计
package models

import (
	"StarDreamerCyberNook/models/enum"
	"errors"
	"time"

	"gorm.io/gorm"
)

// UserModel 用户模型
// 存储用户核心信息，与用户配置表进行垂直分表
// 将高频访问的核心字段（登录、基本信息）与低频访问的配置字段分离
// 主表更轻量，配置表的扩展性更好
// 当使用的时候password,Username,Email,OpenID,UserConfModel这些字段要屏蔽掉
// TODO:上一次登录日期,前端动态计算出距今多少天未登录
type UserModel struct {
	Model
	UserName       string              `gorm:"size:32;uniqueIndex" json:"username"`          // 用户名，唯一索引
	NickName       string              `gorm:"size:32" json:"nickname"`                      // 昵称
	Avatar         string              `gorm:"size:512" json:"avatar"`                       // 头像URL
	Abstract       string              `gorm:"size:512" json:"abstract"`                     // 个人简介
	RegisterSource enum.RegisterSource `json:"registerSource"`                               // 注册来源枚举
	Age            int                 `json:"Age"`                                          // 年龄
	Password       string              `gorm:"size:64" json:"-"`                             // 哈希密码，不序列化到JSON
	ContactInfo    map[string]string   `gorm:"type:json;serializer:json" json:"contactInfo"` // 联系方式，JSON格式存储
	Email          string              `gorm:"size:128;uniqueIndex" json:"email"`            // 邮箱，唯一索引
	OpenID         string              `gorm:"size:128" json:"openID"`                       // 第三方登录唯一ID
	Role           enum.RoleType       `json:"role"`                                         // 用户角色枚举：1-管理员,2-VIP用户,3-普通用户,4-游客,5-封禁用户
	UserConfModel  *UserConfModel      `gorm:"foreignKey:UserID" json:"-"`                   // 用户配置信息，外键关联
	LikeTags       []string            `gorm:"type:text;serializer:json" json:"likeTags"`    // 兴趣标签列表，JSON格式存储
	LastLoginTime  time.Time           `json:"lastLoginTime"`                                // 上次登录时间
	IP             string              `gorm:"size:64" json:"IP"`                            // IP属地
}

func (self UserModel) AfterCreate(tx *gorm.DB) error { //用户创建后,自动创建用户配置
	// 钩子函数创建用户配置
	err1 := tx.Create(&UserConfModel{
		UserID:      self.ID,
		OpenCollect: true,
		OpenFollow:  true,
		OpenFans:    true, //默认开启
		HomeStyleID: 1,    //默认样式为1
	}).Error
	err2 := tx.Create(&UserMessageConfModel{
		UserID:             self.ID,
		OpenCommentMessage: true,
		OpenReplyMessage:   true,
		OpenDiggMessage:    true,
		OpenCollectMessage: true,
		OpenPrivateMessage: true, //默认开启
	}).Error
	if err1 != nil || err2 != nil {
		return errors.New("创建用户配置失败")
	}
	return nil
}

// UserConfModel 用户配置模型
// 存储用户的隐私设置和个性化配置
// 与用户主表进行垂直分表，提高查询性能
// TODO:把open改为public,更符合语法习惯
type UserConfModel struct {
	UserID             uint       `gorm:"primaryKey;unique" json:"userID"` // 用户ID，唯一索引
	UserModel          UserModel  `gorm:"foreignKey:UserID" json:"-"`      // 关联的用户信息
	UpdateUsernameDate *time.Time `json:"updateUsernameDate"`              // 上次修改用户名时间，使用指针区分是否修改过
	OpenCollect        bool       `json:"openCollect"`                     // 公开我的收藏
	OpenFollow         bool       `json:"openFollow"`                      // 公开我的关注
	OpenFans           bool       `json:"openFans"`                        // 公开我的粉丝
	OpenHistory        bool       `json:"openHistory"`                     // 公开我的浏览记录
	HomeStyleID        uint       `json:"homeStyleID"`                     // 主页样式ID
}

// ExistDays 计算用户存在天数
// 返回从用户创建时间到当前时间的天数
// 参数: 无
// 返回: int - 用户存在的天数
func (this *UserModel) ExistDays() int {
	duration := time.Since(this.CreatedAt)
	return int(duration.Hours()/24 + 1)
}

type UserMessageConfModel struct {
	UserID             uint      `gorm:"primaryKey;unique" json:"userID"`    // 用户id
	UserModel          UserModel `gorm:"foreignKey:UserID" json:"userModel"` // 用户信息
	OpenCommentMessage bool      `json:"openCommentMessage"`                 // 是否开启评论通知
	OpenReplyMessage   bool      `json:"openReplyMessage"`                   // 是否开启回复通知
	OpenDiggMessage    bool      `json:"openDiggMessage"`                    // 是否开启点赞	通知
	OpenCollectMessage bool      `json:"openCollectMessage"`                 // 是否开启收藏通知
	OpenPrivateMessage bool      `json:"openPrivateMessage"`                 // 是否开启私信通知
}
