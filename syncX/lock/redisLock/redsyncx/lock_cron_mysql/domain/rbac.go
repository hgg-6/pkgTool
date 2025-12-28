package domain

// Department 部门
type Department struct {
	ID          int64   `json:"id"`
	DeptId      int64   `json:"dept_id"`
	Name        string  `json:"name"`
	ParentId    int64   `json:"parent_id"` // 父部门ID，支持部门层级
	Description string  `json:"description"`
	Ctime       float64 `json:"ctime"`
	Utime       float64 `json:"utime"`
}

// User 用户
type User struct {
	ID       int64   `json:"id"`
	UserId   int64   `json:"user_id"`
	Username string  `json:"username"`
	Password string  `json:"password"` // 加密后的密码
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	DeptId   int64   `json:"dept_id"` // 所属部门
	Status   string  `json:"status"`  // active, inactive, locked
	Ctime    float64 `json:"ctime"`
	Utime    float64 `json:"utime"`
}

// Role 角色
type Role struct {
	ID          int64   `json:"id"`
	RoleId      int64   `json:"role_id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"` // 角色代码，如 admin, operator, viewer
	Description string  `json:"description"`
	Ctime       float64 `json:"ctime"`
	Utime       float64 `json:"utime"`
}

// Permission 权限
type Permission struct {
	ID          int64   `json:"id"`
	PermId      int64   `json:"perm_id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`     // 权限代码，如 cron:create, cron:delete
	Resource    string  `json:"resource"` // 资源类型，如 cron, department, user
	Action      string  `json:"action"`   // 操作，如 create, read, update, delete
	Description string  `json:"description"`
	Ctime       float64 `json:"ctime"`
	Utime       float64 `json:"utime"`
}

// UserRole 用户-角色关联
type UserRole struct {
	ID     int64   `json:"id"`
	UserId int64   `json:"user_id"`
	RoleId int64   `json:"role_id"`
	Ctime  float64 `json:"ctime"`
}

// RolePermission 角色-权限关联
type RolePermission struct {
	ID     int64   `json:"id"`
	RoleId int64   `json:"role_id"`
	PermId int64   `json:"perm_id"`
	Ctime  float64 `json:"ctime"`
}

// CronPermission 任务权限控制
type CronPermission struct {
	ID     int64   `json:"id"`
	CronId int64   `json:"cron_id"`
	DeptId int64   `json:"dept_id"` // 哪个部门可以管理此任务
	Ctime  float64 `json:"ctime"`
}
