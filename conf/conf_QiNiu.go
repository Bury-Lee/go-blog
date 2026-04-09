package conf

type QiNiu struct { //预留配置:七牛云
	Enable    bool   `yaml:"enable" json:"enable"`       // 是否启用七牛云
	AccessKey string `yaml:"accessKey" json:"accessKey"` // 七牛云AccessKey
	SecretKey string `yaml:"secretKey" json:"secretKey"` // 七牛云SecretKey
	Bucket    string `yaml:"bucket" json:"bucket"`       // 七牛云存储桶
	Uri       string `yaml:"uri" json:"uri"`             // 七牛云存储桶URI
	Region    string `yaml:"region" json:"region"`       // 七牛云存储桶区域
	Prefix    string `yaml:"prefix" json:"prefix"`       // 七牛云存储桶前缀
	Size      uint   `yaml:"size" json:"size"`           // 七牛云存储桶大小
}

//TODO:多级存储,使用缩略图->原图二级图片设置
