package models

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserFollowModel struct { //以后改名为follow表
	Model
	UserID         uint      `json:"userID"` // 用户id
	UserModel      UserModel `gorm:"foreignKey:UserID" json:"-"`
	FocusUserID    uint      `json:"focusUserID"` // 关注的其他用户的ID
	FocusUserModel UserModel `gorm:"foreignKey:FocusUserID" json:"-"`
	Friend         bool      `json:"friend"` // 是否是好友,初始创建时一定得是false
}

// 注:当两个用户同时关注对方时就会导致双方只能建立关注关系而无法成为好友关系,这是高并发下出现的问题
// 因为此时对方的记录还未写入,但是检查已经执行,执行之后的结构就是没有对方的记录,所以只能建立关注关系
// 而且有死锁的风险
// AfterCreate 在创建关注记录后，检查是否形成双向关注，并更新好友状态
func (self *UserFollowModel) AfterCreate(DB *gorm.DB) error {
	if self.UserID == 0 || self.FocusUserID == 0 {
		return nil
	}

	// 创建一个带超时的上下文，用于控制数据库操作的最长时间
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	// 使用 defer 确保在函数结束时释放上下文资源
	defer cancel()

	// 在一个新的 goroutine 中异步处理好友关系建立逻辑
	// 这样不会阻塞主创建流程
	go func(ctx context.Context) {
		// 1. 检查上下文是否已取消（例如超时）
		select {
		case <-ctx.Done():
			return // 上下文已完成，直接返回
		default:
			// 上下文未完成，继续执行业务逻辑
		}
		// 查询是否存在反向关注，即对方也关注了自己
		var oppositeRelation UserFollowModel
		err := DB.WithContext(ctx).Where(&UserFollowModel{
			UserID:      self.FocusUserID, // 对方的UserID
			FocusUserID: self.UserID,      // 对方关注的是我
		}).First(&oppositeRelation).Error
		if err != nil {
			// 如果没找到反向关注记录，则说明还不是好友，仅记录日志并返回
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// log.Printf("No mutual focus found, not becoming friends: user %d and user %d", self.UserID, self.FocusUserID)
				return
			}
			logrus.Errorf("查询反向关注记录失败: %v", err)
			return
		}
		// 找到了反向关注，说明已经形成互相关注，需要更新双方的 friend 字段为 true
		// 启动一个事务来保证两个更新操作的原子性
		err = DB.Transaction(func(tx *gorm.DB) error {
			// 更新当前记录（self）的 friend 状态
			if err := tx.Model(self).Update("friend", true).Error; err != nil {
				return err // 返回错误以回滚事务
			}
			// 更新对方记录（oppositeRelation）的 friend 状态
			if err := tx.Model(&oppositeRelation).Update("friend", true).Error; err != nil {
				return err // 返回错误以回滚事务
			}
			return nil // 事务成功提交
		})
		if err != nil {
			// 如果事务失败，记录错误
			logrus.Errorf("好友关系更新失败: %v", err)
			return
		}
		// 成功建立好友关系
	}(ctx) // 将上下文传递给 goroutine
	return nil
}

// BeforeDelete 在删除关注记录前，解除可能存在的好友关系
func (self *UserFollowModel) BeforeDelete(DB *gorm.DB) error {
	if self.UserID == 0 || self.FocusUserID == 0 {
		return nil
	}

	// 创建一个带超时的上下文，用于控制数据库操作的最长时间
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	// 使用 defer 确保在函数结束时释放上下文资源
	defer cancel()

	// 在一个新的 goroutine 中异步处理好友关系解除逻辑
	// 这样不会阻塞主删除流程
	go func(ctx context.Context) {
		// 1. 检查上下文是否已取消（例如超时）
		select {
		case <-ctx.Done():
			return // 上下文已完成，直接返回
		default:
			// 上下文未完成，继续执行业务逻辑
		}

		// 查询是否存在反向关注，即对方也关注了自己
		var oppositeRelation UserFollowModel
		err := DB.WithContext(ctx).Where(&UserFollowModel{
			UserID:      self.FocusUserID, // 对方的UserID
			FocusUserID: self.UserID,      // 对方关注的是我
		}).First(&oppositeRelation).Error
		if err != nil {
			// 如果没找到反向关注记录，则说明当前删除操作不会破坏好友关系，仅记录日志并返回
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// log.Printf("No mutual focus to break, deletion has no effect on friendship: user %d and user %d", self.UserID, self.FocusUserID)
				return
			}
			// 如果是其他查询错误，则记录
			logrus.Errorf("查询反向关注关系时出错: %v", err)
			return
		}

		// 找到了反向关注，说明删除当前记录将破坏互相关注，需要解除双方的好友关系。
		// 注意：在此钩子中，我们只负责将对方记录的 friend 状态置为 false。
		// 当前记录（self）即将被删除，其 friend 状态无需再更新。
		if err := DB.WithContext(ctx).Model(&oppositeRelation).Update("friend", false).Error; err != nil {
			// 如果更新失败，记录错误
			logrus.Errorf("解除好友关系失败 (user %d -> user %d): %v", oppositeRelation.UserID, oppositeRelation.FocusUserID, err)
			return
		}
		// 成功解除好友关系
	}(ctx) // 将上下文传递给 goroutine

	return nil
}
