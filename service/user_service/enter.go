package user_service

import (
	// 核心工具包，包含IP地址解析等功能
	"StarDreamerCyberNook/global" // 全局变量包，包含数据库连接等
	"StarDreamerCyberNook/models" // 数据模型包，包含所有数据库表结构
	"StarDreamerCyberNook/utils/ip"

	"github.com/gin-gonic/gin"   // Gin Web框架，用于处理HTTP请求
	"github.com/sirupsen/logrus" // 日志库，用于记录程序运行日志
)

// UserService 用户服务结构体
// 封装与用户相关的业务逻辑操作
type UserService struct {
	UserModel models.UserModel // 用户模型实例，包含用户基础信息
}

// New 创建用户服务实例
// 参数: User - 用户模型对象
// 返回: *UserService - 用户服务实例指针
func New(User models.UserModel) *UserService {
	return &UserService{
		UserModel: User,
	}
}

// UserLogin 记录用户登录信息
// 获取用户登录时的IP地址、地理位置、设备信息并保存到数据库
// 参数: c - Gin上下文对象，包含请求相关信息
func (this *UserService) UserLogin(c *gin.Context) {
	// 获取客户端IP地址
	ipAdd := c.ClientIP()
	// 根据IP地址获取地理位置信息
	addr := ip.GetIpAddr(ipAdd)
	// 创建用户登录记录并保存到数据库
	err := global.DB.Create(&models.UserLoginModel{
		UserID:    this.UserModel.ID,         // 当前登录用户的ID
		IP:        ipAdd,                     // 客户端IP地址
		Addr:      addr,                      // IP对应的地理位置
		UserAgent: c.GetHeader("User-Agent"), // 用户代理信息（浏览器/设备信息）
	}).Error
	if err != nil {
		// 记录数据库写入失败错误日志
		logrus.Errorf("写入失败:%v", err)
	}
}
