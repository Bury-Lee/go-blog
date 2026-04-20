// api/enter.go
package api

import (
	"StarDreamerCyberNook/api/OSS_img_api"
	"StarDreamerCyberNook/api/ai_api"
	"StarDreamerCyberNook/api/article_api"
	"StarDreamerCyberNook/api/banner_api"
	"StarDreamerCyberNook/api/captcha_api"
	"StarDreamerCyberNook/api/chat_api"
	"StarDreamerCyberNook/api/comment_api"
	"StarDreamerCyberNook/api/follow_api"
	friendlink_and_friendpromote "StarDreamerCyberNook/api/friendLink_and_friendPromote"
	"StarDreamerCyberNook/api/hearth_api"
	"StarDreamerCyberNook/api/image_api"
	"StarDreamerCyberNook/api/log_api"
	site_message_api "StarDreamerCyberNook/api/message_api"
	"StarDreamerCyberNook/api/site_api"
	"StarDreamerCyberNook/api/test_api"
	"StarDreamerCyberNook/api/user_api"
)

type Api struct { //在这里注册路由
	SiteApi        site_api.SiteApi
	LogApi         log_api.LogApi
	ImageApi       image_api.ImageApi
	BannerApi      banner_api.BannerApi
	FriendApi      friendlink_and_friendpromote.FriendApi
	CaptchaApi     captcha_api.CaptchaApi
	UserApi        user_api.UserApi
	ArticleApi     article_api.ArticleApi
	CommentApi     comment_api.CommentApi
	SiteMessageApi site_message_api.MessageApi
	FollowApi      follow_api.FollowApi
	ChatApi        chat_api.ChatApi
	AIApi          ai_api.AIApi
	HearthApi      hearth_api.HearthApi  //心跳接口
	OSSImgApi      OSS_img_api.OSSImgApi //OSS图片接口

	TestApi test_api.TestApi
}

var App = Api{} //全局Api对象，包含所有的Api接口
