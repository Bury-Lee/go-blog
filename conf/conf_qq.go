package conf

import "fmt"

type QQ struct { //预留配置:QQ邮箱
	AppID    string `yaml:"appID" json:"appID"`       // QQ邮箱应用ID
	AppKey   string `yaml:"appKey" json:"appKey"`     // QQ邮箱应用密钥
	Redirect string `yaml:"redirect" json:"redirect"` // QQ邮箱回调URI
}

func (this QQ) Url() string {
	return fmt.Sprintf("https://graph.qq.com/oauth2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=get_user_info", this.AppID, this.Redirect)
} //TODO:注:该回调地址一定要在QQ官方配置过之后才能真正使用,并且要填写自己的回调地址
