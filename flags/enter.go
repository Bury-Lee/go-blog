package flags

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	Hash "StarDreamerCyberNook/utils/hash"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

type Option struct {
	File    string //配置文件
	DB      bool   //迁移数据库
	Version bool   //查看版本
	Type    string
	Sub     string
	ES      bool //建立索引
}

var FlagOptions = new(Option) //一个指针命令行参数指针

func Parse() { //注册然后解析
	flag.StringVar(&FlagOptions.File, "f", "setting.yaml", "配置文件")
	flag.BoolVar(&FlagOptions.DB, "db", false, "数据库迁移")
	flag.BoolVar(&FlagOptions.ES, "es", false, "es建立索引")
	flag.BoolVar(&FlagOptions.Version, "v", false, "版本")
	flag.StringVar(&FlagOptions.Type, "t", "", "操作类型")
	flag.StringVar(&FlagOptions.Sub, "s", "", "子类型")
	flag.Parse() //解析部分
}

func Run() {
	if FlagOptions.DB {
		//执行数据库迁移
		FlagDB()
		os.Exit(0)
	}
	if FlagOptions.ES {
		EsIndex()
		os.Exit(0)
	}
	if FlagOptions.Version {
		fmt.Print("当前版本为:")
		fmt.Println(1)
		os.Exit(0)
	}
	switch FlagOptions.Type {
	case "user":
		switch FlagOptions.Sub {
		case "create":
			Create()
			os.Exit(0)
		}
	}
}

func Create() {
	var role enum.RoleType
	fmt.Println("选择角色     1 超级管理员   2 普通用户   3 访客")
	_, err := fmt.Scan(&role)
	if err != nil {
		logrus.Errorf("输入错误 %s", err)
		return
	}
	if !(role == 1 || role == 2 || role == 3) {
		logrus.Errorf("输入角色错误")
		return
	}

	var username string
	fmt.Println("请输入用户名:")
	fmt.Scan(&username)

	// 查用户名是否存在
	var model models.UserModel
	err = global.DB.Take(&model, "user_name = ?", username).Error
	if err == nil {
		logrus.Errorf("此用户名已存在")
		return
	}

	var email string
	fmt.Println("请输入邮箱:")
	fmt.Scan(&email)

	// 查用户名是否存在
	err = global.DB.Take(&model, "email = ?", email).Error
	if err == nil {
		logrus.Errorf("此邮箱已存在")
		return
	}

	fmt.Println("请输入密码:")
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("读取密码时出错:", err)
		return
	}
	fmt.Println("请再次输入密码:")
	rePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("读取密码时出错:", err)
		return
	}
	if string(password) != string(rePassword) {
		fmt.Println("两次密码不一致")
		return
	}
	var nickname string
	fmt.Println("请输入昵称(为空自动填入):")
	fmt.Scan(&nickname)
	if nickname == "" {
		nickname = "默认"
	}

	hashPwd, _ := Hash.HashPassword(string(password))
	// 创建用户
	err = global.DB.Create(&models.UserModel{
		UserName:       username,
		NickName:       nickname,
		RegisterSource: enum.RegisterTerminal,
		Password:       hashPwd,
		Email:          email,
		Role:           role,
		LastLoginTime:  time.Now(),
	}).Error
	if err != nil {
		fmt.Println("创建用户失败", err)
		return
	}
	logrus.Infof("创建用户成功")
}
