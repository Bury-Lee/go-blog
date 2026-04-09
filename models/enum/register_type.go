package enum

type RegisterSource int8

const (
	RegisterEmail    RegisterSource = 1 //邮箱注册
	RegisterQQ       RegisterSource = 2 //QQ注册
	RegisterTerminal RegisterSource = 3 //命令行
)
