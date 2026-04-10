// core/init_db.go
package core

import (
	"StarDreamerCyberNook/global"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// InitDB 初始化数据库连接
// 参数: 无
// 返回: *gorm.DB - 数据库连接实例
// 说明: 连接主库,配置连接池,支持读写分离
func InitDB() *gorm.DB {
	if len(global.Config.DB) == 0 {
		logrus.Fatalf("未配置数据库")
	}

	// 获取数据库配置
	dc := global.Config.DB[0] //写库

	//要求数据库配置要一致
	for i := 1; i < len(global.Config.DB)-1; i++ {
		v := global.Config.DB[i]
		logrus.Infof("数据库配置: 模式=%s, 主机=%s, 端口=%d, 用户=%s, 数据库=%s",
			v.SqlName, v.Host, v.Port, v.User, v.DBName)
		if v.SqlName != global.Config.DB[i-1].SqlName {
			logrus.Fatalf("数据库配置错误: 模式不一致, %s != %s", v.SqlName, global.Config.DB[i-1].SqlName)
		}
	}

	// 连接主数据库
	db, err := gorm.Open(dc.DSN(), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // 不生成外键约束
	})
	if err != nil {
		logrus.Fatalf("数据库连接失败 %s", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
	logrus.Infof("数据库连接成功")

	if len(global.Config.DB) > 1 {
		// 注册读写分离配置
		var readList []gorm.Dialector
		for _, v := range global.Config.DB[1:] {
			readList = append(readList, v.DSN())
		}
		err = db.Use(dbresolver.Register(dbresolver.Config{
			// 使用 db0 作为主库（sources），db1 作为从库（replicas）
			Sources:  []gorm.Dialector{dc.DSN()}, //写库,如果需要就自己在列表里加
			Replicas: readList,                   //读库
			//Replicas: []gorm.Dialector{mysql.Open(read_dc.DSN()), mysql.Open("db4_dsn")}这样可以注册多个读库
			// 负载均衡策略：随机选择 replica
			Policy: dbresolver.RandomPolicy{},
		}))
		if err != nil {
			logrus.Fatalf("读写配置出错: %s", err)
		}
	}

	if global.Config.System.RunMode == "debug" {
		db := db.Debug()
		logrus.Debug("数据库调试模式已开启")
		return db
	}
	return db
}
