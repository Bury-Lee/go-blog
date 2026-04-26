package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/models/enum"
	"StarDreamerCyberNook/service/ai_service"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ArticleCreateRequest struct {
	Title       string        `json:"title" binding:"required"`   // 文章标题，最大32字符
	Abstract    string        `json:"abstract"`                   // 文章摘要，最大256字符
	Content     string        `json:"content" binding:"required"` // 文章内容
	CategoryID  *uint         `json:"categoryID"`                 // 文章分类ID，关联分类表
	TagList     []string      `json:"tagList"`                    // 标签列表，JSON序列化存储 //serializer:json要删掉?似乎要换成自己定义的taglist数据类型
	Cover       string        `json:"cover"`                      // 文章封面图片URL
	OpenComment bool          `json:"openComment"`                // 是否开启评论：true-开启 false-关闭
	Stats       models.Status `json:"status"`                     //状态,普通用户只能设置为草稿或者审核中,管理员可设置为任意值
}

func (ArticleApi) ArticleCreateView(c *gin.Context) {
	//注:刚创建的文章不可能成为热门文章,所以不需要在这里使用Redis
	var req ArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("文章参数绑定失败: "+err.Error(), c)
		return
	}

	User, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("获取用户信息失败", c)
		return
	}
	if global.Config.Site.SiteInfo.Mode == 2 && User.Role != enum.AdminRole {
		response.FailWithMsg("未开放文章创建", c)
		return
	}
	if (req.Stats != 1 && req.Stats != 0) && User.Role != enum.AdminRole {
		response.FailWithMsg("非法参数", c)
		return
	}
	if len(req.Abstract) > 200 { //避免在搜索时刷屏
		response.FailWithMsg("简介过长", c)
		return
	}

	if !global.Config.Site.Article.EnableExamination && req.Stats == models.StatusPending {
		req.Stats = models.StatusPublished
		//未启用审核且设置为审核中状态时跳过审核
	}

	//判断分类id是否为自己创建
	var category models.CategoryModel
	if req.CategoryID != nil && *req.CategoryID != 0 {
		err := global.DB.Take(&category, "id = ? and user_id = ?", *req.CategoryID, User.ID).Error
		if err != nil {
			response.FailWithMsg("分类不存在", c)
			return
		}
	}

	//防xss注入
	xssFilter := xss_filter.NewXSSFilter()
	req.Content = xssFilter.Sanitize(req.Content)
	if req.Content == "" {
		response.FailWithMsg("正文解析错误", c)
		return
	}
	//不传简介时就设为无,传了就做清洗
	if req.Abstract != "" {
		xssFilter := xss_filter.NewXSSFilter()
		req.Abstract = xssFilter.Sanitize(req.Abstract)
	} else {
		req.Abstract = "该文章未设置简介"
	}

	// 正文内容图片转存
	// 但是吧,如果图片过多时，同步做，接口耗时高,异步做又很麻烦...但是以后肯定要考虑变成异步的

	if global.Config.AI.Enable && global.Config.Site.Article.EnableExamination { //启用ai审核
		reply, err := ai_service.CreateSingleReply(
			"文章标题:"+req.Title+"\n文章摘要:"+req.Abstract+"\n文章内容:"+req.Content,
			global.SystemPromptArticleReview.String(),
		)
		if err != nil {
			logrus.Error("ai审核失败:" + err.Error())
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态", c)
			return
		}
		switch reply { //TODO:这里无论成功还是失败都应该插入消息,告知原因
		//注:一般我们认为ai审核是很迅速的,可以在3秒内看到结果,所以不考虑发送消息通知
		case "通过":
			req.Stats = models.StatusPublished
		case "拒绝":
			req.Stats = models.StatusDraft
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,文章详情:%s\n已自动换为待审核状态", reply, fmt.Sprintf("%#v", req))
			req.Stats = models.StatusPending
		}
	}

	// 构建模型实例
	var article = models.ArticleModel{
		Title:       req.Title,       // 文章标题
		UserID:      User.ID,         // 用户ID
		Abstract:    req.Abstract,    // 文章摘要
		Content:     req.Content,     // 文章内容
		CategoryID:  req.CategoryID,  // 分类ID
		TagList:     req.TagList,     // 标签列表 (确保模型层字段类型兼容 []string 或已配置 GORM serializer)
		Cover:       req.Cover,       // 封面图
		OpenComment: req.OpenComment, // 是否开启评论
		Status:      req.Stats,
	}

	//追加ai摘要和ai评级
	if global.Config.AI.Enable {
		{ //ai摘要
			// 构建完整的消息列表
			reply, err := ai_service.CreateSingleReply(
				"文章标题:"+req.Title+"\n文章摘要:"+req.Abstract+"\n文章内容:"+req.Content,
				global.SystemPromptArticleAbstract.String(),
			)
			if err != nil {
				logrus.Errorf("ai自动创建摘要和评级失败: %s", err.Error())
			} else {
				article.AIAbstract = reply
			}
		}
		{ //ai评级
			// 构建完整的消息列表
			reply, err := ai_service.CreateSingleReply(
				"文章标题:"+req.Title+"\n文章摘要:"+req.Abstract+"\n文章内容:"+req.Content,
				global.SystemPromptArticleAiQuality.String(),
			)
			if err != nil {
				logrus.Errorf("ai自动创建摘要和评级失败: %s", err.Error())
			} else {
				article.AIQuality = reply
			}
		}
	}
	if err = global.DB.Create(&article).Error; err != nil {
		response.FailWithMsg("文章创建失败", c)
		return
	}

	response.OkWithMsg(fmt.Sprintf("文章创建成功,当前状态:%s", req.Stats.String()), c)
}
