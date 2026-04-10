package flags

import (
	"StarDreamerCyberNook/global"
	"StarDreamerCyberNook/models"

	"github.com/sirupsen/logrus"
)

func FlagDB() { //数据库迁移
	err := global.DB.AutoMigrate(
		&models.UserModel{},
		&models.UserConfModel{},
		&models.ArticleDiggModel{},
		&models.ArticleModel{},
		&models.CategoryModel{},
		&models.CollectModel{},
		&models.UserArticleCollectModel{},
		&models.UserArticleHistoryModel{},
		&models.ImageModel{},
		&models.CommentModel{},
		&models.LogModel{},
		&models.BannerModel{},
		&models.GlobalNotificationModel{},
		&models.FriendLink{},
		&models.FriendPromotion{},
		&models.UserLoginModel{},
		&models.UserMessageConfModel{},
		&models.MessageModel{},
		&models.UserFollowModel{},
		&models.ChatModel{},
		&models.TextMsg{},
		&models.ImageMsg{},
		&models.MarkdownMsg{},
		&models.ChatMsg{},
		&models.UserChatActionModel{},
		&models.TextModel{},
		&models.UserTopArticleModel{},
		&models.CommentDiggModel{},
	)
	if err != nil {
		logrus.Errorf("数据库迁移失败 %s", err)
	} else {
		logrus.Info("数据库已迁移")
	}
}
