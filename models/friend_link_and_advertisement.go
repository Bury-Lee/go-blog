package models

type FriendLink struct {
	Model
	Name      string `gorm:"size:100;not null;" json:"name"`     // 名称
	URL       string `gorm:"size:255;" json:"url"`               // 链接地址
	Logo      string `gorm:"size:255;default:'';" json:"logo"`   // Logo 图片地址 (可选)
	IsShow    bool   `gorm:"default:true" json:"is_show"`        // 是否启用/显示
	SortOrder int    `gorm:"default:0" json:"sort_order"`        // 排序权重（数字越小越靠前）
	Remark    string `gorm:"size:2048;default:''" json:"remark"` // 简短描述或备注
}
type FriendPromotion struct { //友情推广
	Model
	// 基本信息
	Title      string `gorm:"size:100;not null;" json:"title"`      // 标题/业务名称
	FriendName string `gorm:"size:50;not null;" json:"friend_name"` // 朋友昵称/姓名
	Avatar     string `gorm:"size:500;" json:"avatar"`              // 头像（个人形象）

	// 业务信息
	Category      string `gorm:"size:30;" json:"category"`         // 业务类型: 画师/摄影/开发/设计/音乐/其他
	Description   string `gorm:"type:text;" json:"description"`    // 详细介绍（支持长文本）
	PreviewImages string `gorm:"type:text;" json:"preview_images"` // 作品预览图，JSON数组 ["url1","url2"]

	// 联系方式
	ContactInfo []string `gorm:"type:text;serializer:json" json:"contact_info"`

	// 展示控制
	IsShow    bool   `gorm:"default:true" json:"is_show"`            // 是否展示
	SortOrder int    `gorm:"default:0" json:"sort_order"`            // 排序
	Position  string `gorm:"size:20;default:'page'" json:"position"` // 位置: home(首页推荐)/page(友链页)/both

	// 备注
	Remark string `gorm:"size:255;" json:"remark"` // 备注（合作时间、关系等）
}
