package email_store

import (
	"sync"
)

// 已迁移至Redis验证和存储,以下是原来的内存存储实现,现在保留在这里以供参考,但不再使用
type EmailStoreInfo struct {
	Email string
	Code  string
}

var EmailVerifyStore = sync.Map{}

func Set(id, email, code string) {
	EmailVerifyStore.Store(id, EmailStoreInfo{
		Email: email,
		Code:  code,
	})
}

func Verify(id, code string) (EmailStoreInfo, bool) {
	var result EmailStoreInfo
	var ok bool
	value, ok := EmailVerifyStore.Load(id)
	if !ok {
		return EmailStoreInfo{}, false
	}
	result, ok = value.(EmailStoreInfo)
	if !ok {
		return EmailStoreInfo{}, false
	}
	if result.Code != code {
		EmailVerifyStore.Delete(id)
		return EmailStoreInfo{}, false
	}
	EmailVerifyStore.Delete(id)
	return result, true
}
