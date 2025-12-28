package service

import (
	"context"
	"errors"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials     = errors.New("invalid username or password")
	ErrInsufficientPermission = errors.New("insufficient permission")
)

// DepartmentService 部门服务接口
type DepartmentService interface {
	CreateDepartment(ctx context.Context, dept domain.Department) error
	GetDepartment(ctx context.Context, deptId int64) (domain.Department, error)
	GetAllDepartments(ctx context.Context) ([]domain.Department, error)
	GetSubDepartments(ctx context.Context, parentId int64) ([]domain.Department, error)
	UpdateDepartment(ctx context.Context, dept domain.Department) error
	DeleteDepartment(ctx context.Context, deptId int64) error
}

type departmentService struct {
	deptRepo repository.DepartmentRepository
}

// NewDepartmentService 创建DepartmentService实例
func NewDepartmentService(deptRepo repository.DepartmentRepository) DepartmentService {
	return &departmentService{deptRepo: deptRepo}
}

func (d *departmentService) CreateDepartment(ctx context.Context, dept domain.Department) error {
	return d.deptRepo.Create(ctx, dept)
}

func (d *departmentService) GetDepartment(ctx context.Context, deptId int64) (domain.Department, error) {
	return d.deptRepo.FindById(ctx, deptId)
}

func (d *departmentService) GetAllDepartments(ctx context.Context) ([]domain.Department, error) {
	return d.deptRepo.FindAll(ctx)
}

func (d *departmentService) GetSubDepartments(ctx context.Context, parentId int64) ([]domain.Department, error) {
	return d.deptRepo.FindByParentId(ctx, parentId)
}

func (d *departmentService) UpdateDepartment(ctx context.Context, dept domain.Department) error {
	return d.deptRepo.Update(ctx, dept)
}

func (d *departmentService) DeleteDepartment(ctx context.Context, deptId int64) error {
	return d.deptRepo.Delete(ctx, deptId)
}

// UserService 用户服务接口
type UserService interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUser(ctx context.Context, userId int64) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	GetUsersByDepartment(ctx context.Context, deptId int64) ([]domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	DeleteUser(ctx context.Context, userId int64) error
	Login(ctx context.Context, username, password string) (domain.User, error)
	GetUserRoles(ctx context.Context, userId int64) ([]domain.Role, error)
	GetUserPermissions(ctx context.Context, userId int64) ([]domain.Permission, error)
	ChangePassword(ctx context.Context, userId int64, oldPassword, newPassword string) error
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建UserService实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (u *userService) CreateUser(ctx context.Context, user domain.User) error {
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return u.userRepo.Create(ctx, user)
}

func (u *userService) GetUser(ctx context.Context, userId int64) (domain.User, error) {
	return u.userRepo.FindById(ctx, userId)
}

func (u *userService) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	return u.userRepo.FindByUsername(ctx, username)
}

func (u *userService) GetUsersByDepartment(ctx context.Context, deptId int64) ([]domain.User, error) {
	return u.userRepo.FindByDeptId(ctx, deptId)
}

func (u *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.FindAll(ctx)
}

func (u *userService) UpdateUser(ctx context.Context, user domain.User) error {
	return u.userRepo.Update(ctx, user)
}

func (u *userService) DeleteUser(ctx context.Context, userId int64) error {
	return u.userRepo.Delete(ctx, userId)
}

func (u *userService) Login(ctx context.Context, username, password string) (domain.User, error) {
	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (u *userService) GetUserRoles(ctx context.Context, userId int64) ([]domain.Role, error) {
	return u.userRepo.FindUserRoles(ctx, userId)
}

func (u *userService) GetUserPermissions(ctx context.Context, userId int64) ([]domain.Permission, error) {
	return u.userRepo.FindUserPermissions(ctx, userId)
}

func (u *userService) ChangePassword(ctx context.Context, userId int64, oldPassword, newPassword string) error {
	user, err := u.userRepo.FindById(ctx, userId)
	if err != nil {
		return err
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return u.userRepo.Update(ctx, user)
}

// RoleService 角色服务接口
type RoleService interface {
	CreateRole(ctx context.Context, role domain.Role) error
	GetRole(ctx context.Context, roleId int64) (domain.Role, error)
	GetRoleByCode(ctx context.Context, code string) (domain.Role, error)
	GetAllRoles(ctx context.Context) ([]domain.Role, error)
	UpdateRole(ctx context.Context, role domain.Role) error
	DeleteRole(ctx context.Context, roleId int64) error
	AssignPermission(ctx context.Context, roleId, permId int64) error
	RemovePermission(ctx context.Context, roleId, permId int64) error
	GetRolePermissions(ctx context.Context, roleId int64) ([]domain.Permission, error)
}

type roleService struct {
	roleRepo repository.RoleRepository
}

// NewRoleService 创建RoleService实例
func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (r *roleService) CreateRole(ctx context.Context, role domain.Role) error {
	return r.roleRepo.Create(ctx, role)
}

func (r *roleService) GetRole(ctx context.Context, roleId int64) (domain.Role, error) {
	return r.roleRepo.FindById(ctx, roleId)
}

func (r *roleService) GetRoleByCode(ctx context.Context, code string) (domain.Role, error) {
	return r.roleRepo.FindByCode(ctx, code)
}

func (r *roleService) GetAllRoles(ctx context.Context) ([]domain.Role, error) {
	return r.roleRepo.FindAll(ctx)
}

func (r *roleService) UpdateRole(ctx context.Context, role domain.Role) error {
	return r.roleRepo.Update(ctx, role)
}

func (r *roleService) DeleteRole(ctx context.Context, roleId int64) error {
	return r.roleRepo.Delete(ctx, roleId)
}

func (r *roleService) AssignPermission(ctx context.Context, roleId, permId int64) error {
	return r.roleRepo.AssignPermission(ctx, roleId, permId)
}

func (r *roleService) RemovePermission(ctx context.Context, roleId, permId int64) error {
	return r.roleRepo.RemovePermission(ctx, roleId, permId)
}

func (r *roleService) GetRolePermissions(ctx context.Context, roleId int64) ([]domain.Permission, error) {
	return r.roleRepo.FindRolePermissions(ctx, roleId)
}

// PermissionService 权限服务接口
type PermissionService interface {
	CreatePermission(ctx context.Context, perm domain.Permission) error
	GetPermission(ctx context.Context, permId int64) (domain.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (domain.Permission, error)
	GetAllPermissions(ctx context.Context) ([]domain.Permission, error)
	UpdatePermission(ctx context.Context, perm domain.Permission) error
	DeletePermission(ctx context.Context, permId int64) error
}

type permissionService struct {
	permRepo repository.PermissionRepository
}

// NewPermissionService 创建PermissionService实例
func NewPermissionService(permRepo repository.PermissionRepository) PermissionService {
	return &permissionService{permRepo: permRepo}
}

func (p *permissionService) CreatePermission(ctx context.Context, perm domain.Permission) error {
	return p.permRepo.Create(ctx, perm)
}

func (p *permissionService) GetPermission(ctx context.Context, permId int64) (domain.Permission, error) {
	return p.permRepo.FindById(ctx, permId)
}

func (p *permissionService) GetPermissionByCode(ctx context.Context, code string) (domain.Permission, error) {
	return p.permRepo.FindByCode(ctx, code)
}

func (p *permissionService) GetAllPermissions(ctx context.Context) ([]domain.Permission, error) {
	return p.permRepo.FindAll(ctx)
}

func (p *permissionService) UpdatePermission(ctx context.Context, perm domain.Permission) error {
	return p.permRepo.Update(ctx, perm)
}

func (p *permissionService) DeletePermission(ctx context.Context, permId int64) error {
	return p.permRepo.Delete(ctx, permId)
}

// AuthService 权限认证服务接口
type AuthService interface {
	AssignRoleToUser(ctx context.Context, userId, roleId int64) error
	RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error
	GrantCronPermission(ctx context.Context, cronId, deptId int64) error
	RevokeCronPermission(ctx context.Context, cronId, deptId int64) error
	CheckUserPermission(ctx context.Context, userId int64, permissionCode string) (bool, error)
	CheckCronPermission(ctx context.Context, userId, cronId int64) (bool, error)
}

type authService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
}

// NewAuthService 创建AuthService实例
func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserRepository) AuthService {
	return &authService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (a *authService) AssignRoleToUser(ctx context.Context, userId, roleId int64) error {
	return a.authRepo.AssignRoleToUser(ctx, userId, roleId)
}

func (a *authService) RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error {
	return a.authRepo.RemoveRoleFromUser(ctx, userId, roleId)
}

func (a *authService) GrantCronPermission(ctx context.Context, cronId, deptId int64) error {
	return a.authRepo.GrantCronPermission(ctx, cronId, deptId)
}

func (a *authService) RevokeCronPermission(ctx context.Context, cronId, deptId int64) error {
	return a.authRepo.RevokeCronPermission(ctx, cronId, deptId)
}

func (a *authService) CheckUserPermission(ctx context.Context, userId int64, permissionCode string) (bool, error) {
	// 获取用户所有权限
	permissions, err := a.userRepo.FindUserPermissions(ctx, userId)
	if err != nil {
		return false, err
	}

	// 检查是否包含目标权限
	for _, perm := range permissions {
		if perm.Code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

func (a *authService) CheckCronPermission(ctx context.Context, userId, cronId int64) (bool, error) {
	// 获取用户所属部门
	user, err := a.userRepo.FindById(ctx, userId)
	if err != nil {
		return false, err
	}

	// 检查部门是否有权限管理该任务
	return a.authRepo.CheckCronPermission(ctx, cronId, user.DeptId)
}
