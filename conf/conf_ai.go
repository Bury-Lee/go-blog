package conf

type AI struct { //AI模型配置
	Enable bool `yaml:"enable" json:"enable"` // 是否启用AI模型

	Model       string  `yaml:"model" json:"model"`             // AI模型名称,为local时使用本地模型
	Temperature float32 `yaml:"temperature" json:"temperature"` // 温度参数，控制生成文本的随机性
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`   // 最大生成令牌数
	Host        string  `yaml:"host" json:"host"`               // 本地AI模型主机地址,默认http://localhost:1234/v1,当model为local时生效
	APIType     string  `yaml:"api_type" json:"api_type"`       // AI模型API类型,默认openai

	ApiKey   string `yaml:"ApiKey" json:"-"`          // AI模型密钥
	NickName string `yaml:"nickName" json:"nickName"` // AI模型昵称
	Avatar   string `yaml:"avatar" json:"avatar"`     // AI模型头像URL
	Platform string `yaml:"platform" json:"platform"` // AI模型平台TODO:后期适配
}
