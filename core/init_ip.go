// core/init_ip_db.go
package core

import (
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/sirupsen/logrus"
)

// searcher 是IP地址数据库查询器实例，用于IP地址到地理位置的查询

// InitIPDB 初始化IP地址数据库
// 加载ip2region.xdb文件并创建查询器实例
// 如果数据库文件加载失败，程序将终止运行
func InitIPDB() *xdb.Searcher {
	var dbPath = "init/ip2region.xdb"
	searcher, err := xdb.NewWithFileOnly(xdb.IPv4, dbPath)
	if err != nil {
		logrus.Fatalf("ip地址数据库加载失败 %s", err)
		return nil
	}
	return searcher
}
