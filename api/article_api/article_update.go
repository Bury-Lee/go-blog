package article_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"
	xss_filter "StarDreamerCyberNook/utils/XSSfilter"
	jwts "StarDreamerCyberNook/utils/jwts"
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ArticleUpdateRequest struct {
	ID          uint     `json:"id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Abstract    string   `json:"abstract"` //要考虑一个问题,如果用户想设置为空简介,那么就设为"该文章未设置简介"
	Content     string   `json:"content" binding:"required"`
	CategoryID  uint     `json:"categoryID"`
	TagList     []string `json:"tagList"`
	Cover       string   `json:"cover"`
	OpenComment bool     `json:"openComment"`
}

func (ArticleApi) ArticleUpdateView(c *gin.Context) {
	var req ArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg("参数错误", c)
		return
	}

	user, err := jwts.GetClaims(c).GetUser()
	if err != nil {
		response.FailWithMsg("用户不存在", c)
		return
	}

	var article models.ArticleModel
	err = global.DB.Take(&article, req.ID).Error
	if err != nil {
		response.FailWithMsg("文章不存在", c)
		return
	}

	// 更新的文章必须是自己的
	if article.UserID != user.ID {
		response.FailWithMsg("只能更新自己的文章", c)
		return
	}

	// 判断分类id是不是自己创建的
	var category models.CategoryModel
	if req.CategoryID != 0 {
		err = global.DB.Take(&category, "id = ? and user_id = ?", req.CategoryID, user.ID).Error
		if err != nil {
			response.FailWithMsg("文章分类不存在", c)
			return
		}
	}

	// 文章正文防xss注入
	xssFilter := xss_filter.NewXSSFilter()
	if req.Content == "" {
		response.FailWithMsg("正文解析错误", c)
		return
	} else {
		req.Content = xssFilter.Sanitize(req.Content)
	}
	//不传简介时就设为无,传了就做清洗
	if req.Abstract != "" {
		req.Abstract = xssFilter.Sanitize(req.Abstract)
	} else {
		req.Abstract = "该文章未设置简介"
	}

	mps := map[string]any{
		"title":        req.Title,
		"abstract":     req.Abstract,
		"content":      req.Content,
		"category_id":  req.CategoryID,
		"tag_list":     req.TagList,
		"cover":        req.Cover,
		"open_comment": req.OpenComment,
	}
	if article.Status == models.StatusPublished && !global.Config.Site.Article.DisableExamination {
		// 如果是已发布的文章，进行编辑，那么就要改成待审核
		mps["status"] = models.StatusPending
	}

	if global.Config.AI.Enable { //如果启用ai审核
		ctx := context.Background()

		// 构建完整的消息列表
		var messages []openai.ChatCompletionMessage

		// 添加系统提示词作为第一条消息

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: global.SystemPromptArticle.String(),
		})

		// 添加对话历史消息
		messages = append(messages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "文章标题:" + req.Title + "\n文章摘要:" + req.Abstract + "\n文章内容:" + req.Content,
			})

		// 创建非流式请求
		res, err := global.LocalAIClient.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    global.Config.AI.Model,
				Messages: messages,
			},
		)
		if err != nil {
			response.FailWithMsg("ai审核失败,已经自动创建为待审核状态: "+err.Error(), c)
			return
		}
		switch res.Choices[0].Message.Content { //TODO:这里无论成功还是失败都应该插入消息,告知原因
		case "通过":
			mps["status"] = models.StatusPublished
		case "拒绝":
			mps["status"] = models.StatusDraft
		default:
			logrus.Errorf("ai审核出错,回复内容:%s,文章详情:%s\n已自动换为待审核状态", res.Choices[0].Message.Content, fmt.Sprintf("%#v", req))
			mps["status"] = models.StatusPending
		}
	}

	err = global.DB.Model(&article).Updates(mps).Error
	if err != nil {
		response.FailWithMsg("更新失败", c)
		return
	}

	//查询Redis是否存在该文章缓存

	//redis中存在就更新的策略
	// idStr := strconv.FormatUint(uint64(article.ID), 10)
	// _, err = global.RedisHotPool.Get("ArticleID" + idStr).Result()
	// if err == nil {
	// 	articleJSON, err := json.Marshal(&article)
	// 	if err != nil {
	// 		logrus.Error("文章创建失败,缓存数据解析错误: " + err.Error())
	// 		return
	// 	}
	// 	global.RedisHotPool.Set("ArticleID"+idStr, articleJSON, 0)
	// }

	//redis中存在就删除的策略
	idStr := strconv.FormatUint(uint64(article.ID), 10)
	ctx := context.Background()
	global.RedisHotPool.Del(ctx, "ArticleID"+idStr)
	response.OkWithMsg("文章更新成功,当前状态为:"+article.Status.String(), c)
}
