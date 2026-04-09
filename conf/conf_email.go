package conf

type Email struct { //邮箱
	Domain       string `yaml:"domain" json:"domain"`             // 邮箱域名
	Port         int    `yaml:"port" json:"port"`                 // 邮箱SMTP服务器端口
	SendEmail    string `yaml:"sendEmail" json:"sendEmail"`       // 发送邮箱
	AuthCode     string `yaml:"authCode" json:"authCode"`         // api代码一类的?
	SendNickname string `yaml:"sendNickname" json:"sendNickname"` // 发信人昵称
	SSL          bool   `yaml:"SSL" json:"SSL"`                   // 是否启用SSL
	TLS          bool   `yaml:"TLS" json:"TLS"`                   // 是否启用TLS
}
