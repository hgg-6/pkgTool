package dao

import "database/sql"

// Department 部门表
type Department struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"`
	DeptId      int64          `gorm:"unique"`
	Name        string         `gorm:"column:name;type:varchar(128);size:128;not null"`
	ParentId    int64          `gorm:"column:parent_id;default:0"`
	Description sql.NullString `gorm:"column:description;type:varchar(512);size:512"`
	Ctime       float64        `gorm:"column:ctime"`
	Utime       float64        `gorm:"column:utime"`
}

func (Department) TableName() string {
	return "departments"
}

// User 用户表
type User struct {
	ID       int64          `gorm:"primaryKey;autoIncrement"`
	UserId   int64          `gorm:"unique"`
	Username string         `gorm:"column:username;type:varchar(64);size:64;not null;uniqueIndex"`
	Password string         `gorm:"column:password;type:varchar(256);size:256;not null"`
	Email    sql.NullString `gorm:"column:email;type:varchar(128);size:128"`
	Phone    sql.NullString `gorm:"column:phone;type:varchar(32);size:32"`
	DeptId   int64          `gorm:"column:dept_id;index"`
	Status   string         `gorm:"column:status;type:varchar(32);size:32;default:'active'"`
	Ctime    float64        `gorm:"column:ctime"`
	Utime    float64        `gorm:"column:utime"`
}

func (User) TableName() string {
	return "users"
}

// Role 角色表
type Role struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"`
	RoleId      int64          `gorm:"unique"`
	Name        string         `gorm:"column:name;type:varchar(64);size:64;not null"`
	Code        string         `gorm:"column:code;type:varchar(64);size:64;not null;uniqueIndex"`
	Description sql.NullString `gorm:"column:description;type:varchar(512);size:512"`
	Ctime       float64        `gorm:"column:ctime"`
	Utime       float64        `gorm:"column:utime"`
}

func (Role) TableName() string {
	return "roles"
}

// Permission 权限表
type Permission struct {
	ID          int64          `gorm:"primaryKey;autoIncrement"`
	PermId      int64          `gorm:"unique"`
	Name        string         `gorm:"column:name;type:varchar(64);size:64;not null"`
	Code        string         `gorm:"column:code;type:varchar(128);size:128;not null;uniqueIndex"`
	Resource    string         `gorm:"column:resource;type:varchar(64);size:64"`
	Action      string         `gorm:"column:action;type:varchar(32);size:32"`
	Description sql.NullString `gorm:"column:description;type:varchar(512);size:512"`
	Ctime       float64        `gorm:"column:ctime"`
	Utime       float64        `gorm:"column:utime"`
}

func (Permission) TableName() string {
	return "permissions"
}

// UserRole 用户-角色关联表
type UserRole struct {
	ID     int64   `gorm:"primaryKey;autoIncrement"`
	UserId int64   `gorm:"column:user_id;index"`
	RoleId int64   `gorm:"column:role_id;index"`
	Ctime  float64 `gorm:"column:ctime"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission 角色-权限关联表
type RolePermission struct {
	ID     int64   `gorm:"primaryKey;autoIncrement"`
	RoleId int64   `gorm:"column:role_id;index"`
	PermId int64   `gorm:"column:perm_id;index"`
	Ctime  float64 `gorm:"column:ctime"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// CronPermission 任务权限控制表
type CronPermission struct {
	ID     int64   `gorm:"primaryKey;autoIncrement"`
	CronId int64   `gorm:"column:cron_id;index"`
	DeptId int64   `gorm:"column:dept_id;index"`
	Ctime  float64 `gorm:"column:ctime"`
}

func (CronPermission) TableName() string {
	return "cron_permissions"
}
