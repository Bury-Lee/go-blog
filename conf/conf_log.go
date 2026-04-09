package conf

type Log struct {
	App string `yaml:"app"` //来自哪个服务
	Dir string `yaml:"dir"` //日志目录
}
