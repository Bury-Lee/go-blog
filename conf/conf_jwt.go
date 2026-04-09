package conf

type Jwt struct {
	AccessExpire       int    `yaml:"accessExpire"`       //过期时间,单位为分钟
	RefreshExpire      int    `yaml:"refreshExpire"`      //刷新令牌过期时间,单位为小时
	AccessTokenSecret  string `yaml:"accessTokenSecret"`  //JWT密钥
	RefreshTokenSecret string `yaml:"refreshTokenSecret"` //刷新令牌密钥
	Issuer             string `yaml:"issuer"`             //JWT签发者
}
