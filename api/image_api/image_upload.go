// api/image_api/image_upload.go
package image_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/utils"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// func (ImageApi) ImageUploadView(c *gin.Context) { //TODO:存储图片的方式要改一下,图片的路径应该是生成三段哈希,实现文件的三层文件夹存储
// 	fileHeader, err := c.FormFile("file")
// 	response.OkWithMsg("连接成功", c)
// 	if err != nil {
// 		response.FailWithMsg("参数错误", c)
// 		return
// 	}

// 	filePath := filepath.Join(global.Config.Upload.UploadDir, fileHeader.Filename)
// 	c.SaveUploadedFile(fileHeader, filePath)
// 	response.Ok(filePath, "图片上传成功", c)

// }

// api/image_api/image_upload.go

func (ImageApi) ImageUploadView(c *gin.Context) {
	//前端看这里,图片上传完成后会返回一个图片的ID,这个ID可以用来访问图片,访问图片的接口是 /api/image/:id
	//所以图文就是,前端上传图片,后端返回一个ID,前端拿到这个ID,就可以通过把路径替换为 /api/image/:id 来实现图文博客功能
	//TODO:修复完成,迟点再修饰一下,加入配置路径等功能
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	// 文件大小判断
	s := global.Config.Upload.Size
	if fileHeader.Size > s*1024*1024 {
		response.FailWithMsg(fmt.Sprintf("文件大小大于%dMB", s), c)
		return
	}
	// 后缀判断
	filename := fileHeader.Filename
	suffix, ok := imageSuffixJudge(filename)
	if ok {
		response.FailWithMsg("文件名非法", c)
		return
	}
	// 文件hash
	file, err := fileHeader.Open()
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	byteData, _ := io.ReadAll(file)
	hash := utils.Md5(byteData)
	// 判断这个hash有没有
	var model models.ImageModel
	err = global.DB.Take(&model, "hash = ?", hash).Error
	if err == nil {
		// 找到了
		logrus.Infof("上传图片重复 %s = %s  %s", filename, model.Filename, hash)
		response.Ok(model.ID, "上传成功", c)
		return
	}
	// 文件名称一样，但是文件内容不一样
	filePath := fmt.Sprintf("uploads/%s/%s.%s", global.Config.Upload.UploadDir, hash, suffix)
	// 入库
	model = models.ImageModel{
		Filename: filename,
		Path:     filePath,
		Size:     fileHeader.Size,
		Hash:     hash,
	}
	err = global.DB.Create(&model).Error
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	c.SaveUploadedFile(fileHeader, filePath)
	response.Ok(model.ID, "图片上传成功", c)
}

func imageSuffixJudge(filename string) (string, bool) {
	_list := strings.Split(filename, ".")
	var suffix string
	if len(_list) == 1 {

		return suffix, false
	}
	// xxx.jpg   xxx  xxx.jpg.exe
	suffix = _list[len(_list)-1]
	if !utils.InList(suffix, global.Config.Upload.WhiteList) {
		return suffix, false
	}
	return suffix, true
}
