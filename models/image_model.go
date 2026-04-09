// models/image_model.go
// 图片模型定义
// 管理上传的图片文件信息，支持文件去重和路径管理
package models

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

//TODO:改造一下,允许视频类型的存储?

// ImageModel 图片模型
// 存储上传图片的元数据信息，包括文件名、路径、大小和哈希值
// 通过哈希值实现文件去重，避免重复存储相同文件
type ImageModel struct {
	Model
	Filename string `gorm:"size:64" json:"filename"`             // 文件名，后续可根据需要扩展长度
	Path     string `gorm:"size:256;index:idx_path" json:"path"` // 文件存储路径，添加索引优化查询
	Size     int64  `json:"size"`                                // 文件大小（字节）
	Hash     string `gorm:"size:32" json:"hash"`                 // 文件哈希值，用于去重检测
}

// WebPath 生成图片Web访问路径
// 参数: 无
// 返回: string - 图片的Web访问路径
// 说明: TODO:根据实际情况修改路径格式，目前只是占位实现
func (i ImageModel) WebPath() string {
	return fmt.Sprintf("/%s", i.Path)
}

// BeforeDelete 删除前的钩子函数
// 参数: tx - GORM数据库事务
// 返回: error - 删除过程中的错误
// 说明: 在数据库记录删除时，同步删除物理文件
func (l ImageModel) BeforeDelete(tx *gorm.DB) error {
	return os.Remove(l.Path)
}
