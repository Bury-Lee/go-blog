package conf

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// import "fmt"

// type DB struct {
// 	User     string `yaml:"user"`
// 	Password string `yaml:"password"`
// 	Host     string `yaml:"host"`
// 	Port     int    `yaml:"port"`
// 	DB       string `yaml:"db"`
// 	Debug    bool   `yaml:"debug"`  //是否启用打印全部消息
// 	Source   string `yaml:"source"` //数据库的源,默认mysql,可能会有pgsql
// }

// func (this *DB) DSN() string { //应该不用使用结构体复制
// 	if this.Source != "mysql" {
// 		return "DB.source填写不符合标准?"
// 	}
// 	return fmt.Sprintf(
// 		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
// 		this.User, this.Password, this.Host, this.Port, this.DB)
// }

// func (this *DB) Empty() bool {
// 	return this.User == "" && this.Password == "" && this.Host == "" && this.Port == 0
// }

type SqlName string

type DB struct {
	SqlName  SqlName `yaml:"sql_name"` // 模式 mysql pgsql sqlite
	DBName   string  `yaml:"db_name"`
	Host     string  `yaml:"host"`
	Port     int     `yaml:"port"`
	User     string  `yaml:"user"`
	Password string  `yaml:"password"`
}

const (
	DBMysqlMode  = "mysql"
	DBPgsqlMode  = "pgsql"
	DBSqliteMode = "sqlite"
)

func (db *DB) DSN() gorm.Dialector {
	switch db.SqlName {
	case DBMysqlMode:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.User,
			db.Password,
			db.Host,
			db.Port,
			db.DBName,
		)
		return mysql.Open(dsn)
	case DBPgsqlMode:
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			db.Host,
			db.User,
			db.Password,
			db.DBName,
			db.Port,
		)
		return postgres.Open(dsn)
	case DBSqliteMode:
		return sqlite.Open(db.DBName)
	default:
		logrus.Warnf("未配置mysql连接")
		return nil
	}
}
