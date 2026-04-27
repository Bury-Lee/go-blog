package ip

import (
	"net"
	"testing"
)

// TestHasLocalIPAddr 测试字符串IP地址是否为内网地址
// 说明:测试各种常见的内网IP段、公网IP、环回地址和无效IP字符串
func TestHasLocalIPAddr(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"Loopback IPv4", "127.0.0.1", true},
		{"Loopback IPv6", "::1", true},
		{"Class A Private", "10.0.0.1", true},
		{"Class B Private Start", "172.16.0.1", true},
		{"Class B Private End", "172.31.255.255", true},
		{"Class C Private", "192.168.1.100", true},
		{"Link Local", "169.254.1.1", true},
		{"Public IP 1", "8.8.8.8", false},
		{"Public IP 2", "114.114.114.114", false},
		{"Invalid IP", "invalid_ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasLocalIPAddr(tt.ip); got != tt.want {
				t.Errorf("HasLocalIPAddr(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// TestHasLocalIP 测试 net.IP 对象是否为内网地址
// 说明:直接传入 net.IP 进行边界测试
func TestHasLocalIP(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want bool
	}{
		{"Loopback", net.ParseIP("127.0.0.1"), true},
		{"Class A Private", net.ParseIP("10.255.255.255"), true},
		{"Public IP", net.ParseIP("1.1.1.1"), false},
		{"Nil IP", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasLocalIP(tt.ip); got != tt.want {
				t.Errorf("HasLocalIP(%v) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
