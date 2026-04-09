package conf

type Redis struct { //redis配置
	Addr     string `yaml:"addr"`     //redis地址
	Password string `yaml:"password"` //redis密码
	DB       int    `yaml:"db"`       //redis数据库
}
