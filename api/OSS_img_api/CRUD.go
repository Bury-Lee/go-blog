package OSS_img_api

import (
	"StarDreamerCyberNook/common"
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/log_service"
	"StarDreamerCyberNook/utils"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

// ImageListResponse 图片列表响应结构体
// 说明: 包含图片模型和Web访问路径
type ImageListResponse struct {
	models.ImageModel
	WebPath string `json:"webPath"` // Web访问路径
}

// RemoveRequest 图片删除请求结构体
type RemoveRequest struct {
	IDlist []uint `json:"IDlist" binding:"required"` // 要删除的图片ID列表
}

func (OSSImgApi) ImageUploadView(c *gin.Context) {
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
	suffix, ok := utils.ImageSuffixJudge(filename)
	if !ok {
		response.FailWithMsg("文件名非法:"+filename, c)
		return
	}

	// 打开文件流并一次性读取到内存
	file, err := fileHeader.Open()
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	byteData, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	// 计算 hash
	hash := utils.Md5(byteData)

	// 判断这个 hash 有没有
	var model models.ImageModel
	err = global.DB.Take(&model, "hash = ?", hash).Error
	if err == nil {
		logrus.Infof("上传图片重复 %s = %s  %s", filename, model.Filename, hash)
		response.Ok(model.ID, "上传成功", c)
		return
	}

	// 上传到 MinIO
	objectName := fmt.Sprintf("%s/%s.%s", global.Config.Upload.UploadDir, hash, suffix)
	ctx := context.Background()

	_, err = global.StorageClient.PutObject(
		ctx,
		global.Config.ObjectStorage.Bucket,
		objectName,
		bytes.NewReader(byteData),
		int64(len(byteData)),
		minio.PutObjectOptions{
			ContentType: utils.GetContentType(suffix),
		},
	)
	if err != nil {
		response.FailWithError(fmt.Errorf("上传到 MinIO 失败: %v", err), c)
		return
	}

	// 入库
	model = models.ImageModel{
		Filename: filename,
		Path:     objectName,
		Size:     fileHeader.Size,
		Hash:     hash,
	}
	err = global.DB.Create(&model).Error
	if err != nil { // 入库失败,清理 MinIO 对象
		// 清理 MinIO 对象
		global.StorageClient.RemoveObject(ctx, global.Config.ObjectStorage.Bucket, objectName, minio.RemoveObjectOptions{})
		response.FailWithError(err, c)
		return
	}
	response.Ok(model.ID, "图片上传成功", c)
}

// GetImage 查询图片
// 参数: c - gin上下文
// 说明: 根据ID查询图片并返回对象存储中的文件内容
func (OSSImgApi) GetImage(c *gin.Context) {
	id := c.Query("id")

	var img models.ImageModel
	// 查询图片是否存在
	if global.DB.Take(&img, "id = ?", id).Error != nil {
		response.FailWithMsg("图片已被删除", c)
		return
	}

	// 从对象存储读取图片
	obj, err := global.StorageClient.GetObject(
		context.Background(),
		global.Config.ObjectStorage.Bucket,
		img.Path,
		minio.GetObjectOptions{},
	)
	if err != nil {
		response.FailWithError(fmt.Errorf("读取图片失败: %v", err), c)
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		response.FailWithMsg("图片已被删除", c)
		return
	}

	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size))
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, img.Filename))
	if _, err = io.Copy(c.Writer, obj); err != nil {
		logrus.Errorf("返回图片失败:%v", err)
	}
}

// ImageList 管理员查询图片列表
// 参数: c - gin上下文
// 说明: 分页查询图片列表,支持文件名模糊搜索
func (OSSImgApi) ImageList(c *gin.Context) {
	var req common.PageInfo
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	// 查询图片列表
	_list, count, err := common.ListQuery[models.ImageModel](models.ImageModel{}, common.Options{
		PageInfo: req,
		Likes:    []string{"filename"},
	})
	if err != nil {
		response.FailWithMsg("查询失败", c)
		return
	}

	// 构建响应数据
	var list = make([]ImageListResponse, 0)
	for _, model := range _list {
		list = append(list, ImageListResponse{
			ImageModel: model,
			WebPath:    model.WebPath(),
		})
	}
	response.OkWithList(list, count, c)
}

// ImageRemoveView 管理员批量删除图片
// 参数: c - gin上下文
// 说明: 根据ID列表删除对象存储文件并删除数据库记录
func (OSSImgApi) ImageRemoveView(c *gin.Context) {
	var req RemoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	// 记录操作日志
	log := log_service.GetLog(c)
	log.ShowRequest()
	log.ShowResponse()

	// 查询要删除的图片
	var list []models.ImageModel
	global.DB.Find(&list, "id IN ?", req.IDlist)
	if len(list) == 0 {
		response.OkWithMsg("图片删除成功,共删除0张", c)
		return
	}
	// 删除数据库记录
	err := global.DB.Delete(&list).Error
	if err != nil {
		logrus.Errorf("删除数据库记录失败:%s", err)
		response.FailWithMsg("删除失败", c)
		return
	}
	// 删除对象存储文件
	ctx := context.Background()
	for _, model := range list {
		err := global.StorageClient.RemoveObject(
			ctx,
			global.Config.ObjectStorage.Bucket,
			model.Path,
			minio.RemoveObjectOptions{},
		)
		if err != nil {
			logrus.Errorf("删除对象失败:%s", err)
		}
	}

	response.OkWithMsg(fmt.Sprintf("图片删除成功,共删除%d张", len(list)), c)
}
