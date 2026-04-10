package ip

import (
	"StarDreamerCyberNook/global"
	"fmt"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetIpAddr 根据IP地址获取地理位置信息
// 参数:ip - 要查询的IP地址字符串
// 返回:addr - 格式化后的地理位置字符串，格式为"省份·城市"或"国家·省份"等
// 说明:对于本地IP地址，返回"未知的本地ip"对于无效IP地址，返回"异常地址",对于格式异常的查询结果，返回"未知地址",优先显示省份和城市信息，其次是国家信息
func GetIpAddr(ip string) (addr string) {
	if HasLocalIPAddr(ip) {
		return "本地ip"
	}

	region, err := global.IPsearcher.Search(ip)
	if err != nil {
		logrus.Warnf("错误的ip地址 %s", err)
		return "异常地址"
	}
	_addrList := strings.Split(region, "|")
	if len(_addrList) != 5 {
		// 数据库返回的格式异常，记录警告日志
		logrus.Warnf("异常的ip地址 %s", ip)
		return "异常地址"
	}

	// _addrList 五个部分分别代表：
	// 国家(0) | 区域(1) | 省份(2) | 城市(3) | 运营商(4)
	country := _addrList[0]
	province := _addrList[2]
	city := _addrList[3]

	// 按照优先级格式化地址信息
	// 1. 优先显示省份和城市（当两者都有效时）
	if province != "0" && city != "0" {
		return fmt.Sprintf("%s·%s", province, city)
	}
	// 2. 其次显示国家和省份
	if country != "0" && province != "0" {
		return fmt.Sprintf("%s·%s", country, province)
	}
	// 3. 最后只显示国家
	if country != "0" {
		return country
	}
	// 4. 如果以上都无效，返回原始查询结果
	return region
}

func HasLocalIPAddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocaLIP 检测 IP 地址是否是内网地址// 通过直接对比ip段范围效率更高
func HasLocalIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}
