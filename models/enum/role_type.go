package enum

type RoleType int8

const (
	AdminRole    RoleType = 1 //管理员
	SuperVipRole RoleType = 2 //超级会员
	VipRole      RoleType = 3 //会员
	UserRole     RoleType = 4 //用户
	GuestRole    RoleType = 5 //访客
	BlackRole    RoleType = 6 //封禁用户
)
