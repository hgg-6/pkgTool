package repository

import (
	"context"
	"database/sql"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
)

// DepartmentRepository 部门仓储接口
type DepartmentRepository interface {
	Create(ctx context.Context, dept domain.Department) error
	FindById(ctx context.Context, deptId int64) (domain.Department, error)
	FindAll(ctx context.Context) ([]domain.Department, error)
	FindByParentId(ctx context.Context, parentId int64) ([]domain.Department, error)
	Update(ctx context.Context, dept domain.Department) error
	Delete(ctx context.Context, deptId int64) error
}

type departmentRepository struct {
	db dao.DepartmentDb
}

// NewDepartmentRepository 创建DepartmentRepository实例
func NewDepartmentRepository(db dao.DepartmentDb) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (d *departmentRepository) Create(ctx context.Context, dept domain.Department) error {
	daoDept := dao.Department{
		DeptId:      dept.DeptId,
		Name:        dept.Name,
		ParentId:    dept.ParentId,
		Description: sql.NullString{String: dept.Description, Valid: dept.Description != ""},
	}
	return d.db.Create(ctx, daoDept)
}

func (d *departmentRepository) FindById(ctx context.Context, deptId int64) (domain.Department, error) {
	daoDept, err := d.db.FindById(ctx, deptId)
	if err != nil {
		return domain.Department{}, err
	}
	return dao.ToDepartmentDomain(daoDept), nil
}

func (d *departmentRepository) FindAll(ctx context.Context) ([]domain.Department, error) {
	daoDepts, err := d.db.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Department, 0, len(daoDepts))
	for _, dd := range daoDepts {
		result = append(result, dao.ToDepartmentDomain(dd))
	}
	return result, nil
}

func (d *departmentRepository) FindByParentId(ctx context.Context, parentId int64) ([]domain.Department, error) {
	daoDepts, err := d.db.FindByParentId(ctx, parentId)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Department, 0, len(daoDepts))
	for _, dd := range daoDepts {
		result = append(result, dao.ToDepartmentDomain(dd))
	}
	return result, nil
}

func (d *departmentRepository) Update(ctx context.Context, dept domain.Department) error {
	daoDept := dao.Department{
		DeptId:      dept.DeptId,
		Name:        dept.Name,
		ParentId:    dept.ParentId,
		Description: sql.NullString{String: dept.Description, Valid: dept.Description != ""},
	}
	return d.db.Update(ctx, daoDept)
}

func (d *departmentRepository) Delete(ctx context.Context, deptId int64) error {
	return d.db.Delete(ctx, deptId)
}

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, userId int64) (domain.User, error)
	FindByUsername(ctx context.Context, username string) (domain.User, error)
	FindByDeptId(ctx context.Context, deptId int64) ([]domain.User, error)
	FindAll(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, user domain.User) error
	Delete(ctx context.Context, userId int64) error
	FindUserRoles(ctx context.Context, userId int64) ([]domain.Role, error)
	FindUserPermissions(ctx context.Context, userId int64) ([]domain.Permission, error)
}

type userRepository struct {
	userDb     dao.UserDb
	userRoleDb dao.UserRoleDb
	permDb     dao.PermissionDb
}

// NewUserRepository 创建UserRepository实例
func NewUserRepository(userDb dao.UserDb, userRoleDb dao.UserRoleDb, permDb dao.PermissionDb) UserRepository {
	return &userRepository{
		userDb:     userDb,
		userRoleDb: userRoleDb,
		permDb:     permDb,
	}
}

func (u *userRepository) Create(ctx context.Context, user domain.User) error {
	daoUser := dao.User{
		UserId:   user.UserId,
		Username: user.Username,
		Password: user.Password,
		Email:    sql.NullString{String: user.Email, Valid: user.Email != ""},
		Phone:    sql.NullString{String: user.Phone, Valid: user.Phone != ""},
		DeptId:   user.DeptId,
		Status:   user.Status,
	}
	return u.userDb.Create(ctx, daoUser)
}

func (u *userRepository) FindById(ctx context.Context, userId int64) (domain.User, error) {
	daoUser, err := u.userDb.FindById(ctx, userId)
	if err != nil {
		return domain.User{}, err
	}
	return dao.ToUserDomain(daoUser), nil
}

func (u *userRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	daoUser, err := u.userDb.FindByUsername(ctx, username)
	if err != nil {
		return domain.User{}, err
	}
	return dao.ToUserDomain(daoUser), nil
}

func (u *userRepository) FindByDeptId(ctx context.Context, deptId int64) ([]domain.User, error) {
	daoUsers, err := u.userDb.FindByDeptId(ctx, deptId)
	if err != nil {
		return nil, err
	}
	result := make([]domain.User, 0, len(daoUsers))
	for _, du := range daoUsers {
		result = append(result, dao.ToUserDomain(du))
	}
	return result, nil
}

func (u *userRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	daoUsers, err := u.userDb.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.User, 0, len(daoUsers))
	for _, du := range daoUsers {
		result = append(result, dao.ToUserDomain(du))
	}
	return result, nil
}

func (u *userRepository) Update(ctx context.Context, user domain.User) error {
	daoUser := dao.User{
		UserId:   user.UserId,
		Username: user.Username,
		Password: user.Password,
		Email:    sql.NullString{String: user.Email, Valid: user.Email != ""},
		Phone:    sql.NullString{String: user.Phone, Valid: user.Phone != ""},
		DeptId:   user.DeptId,
		Status:   user.Status,
	}
	return u.userDb.Update(ctx, daoUser)
}

func (u *userRepository) Delete(ctx context.Context, userId int64) error {
	return u.userDb.Delete(ctx, userId)
}

func (u *userRepository) FindUserRoles(ctx context.Context, userId int64) ([]domain.Role, error) {
	daoRoles, err := u.userRoleDb.FindRolesByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Role, 0, len(daoRoles))
	for _, dr := range daoRoles {
		result = append(result, dao.ToRoleDomain(dr))
	}
	return result, nil
}

func (u *userRepository) FindUserPermissions(ctx context.Context, userId int64) ([]domain.Permission, error) {
	// 查询用户的所有角色
	roles, err := u.userRoleDb.FindRolesByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	// 收集所有角色的权限（去重）
	permMap := make(map[int64]domain.Permission)
	for _, role := range roles {
		perms, err := u.permDb.FindByRoleId(ctx, role.RoleId)
		if err != nil {
			continue
		}
		for _, p := range perms {
			permMap[p.PermId] = dao.ToPermissionDomain(p)
		}
	}

	// 转为数组
	result := make([]domain.Permission, 0, len(permMap))
	for _, perm := range permMap {
		result = append(result, perm)
	}
	return result, nil
}

// RoleRepository 角色仓储接口
type RoleRepository interface {
	Create(ctx context.Context, role domain.Role) error
	FindById(ctx context.Context, roleId int64) (domain.Role, error)
	FindByCode(ctx context.Context, code string) (domain.Role, error)
	FindAll(ctx context.Context) ([]domain.Role, error)
	Update(ctx context.Context, role domain.Role) error
	Delete(ctx context.Context, roleId int64) error
	AssignPermission(ctx context.Context, roleId, permId int64) error
	RemovePermission(ctx context.Context, roleId, permId int64) error
	FindRolePermissions(ctx context.Context, roleId int64) ([]domain.Permission, error)
}

type roleRepository struct {
	roleDb     dao.RoleDb
	rolePermDb dao.RolePermissionDb
}

// NewRoleRepository 创建RoleRepository实例
func NewRoleRepository(roleDb dao.RoleDb, rolePermDb dao.RolePermissionDb) RoleRepository {
	return &roleRepository{
		roleDb:     roleDb,
		rolePermDb: rolePermDb,
	}
}

func (r *roleRepository) Create(ctx context.Context, role domain.Role) error {
	daoRole := dao.Role{
		RoleId:      role.RoleId,
		Name:        role.Name,
		Code:        role.Code,
		Description: sql.NullString{String: role.Description, Valid: role.Description != ""},
	}
	return r.roleDb.Create(ctx, daoRole)
}

func (r *roleRepository) FindById(ctx context.Context, roleId int64) (domain.Role, error) {
	daoRole, err := r.roleDb.FindById(ctx, roleId)
	if err != nil {
		return domain.Role{}, err
	}
	return dao.ToRoleDomain(daoRole), nil
}

func (r *roleRepository) FindByCode(ctx context.Context, code string) (domain.Role, error) {
	daoRole, err := r.roleDb.FindByCode(ctx, code)
	if err != nil {
		return domain.Role{}, err
	}
	return dao.ToRoleDomain(daoRole), nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	daoRoles, err := r.roleDb.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Role, 0, len(daoRoles))
	for _, dr := range daoRoles {
		result = append(result, dao.ToRoleDomain(dr))
	}
	return result, nil
}

func (r *roleRepository) Update(ctx context.Context, role domain.Role) error {
	daoRole := dao.Role{
		RoleId:      role.RoleId,
		Name:        role.Name,
		Code:        role.Code,
		Description: sql.NullString{String: role.Description, Valid: role.Description != ""},
	}
	return r.roleDb.Update(ctx, daoRole)
}

func (r *roleRepository) Delete(ctx context.Context, roleId int64) error {
	return r.roleDb.Delete(ctx, roleId)
}

func (r *roleRepository) AssignPermission(ctx context.Context, roleId, permId int64) error {
	rolePerm := dao.RolePermission{
		RoleId: roleId,
		PermId: permId,
	}
	return r.rolePermDb.Create(ctx, rolePerm)
}

func (r *roleRepository) RemovePermission(ctx context.Context, roleId, permId int64) error {
	return r.rolePermDb.Delete(ctx, roleId, permId)
}

func (r *roleRepository) FindRolePermissions(ctx context.Context, roleId int64) ([]domain.Permission, error) {
	daoPerms, err := r.rolePermDb.FindPermissionsByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Permission, 0, len(daoPerms))
	for _, dp := range daoPerms {
		result = append(result, dao.ToPermissionDomain(dp))
	}
	return result, nil
}

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	Create(ctx context.Context, perm domain.Permission) error
	FindById(ctx context.Context, permId int64) (domain.Permission, error)
	FindByCode(ctx context.Context, code string) (domain.Permission, error)
	FindAll(ctx context.Context) ([]domain.Permission, error)
	Update(ctx context.Context, perm domain.Permission) error
	Delete(ctx context.Context, permId int64) error
}

type permissionRepository struct {
	permDb dao.PermissionDb
}

// NewPermissionRepository 创建PermissionRepository实例
func NewPermissionRepository(permDb dao.PermissionDb) PermissionRepository {
	return &permissionRepository{permDb: permDb}
}

func (p *permissionRepository) Create(ctx context.Context, perm domain.Permission) error {
	daoPerm := dao.Permission{
		PermId:      perm.PermId,
		Name:        perm.Name,
		Code:        perm.Code,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: sql.NullString{String: perm.Description, Valid: perm.Description != ""},
	}
	return p.permDb.Create(ctx, daoPerm)
}

func (p *permissionRepository) FindById(ctx context.Context, permId int64) (domain.Permission, error) {
	daoPerm, err := p.permDb.FindById(ctx, permId)
	if err != nil {
		return domain.Permission{}, err
	}
	return dao.ToPermissionDomain(daoPerm), nil
}

func (p *permissionRepository) FindByCode(ctx context.Context, code string) (domain.Permission, error) {
	daoPerm, err := p.permDb.FindByCode(ctx, code)
	if err != nil {
		return domain.Permission{}, err
	}
	return dao.ToPermissionDomain(daoPerm), nil
}

func (p *permissionRepository) FindAll(ctx context.Context) ([]domain.Permission, error) {
	daoPerms, err := p.permDb.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Permission, 0, len(daoPerms))
	for _, dp := range daoPerms {
		result = append(result, dao.ToPermissionDomain(dp))
	}
	return result, nil
}

func (p *permissionRepository) Update(ctx context.Context, perm domain.Permission) error {
	daoPerm := dao.Permission{
		PermId:      perm.PermId,
		Name:        perm.Name,
		Code:        perm.Code,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: sql.NullString{String: perm.Description, Valid: perm.Description != ""},
	}
	return p.permDb.Update(ctx, daoPerm)
}

func (p *permissionRepository) Delete(ctx context.Context, permId int64) error {
	return p.permDb.Delete(ctx, permId)
}

// AuthRepository 权限认证仓储接口
type AuthRepository interface {
	AssignRoleToUser(ctx context.Context, userId, roleId int64) error
	RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error
	GrantCronPermission(ctx context.Context, cronId, deptId int64) error
	RevokeCronPermission(ctx context.Context, cronId, deptId int64) error
	CheckCronPermission(ctx context.Context, cronId, deptId int64) (bool, error)
}

type authRepository struct {
	userRoleDb dao.UserRoleDb
	cronPermDb dao.CronPermissionDb
}

// NewAuthRepository 创建AuthRepository实例
func NewAuthRepository(userRoleDb dao.UserRoleDb, cronPermDb dao.CronPermissionDb) AuthRepository {
	return &authRepository{
		userRoleDb: userRoleDb,
		cronPermDb: cronPermDb,
	}
}

func (a *authRepository) AssignRoleToUser(ctx context.Context, userId, roleId int64) error {
	ur := dao.UserRole{
		UserId: userId,
		RoleId: roleId,
	}
	return a.userRoleDb.Create(ctx, ur)
}

func (a *authRepository) RemoveRoleFromUser(ctx context.Context, userId, roleId int64) error {
	return a.userRoleDb.Delete(ctx, userId, roleId)
}

func (a *authRepository) GrantCronPermission(ctx context.Context, cronId, deptId int64) error {
	cp := dao.CronPermission{
		CronId: cronId,
		DeptId: deptId,
	}
	return a.cronPermDb.Create(ctx, cp)
}

func (a *authRepository) RevokeCronPermission(ctx context.Context, cronId, deptId int64) error {
	return a.cronPermDb.Delete(ctx, cronId, deptId)
}

func (a *authRepository) CheckCronPermission(ctx context.Context, cronId, deptId int64) (bool, error) {
	return a.cronPermDb.CheckPermission(ctx, cronId, deptId)
}
