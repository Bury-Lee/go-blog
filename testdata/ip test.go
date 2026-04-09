// core/init_ip_db.go
package teas

import (
	ipUtils "StarDreamerCyberNook/utils/ip"
	"fmt"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/sirupsen/logrus"
)

var searcher *xdb.Searcher

func InitIPDB() {
	var dbPath = "init/ip2region.xdb"
	_searcher, err := xdb.NewWithFileOnly(dbPath)
	if err != nil {
		logrus.Fatalf("ip地址数据库加载失败 %s", err)
		return
	}
	searcher = _searcher
}

func GetIpAddr(ip string) (addr string) {
	if ipUtils.HasLocalIPAddr(ip) {
		return "内网"
	}

	region, err := searcher.SearchByStr(ip)
	if err != nil {
		logrus.Warnf("错误的ip地址 %s", err)
		return "异常地址"
	}
	_addrList := strings.Split(region, "|")
	if len(_addrList) != 5 {
		// 会有这个情况吗？
		logrus.Warnf("异常的ip地址 %s", ip)
		return "未知地址"
	}
	// _addrList 五个部分
	// 国家  0  省份   市   运营商
	country := _addrList[0]
	province := _addrList[2]
	city := _addrList[3]

	if province != "0" && city != "0" {
		return fmt.Sprintf("%s·%s", province, city)
	}
	if country != "0" && province != "0" {
		return fmt.Sprintf("%s·%s", country, province)
	}
	if country != "0" {
		return country
	}
	return region
}
