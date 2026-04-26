package conf

// Site 站点配置
type Site struct {
	SiteInfo   SiteInfo   `yaml:"siteInfo" json:"siteInfo"`     // 站点信息
	Project    Project    `yaml:"project" json:"project"`       // 项目
	Seo        Seo        `yaml:"seo" json:"seo"`               // 搜索引擎优化
	About      About      `yaml:"about" json:"about"`           // 关于
	IndexRight IndexRight `yaml:"indexRight" json:"indexRight"` // 右侧组件
	Article    Article    `yaml:"article" json:"article"`       // 文章设置
	Login      Login      `yaml:"login" json:"login"`           // 登录
}

// SiteInfo 站点信息
type SiteInfo struct {
	Title string `yaml:"title" json:"title"` // 站点标题
	Logo  string `yaml:"logo"`               // 站点logo
	Beian string `yaml:"beian"`              // 站点备案号
	Mode  int8   `yaml:"mode"`               // 站点运行模式
}

// Project 项目
type Project struct {
	Title   string `yaml:"title" json:"title"`     // 项目名称
	Icon    string `yaml:"icon" json:"icon"`       // 项目图标
	WebPath string `yaml:"webPath" json:"webPath"` // 项目访问路径
}

// Seo 搜索引擎优化
type Seo struct {
	Keywords    string `yaml:"keywords" json:"keywords"`       // 站点关键词
	Description string `yaml:"description" json:"description"` // 站点描述
}

// About 关于
type About struct {
	Version  string // 版本号 (硬编码)
	SiteDate string `yaml:"siteDate" json:"siteDate"` // 站点建立时间
	QQ       string `yaml:"qq" json:"qq"`             // QQ群
	Wechat   string `yaml:"wechat" json:"wechat"`     // 微信
	BiliBili string `yaml:"biliBili" json:"biliBili"` // 哔哩哔哩
	GitHub   string `yaml:"gitHub" json:"gitHub"`     // GitHub
}

func (this *About) SetVersion() {
	this.Version = Version
}

// Article 文章设置
type Article struct {
	EnableExamination bool `yaml:"enableExamination" json:"enableExamination"` // 是否启用文章审核
}

// Login 登录
type Login struct {
	QQLogin          bool `yaml:"QQLogin" json:"QQLogin"`                   // 启用QQ登录
	UsernamePassword bool `yaml:"usernamePassword" json:"usernamePassword"` // 启用用户名密码登录
	EmailLogin       bool `yaml:"emailLogin" json:"emailLogin"`             // 启用邮箱登录
	Captcha          bool `yaml:"captcha" json:"captcha"`                   // 启用验证码登录
}

// IndexRight 右侧组件
type IndexRight struct {
	List []Component `yaml:"list" json:"list"` // 组件列表
}

// Component 组件
type Component struct {
	Title  string `yaml:"title" json:"title"`   // 组件标题
	Enable bool   `yaml:"enable" json:"enable"` // 是否启用
}
