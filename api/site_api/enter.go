// api/site_api/enter.go
package site_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/conf"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/middleware"
	"StarDreamerCyberNook/models/enum"
	jwts "StarDreamerCyberNook/utils/jwts"
	"errors"
	"fmt"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SiteApi struct{}

type SiteInfo struct {
	Name string `uri:"name"`
}

func (this *SiteApi) SiteInfoView(c *gin.Context) {
	var siteInfo SiteInfo
	if err := c.ShouldBindUri(&siteInfo); err != nil {
		response.FailWithError(err, c)
		return
	}
	if siteInfo.Name == "site" { //基本服务允许使用
		response.OkWithData(global.Config.Site, c)
		return
	}
	claims, err := jwts.ParseTokenByGin(c)
	if err != nil {
		response.FailWithError(err, c)
		return
	}
	//TODO:和配置有关的信息再单独开个管理员的路由吧
	middleware.AdminMiddleware(c) //检查是否为管理员同时检查黑名单
	if claims.Role == enum.AdminRole {
		var data any
		switch siteInfo.Name { //以下需要管理员权限
		case "email":
			res := global.Config.Email
			res.AuthCode = "******"
			data = res
		case "qq":
			res := global.Config.QQ
			res.AppKey = "******"
			data = res
		case "ai":
			res := global.Config.AI
			res.ApiKey = "******"
			data = res
		default:
			data = "错误的站点名称"
		}
		response.OkWithData(data, c)
	} else {
		response.FailWithMsg("权限不足", c)
	}
}

func (SiteApi) SiteInfoQQView(c *gin.Context) { //QQ登录的接口
	response.OkWithData(global.Config.QQ.Url(), c)
}

type SiteUpdate struct {
	Name string `uri:"name" binding:"required"`
	// Value string `uri:"value" binding:"required"`//似乎用不上
}

// 路由注册需要改为：PUT /site/update/:name/:value

func (SiteApi) SiteUpdateView(c *gin.Context) {
	var req SiteUpdate
	err := c.ShouldBindUri(&req)
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	var rep any
	switch req.Name {
	case "site":
		var data conf.Site
		err = c.ShouldBindJSON(&data)
		rep = data
	case "email":
		var data conf.Email
		err = c.ShouldBindJSON(&data)
		rep = data
	case "qq":
		var data conf.QQ
		err = c.ShouldBindJSON(&data)
		rep = data
	case "qiNiu":
		var data conf.QiNiu
		err = c.ShouldBindJSON(&data)
		rep = data
	case "ai":
		var data conf.AI
		err = c.ShouldBindJSON(&data)
		rep = data
	default:
		response.FailWithMsg("不存在的配置", c)
		return
	}
	if err != nil {
		response.FailWithError(err, c)
		return
	}

	switch s := rep.(type) {
	case conf.Site:
		// 判断站点信息更新前端文件部分
		err = UpdateSite(s)
		if err != nil {
			response.FailWithError(err, c)
			return
		}
		global.Config.Site = s
	case conf.Email:
		if s.AuthCode == "******" {
			s.AuthCode = global.Config.Email.AuthCode
		}
		global.Config.Email = s
	case conf.QQ:
		if s.AppKey == "******" {
			s.AppKey = global.Config.QQ.AppKey
		}
		global.Config.QQ = s
	case conf.QiNiu:
		if s.SecretKey == "******" {
			s.SecretKey = global.Config.QiNiu.SecretKey
		}
		global.Config.QiNiu = s
	case conf.AI:
		if s.ApiKey == "******" {
			s.ApiKey = global.Config.AI.ApiKey
		}
		global.Config.AI = s
	}

	// 改配置文件
	// core.SetConf()//保存配置文件应该单独做一个路由

	response.OkWithMsg("更新站点配置成功", c)
}

/*
<!DOCTYPE html>
<html lang="zh-CN">
<head>

	<!-- 必须包含这些元素，程序会查找并修改它们 -->
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<!-- 1. title 标签 - 必须存在，程序会修改其文本 -->
	<title>默认标题</title>

	<!-- 2. favicon 图标 - 可以存在或不存在，程序会修改或创建 -->
	<link rel="icon" href="/favicon.ico">

	<!-- 3. SEO keywords - 可以存在或不存在，程序会修改或创建 -->
	<meta name="keywords" content="默认关键词">

	<!-- 4. SEO description - 可以存在或不存在，程序会修改或创建 -->
	<meta name="description" content="默认描述">

	<!-- 其他 head 内容... -->
	<link rel="stylesheet" href="/css/style.css">

</head>
<body>

	<div id="app"></div>
	<script src="/js/app.js"></script>

</body>
</html>
*/
func UpdateSite(site conf.Site) error { //更新站点配置文件,部分设置在保存时要验证
	if site.Project.Icon == "" && site.Project.Title == "" &&
		site.Seo.Keywords == "" && site.Seo.Description == "" &&
		site.Project.WebPath == "" {
		return nil
	}

	if site.Project.WebPath == "" {
		return errors.New("请配置前端地址")
	}

	file, err := os.Open(site.Project.WebPath)
	defer func() {
		if file != nil { //没有打开文件,就不需要关闭
			if fileErr := file.Close(); fileErr != nil {
				logrus.Errorf("文件关闭失败 %s", fileErr)
			}
		}
	}() // 关闭文件

	if err != nil {
		return fmt.Errorf("文件不存在,%s", site.Project.WebPath)
	}

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		logrus.Errorf("goquery 解析失败 %s", err)
		return errors.New("文件解析失败")
	}

	if site.Project.Title != "" {
		doc.Find("title").SetText(site.Project.Title)
	}
	if site.Project.Icon != "" {
		selection := doc.Find("link[rel=\"icon\"]")
		if selection.Length() > 0 {
			selection.SetAttr("href", site.Project.Icon)
		} else {
			// 没有就创建
			doc.Find("head").AppendHtml(fmt.Sprintf("<link rel=\"icon\" href=\"%s\">", site.Project.Icon))
		}
	}
	if site.Seo.Keywords != "" {
		selection := doc.Find("meta[name=\"keywords\"]")
		if selection.Length() > 0 {
			selection.SetAttr("content", site.Seo.Keywords)
		} else {
			doc.Find("head").AppendHtml(fmt.Sprintf("<meta name=\"keywords\" content=\"%s\">", site.Seo.Keywords))
		}
	}
	if site.Seo.Description != "" {
		selection := doc.Find("meta[name=\"description\"]")
		if selection.Length() > 0 {
			selection.SetAttr("content", site.Seo.Description)
		} else {
			doc.Find("head").AppendHtml(fmt.Sprintf("<meta name=\"description\" content=\"%s\">", site.Seo.Description))
		}
	}

	html, err := doc.Html()
	if err != nil {
		logrus.Errorf("生成html失败 %s", err)
		return errors.New("生成html失败")
	}

	err = os.WriteFile(site.Project.WebPath, []byte(html), 0666)
	if err != nil {
		logrus.Errorf("文件写入失败 %s", err)
		return errors.New("文件写入失败")
	}
	return nil
}
