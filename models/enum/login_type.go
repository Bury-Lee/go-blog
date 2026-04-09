package enum

type LoginType int8

const (
	UserPwdLoginType LoginType = 1 //账号密码登录
	QQLoginType      LoginType = 2 //QQ登录
	EmailLoginType   LoginType = 3 //邮箱登录
)
