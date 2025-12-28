package dao

import (
	"context"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// DepartmentDb 部门数据访问接口
type DepartmentDb interface {
	Create(ctx context.Context, dept Department) error
	FindById(ctx context.Context, deptId int64) (Department, error)
	FindAll(ctx context.Context) ([]Department, error)
	FindByParentId(ctx context.Context, parentId int64) ([]Department, error)
	Update(ctx context.Context, dept Department) error
	Delete(ctx context.Context, deptId int64) error
}

type departmentDb struct {
	db *gorm.DB
}

// NewDepartmentDb 创建DepartmentDb实例
func NewDepartmentDb(db *gorm.DB) DepartmentDb {
	return &departmentDb{db: db}
}

func (d *departmentDb) Create(ctx context.Context, dept Department) error {
	dept.Ctime = float64(time.Now().Unix())
	dept.Utime = dept.Ctime
	err := d.db.WithContext(ctx).Create(&dept).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (d *departmentDb) FindById(ctx context.Context, deptId int64) (Department, error) {
	var dept Department
	err := d.db.WithContext(ctx).Where("dept_id = ?", deptId).First(&dept).Error
	if err == gorm.ErrRecordNotFound {
		return Department{}, ErrDataRecordNotFound
	}
	return dept, err
}

func (d *departmentDb) FindAll(ctx context.Context) ([]Department, error) {
	var depts []Department
	err := d.db.WithContext(ctx).Find(&depts).Error
	return depts, err
}

func (d *departmentDb) FindByParentId(ctx context.Context, parentId int64) ([]Department, error) {
	var depts []Department
	err := d.db.WithContext(ctx).Where("parent_id = ?", parentId).Find(&depts).Error
	return depts, err
}

func (d *departmentDb) Update(ctx context.Context, dept Department) error {
	dept.Utime = float64(time.Now().Unix())
	return d.db.WithContext(ctx).Model(&Department{}).Where("dept_id = ?", dept.DeptId).Updates(&dept).Error
}

func (d *departmentDb) Delete(ctx context.Context, deptId int64) error {
	return d.db.WithContext(ctx).Where("dept_id = ?", deptId).Delete(&Department{}).Error
}

// UserDb 用户数据访问接口
type UserDb interface {
	Create(ctx context.Context, user User) error
	FindById(ctx context.Context, userId int64) (User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByDeptId(ctx context.Context, deptId int64) ([]User, error)
	FindAll(ctx context.Context) ([]User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, userId int64) error
}

type userDb struct {
	db *gorm.DB
}

// NewUserDb 创建UserDb实例
func NewUserDb(db *gorm.DB) UserDb {
	return &userDb{db: db}
}

func (u *userDb) Create(ctx context.Context, user User) error {
	user.Ctime = float64(time.Now().Unix())
	user.Utime = user.Ctime
	err := u.db.WithContext(ctx).Create(&user).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (u *userDb) FindById(ctx context.Context, userId int64) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("user_id = ?", userId).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrDataRecordNotFound
	}
	return user, err
}

func (u *userDb) FindByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, ErrDataRecordNotFound
	}
	return user, err
}

func (u *userDb) FindByDeptId(ctx context.Context, deptId int64) ([]User, error) {
	var users []User
	err := u.db.WithContext(ctx).Where("dept_id = ?", deptId).Find(&users).Error
	return users, err
}

func (u *userDb) FindAll(ctx context.Context) ([]User, error) {
	var users []User
	err := u.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (u *userDb) Update(ctx context.Context, user User) error {
	user.Utime = float64(time.Now().Unix())
	return u.db.WithContext(ctx).Model(&User{}).Where("user_id = ?", user.UserId).Updates(&user).Error
}

func (u *userDb) Delete(ctx context.Context, userId int64) error {
	return u.db.WithContext(ctx).Where("user_id = ?", userId).Delete(&User{}).Error
}

// RoleDb 角色数据访问接口
type RoleDb interface {
	Create(ctx context.Context, role Role) error
	FindById(ctx context.Context, roleId int64) (Role, error)
	FindByCode(ctx context.Context, code string) (Role, error)
	FindAll(ctx context.Context) ([]Role, error)
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, roleId int64) error
}

type roleDb struct {
	db *gorm.DB
}

// NewRoleDb 创建RoleDb实例
func NewRoleDb(db *gorm.DB) RoleDb {
	return &roleDb{db: db}
}

func (r *roleDb) Create(ctx context.Context, role Role) error {
	role.Ctime = float64(time.Now().Unix())
	role.Utime = role.Ctime
	err := r.db.WithContext(ctx).Create(&role).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (r *roleDb) FindById(ctx context.Context, roleId int64) (Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("role_id = ?", roleId).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return Role{}, ErrDataRecordNotFound
	}
	return role, err
}

func (r *roleDb) FindByCode(ctx context.Context, code string) (Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err == gorm.ErrRecordNotFound {
		return Role{}, ErrDataRecordNotFound
	}
	return role, err
}

func (r *roleDb) FindAll(ctx context.Context) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *roleDb) Update(ctx context.Context, role Role) error {
	role.Utime = float64(time.Now().Unix())
	return r.db.WithContext(ctx).Model(&Role{}).Where("role_id = ?", role.RoleId).Updates(&role).Error
}

func (r *roleDb) Delete(ctx context.Context, roleId int64) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleId).Delete(&Role{}).Error
}

// PermissionDb 权限数据访问接口
type PermissionDb interface {
	Create(ctx context.Context, perm Permission) error
	FindById(ctx context.Context, permId int64) (Permission, error)
	FindByCode(ctx context.Context, code string) (Permission, error)
	FindAll(ctx context.Context) ([]Permission, error)
	Update(ctx context.Context, perm Permission) error
	Delete(ctx context.Context, permId int64) error
	FindByRoleId(ctx context.Context, roleId int64) ([]Permission, error)
}

type permissionDb struct {
	db *gorm.DB
}

// NewPermissionDb 创建PermissionDb实例
func NewPermissionDb(db *gorm.DB) PermissionDb {
	return &permissionDb{db: db}
}

func (p *permissionDb) Create(ctx context.Context, perm Permission) error {
	perm.Ctime = float64(time.Now().Unix())
	perm.Utime = perm.Ctime
	err := p.db.WithContext(ctx).Create(&perm).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (p *permissionDb) FindById(ctx context.Context, permId int64) (Permission, error) {
	var perm Permission
	err := p.db.WithContext(ctx).Where("perm_id = ?", permId).First(&perm).Error
	if err == gorm.ErrRecordNotFound {
		return Permission{}, ErrDataRecordNotFound
	}
	return perm, err
}

func (p *permissionDb) FindByCode(ctx context.Context, code string) (Permission, error) {
	var perm Permission
	err := p.db.WithContext(ctx).Where("code = ?", code).First(&perm).Error
	if err == gorm.ErrRecordNotFound {
		return Permission{}, ErrDataRecordNotFound
	}
	return perm, err
}

func (p *permissionDb) FindAll(ctx context.Context) ([]Permission, error) {
	var perms []Permission
	err := p.db.WithContext(ctx).Find(&perms).Error
	return perms, err
}

func (p *permissionDb) Update(ctx context.Context, perm Permission) error {
	perm.Utime = float64(time.Now().Unix())
	return p.db.WithContext(ctx).Model(&Permission{}).Where("perm_id = ?", perm.PermId).Updates(&perm).Error
}

func (p *permissionDb) Delete(ctx context.Context, permId int64) error {
	return p.db.WithContext(ctx).Where("perm_id = ?", permId).Delete(&Permission{}).Error
}

func (p *permissionDb) FindByRoleId(ctx context.Context, roleId int64) ([]Permission, error) {
	var perms []Permission
	err := p.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.perm_id = role_permissions.perm_id").
		Where("role_permissions.role_id = ?", roleId).
		Find(&perms).Error
	return perms, err
}

// UserRoleDb 用户-角色关联数据访问接口
type UserRoleDb interface {
	Create(ctx context.Context, ur UserRole) error
	Delete(ctx context.Context, userId, roleId int64) error
	FindRolesByUserId(ctx context.Context, userId int64) ([]Role, error)
	FindUsersByRoleId(ctx context.Context, roleId int64) ([]User, error)
}

type userRoleDb struct {
	db *gorm.DB
}

// NewUserRoleDb 创建UserRoleDb实例
func NewUserRoleDb(db *gorm.DB) UserRoleDb {
	return &userRoleDb{db: db}
}

func (ur *userRoleDb) Create(ctx context.Context, userRole UserRole) error {
	userRole.Ctime = float64(time.Now().Unix())
	err := ur.db.WithContext(ctx).Create(&userRole).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (ur *userRoleDb) Delete(ctx context.Context, userId, roleId int64) error {
	return ur.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userId, roleId).Delete(&UserRole{}).Error
}

func (ur *userRoleDb) FindRolesByUserId(ctx context.Context, userId int64) ([]Role, error) {
	var roles []Role
	err := ur.db.WithContext(ctx).
		Table("roles").
		Joins("INNER JOIN user_roles ON roles.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userId).
		Find(&roles).Error
	return roles, err
}

func (ur *userRoleDb) FindUsersByRoleId(ctx context.Context, roleId int64) ([]User, error) {
	var users []User
	err := ur.db.WithContext(ctx).
		Table("users").
		Joins("INNER JOIN user_roles ON users.user_id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleId).
		Find(&users).Error
	return users, err
}

// RolePermissionDb 角色-权限关联数据访问接口
type RolePermissionDb interface {
	Create(ctx context.Context, rp RolePermission) error
	Delete(ctx context.Context, roleId, permId int64) error
	FindPermissionsByRoleId(ctx context.Context, roleId int64) ([]Permission, error)
}

type rolePermissionDb struct {
	db *gorm.DB
}

// NewRolePermissionDb 创建RolePermissionDb实例
func NewRolePermissionDb(db *gorm.DB) RolePermissionDb {
	return &rolePermissionDb{db: db}
}

func (rp *rolePermissionDb) Create(ctx context.Context, rolePerm RolePermission) error {
	rolePerm.Ctime = float64(time.Now().Unix())
	err := rp.db.WithContext(ctx).Create(&rolePerm).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (rp *rolePermissionDb) Delete(ctx context.Context, roleId, permId int64) error {
	return rp.db.WithContext(ctx).Where("role_id = ? AND perm_id = ?", roleId, permId).Delete(&RolePermission{}).Error
}

func (rp *rolePermissionDb) FindPermissionsByRoleId(ctx context.Context, roleId int64) ([]Permission, error) {
	var perms []Permission
	err := rp.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.perm_id = role_permissions.perm_id").
		Where("role_permissions.role_id = ?", roleId).
		Find(&perms).Error
	return perms, err
}

// CronPermissionDb 任务权限控制数据访问接口
type CronPermissionDb interface {
	Create(ctx context.Context, cp CronPermission) error
	Delete(ctx context.Context, cronId, deptId int64) error
	FindDeptsByCronId(ctx context.Context, cronId int64) ([]Department, error)
	CheckPermission(ctx context.Context, cronId, deptId int64) (bool, error)
}

type cronPermissionDb struct {
	db *gorm.DB
}

// NewCronPermissionDb 创建CronPermissionDb实例
func NewCronPermissionDb(db *gorm.DB) CronPermissionDb {
	return &cronPermissionDb{db: db}
}

func (cp *cronPermissionDb) Create(ctx context.Context, cronPerm CronPermission) error {
	cronPerm.Ctime = float64(time.Now().Unix())
	err := cp.db.WithContext(ctx).Create(&cronPerm).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (cp *cronPermissionDb) Delete(ctx context.Context, cronId, deptId int64) error {
	return cp.db.WithContext(ctx).Where("cron_id = ? AND dept_id = ?", cronId, deptId).Delete(&CronPermission{}).Error
}

func (cp *cronPermissionDb) FindDeptsByCronId(ctx context.Context, cronId int64) ([]Department, error) {
	var depts []Department
	err := cp.db.WithContext(ctx).
		Table("departments").
		Joins("INNER JOIN cron_permissions ON departments.dept_id = cron_permissions.dept_id").
		Where("cron_permissions.cron_id = ?", cronId).
		Find(&depts).Error
	return depts, err
}

func (cp *cronPermissionDb) CheckPermission(ctx context.Context, cronId, deptId int64) (bool, error) {
	var count int64
	err := cp.db.WithContext(ctx).Model(&CronPermission{}).
		Where("cron_id = ? AND dept_id = ?", cronId, deptId).
		Count(&count).Error
	return count > 0, err
}

// 将 DAO 实体转换为 Domain 实体的辅助函数
func ToUserDomain(u User) domain.User {
	return domain.User{
		ID:       u.ID,
		UserId:   u.UserId,
		Username: u.Username,
		Password: u.Password,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		DeptId:   u.DeptId,
		Status:   u.Status,
		Ctime:    u.Ctime,
		Utime:    u.Utime,
	}
}

func ToDepartmentDomain(d Department) domain.Department {
	return domain.Department{
		ID:          d.ID,
		DeptId:      d.DeptId,
		Name:        d.Name,
		ParentId:    d.ParentId,
		Description: d.Description.String,
		Ctime:       d.Ctime,
		Utime:       d.Utime,
	}
}

func ToRoleDomain(r Role) domain.Role {
	return domain.Role{
		ID:          r.ID,
		RoleId:      r.RoleId,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description.String,
		Ctime:       r.Ctime,
		Utime:       r.Utime,
	}
}

func ToPermissionDomain(p Permission) domain.Permission {
	return domain.Permission{
		ID:          p.ID,
		PermId:      p.PermId,
		Name:        p.Name,
		Code:        p.Code,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description.String,
		Ctime:       p.Ctime,
		Utime:       p.Utime,
	}
}
