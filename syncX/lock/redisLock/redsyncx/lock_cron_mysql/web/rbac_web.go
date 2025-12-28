package web

import (
	"strconv"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"

	"github.com/gin-gonic/gin"
)

// DepartmentWeb 部门Web接口
type DepartmentWeb struct {
	deptSvc service.DepartmentService
	l       logx.Loggerx
}

// NewDepartmentWeb 创建DepartmentWeb实例
func NewDepartmentWeb(deptSvc service.DepartmentService, l logx.Loggerx) *DepartmentWeb {
	return &DepartmentWeb{deptSvc: deptSvc, l: l}
}

func (d *DepartmentWeb) Register(server *gin.Engine) {
	g := server.Group("/department")
	{
		g.POST("/create", d.CreateDepartment)
		g.GET("/get/:dept_id", d.GetDepartment)
		g.GET("/list", d.GetAllDepartments)
		g.GET("/sub/:parent_id", d.GetSubDepartments)
		g.PUT("/update", d.UpdateDepartment)
		g.DELETE("/delete/:dept_id", d.DeleteDepartment)
	}
}

func (d *DepartmentWeb) CreateDepartment(ctx *gin.Context) {
	var dept domain.Department
	if err := ctx.ShouldBindJSON(&dept); err != nil {
		d.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := d.deptSvc.CreateDepartment(ctx.Request.Context(), dept); err != nil {
		d.l.Error("创建部门失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "创建部门失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "创建成功"})
}

func (d *DepartmentWeb) GetDepartment(ctx *gin.Context) {
	deptId, err := strconv.ParseInt(ctx.Param("dept_id"), 10, 64)
	if err != nil {
		d.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	dept, err := d.deptSvc.GetDepartment(ctx.Request.Context(), deptId)
	if err != nil {
		d.l.Error("查询部门失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询部门失败"})
		return
	}

	ctx.JSON(200, dept)
}

func (d *DepartmentWeb) GetAllDepartments(ctx *gin.Context) {
	depts, err := d.deptSvc.GetAllDepartments(ctx.Request.Context())
	if err != nil {
		d.l.Error("查询部门列表失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询部门列表失败"})
		return
	}

	ctx.JSON(200, depts)
}

func (d *DepartmentWeb) GetSubDepartments(ctx *gin.Context) {
	parentId, err := strconv.ParseInt(ctx.Param("parent_id"), 10, 64)
	if err != nil {
		d.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	depts, err := d.deptSvc.GetSubDepartments(ctx.Request.Context(), parentId)
	if err != nil {
		d.l.Error("查询子部门失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询子部门失败"})
		return
	}

	ctx.JSON(200, depts)
}

func (d *DepartmentWeb) UpdateDepartment(ctx *gin.Context) {
	var dept domain.Department
	if err := ctx.ShouldBindJSON(&dept); err != nil {
		d.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := d.deptSvc.UpdateDepartment(ctx.Request.Context(), dept); err != nil {
		d.l.Error("更新部门失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "更新部门失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "更新成功"})
}

func (d *DepartmentWeb) DeleteDepartment(ctx *gin.Context) {
	deptId, err := strconv.ParseInt(ctx.Param("dept_id"), 10, 64)
	if err != nil {
		d.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := d.deptSvc.DeleteDepartment(ctx.Request.Context(), deptId); err != nil {
		d.l.Error("删除部门失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "删除部门失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "删除成功"})
}

// UserWeb 用户Web接口
type UserWeb struct {
	userSvc service.UserService
	l       logx.Loggerx
}

// NewUserWeb 创建UserWeb实例
func NewUserWeb(userSvc service.UserService, l logx.Loggerx) *UserWeb {
	return &UserWeb{userSvc: userSvc, l: l}
}

func (u *UserWeb) Register(server *gin.Engine) {
	g := server.Group("/user")
	{
		g.POST("/create", u.CreateUser)
		g.POST("/login", u.Login)
		g.GET("/get/:user_id", u.GetUser)
		g.GET("/list", u.GetAllUsers)
		g.GET("/dept/:dept_id", u.GetUsersByDepartment)
		g.PUT("/update", u.UpdateUser)
		g.DELETE("/delete/:user_id", u.DeleteUser)
		g.GET("/roles/:user_id", u.GetUserRoles)
		g.GET("/permissions/:user_id", u.GetUserPermissions)
		g.POST("/change-password", u.ChangePassword)
	}
}

func (u *UserWeb) CreateUser(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := u.userSvc.CreateUser(ctx.Request.Context(), user); err != nil {
		u.l.Error("创建用户失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "创建用户失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "创建成功"})
}

func (u *UserWeb) Login(ctx *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	user, err := u.userSvc.Login(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		u.l.Error("登录失败", logx.Error(err))
		ctx.JSON(401, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 不返回密码
	user.Password = ""
	ctx.JSON(200, user)
}

func (u *UserWeb) GetUser(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	user, err := u.userSvc.GetUser(ctx.Request.Context(), userId)
	if err != nil {
		u.l.Error("查询用户失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询用户失败"})
		return
	}

	// 不返回密码
	user.Password = ""
	ctx.JSON(200, user)
}

func (u *UserWeb) GetAllUsers(ctx *gin.Context) {
	users, err := u.userSvc.GetAllUsers(ctx.Request.Context())
	if err != nil {
		u.l.Error("查询用户列表失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询用户列表失败"})
		return
	}

	// 不返回密码
	for i := range users {
		users[i].Password = ""
	}
	ctx.JSON(200, users)
}

func (u *UserWeb) GetUsersByDepartment(ctx *gin.Context) {
	deptId, err := strconv.ParseInt(ctx.Param("dept_id"), 10, 64)
	if err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	users, err := u.userSvc.GetUsersByDepartment(ctx.Request.Context(), deptId)
	if err != nil {
		u.l.Error("查询部门用户失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询部门用户失败"})
		return
	}

	// 不返回密码
	for i := range users {
		users[i].Password = ""
	}
	ctx.JSON(200, users)
}

func (u *UserWeb) UpdateUser(ctx *gin.Context) {
	var user domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := u.userSvc.UpdateUser(ctx.Request.Context(), user); err != nil {
		u.l.Error("更新用户失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "更新用户失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "更新成功"})
}

func (u *UserWeb) DeleteUser(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := u.userSvc.DeleteUser(ctx.Request.Context(), userId); err != nil {
		u.l.Error("删除用户失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "删除用户失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "删除成功"})
}

func (u *UserWeb) GetUserRoles(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	roles, err := u.userSvc.GetUserRoles(ctx.Request.Context(), userId)
	if err != nil {
		u.l.Error("查询用户角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询用户角色失败"})
		return
	}

	ctx.JSON(200, roles)
}

func (u *UserWeb) GetUserPermissions(ctx *gin.Context) {
	userId, err := strconv.ParseInt(ctx.Param("user_id"), 10, 64)
	if err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	perms, err := u.userSvc.GetUserPermissions(ctx.Request.Context(), userId)
	if err != nil {
		u.l.Error("查询用户权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询用户权限失败"})
		return
	}

	ctx.JSON(200, perms)
}

func (u *UserWeb) ChangePassword(ctx *gin.Context) {
	var req struct {
		UserId      int64  `json:"user_id"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		u.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := u.userSvc.ChangePassword(ctx.Request.Context(), req.UserId, req.OldPassword, req.NewPassword); err != nil {
		u.l.Error("修改密码失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "修改密码失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "修改成功"})
}

// RoleWeb 角色Web接口
type RoleWeb struct {
	roleSvc service.RoleService
	l       logx.Loggerx
}

// NewRoleWeb 创建RoleWeb实例
func NewRoleWeb(roleSvc service.RoleService, l logx.Loggerx) *RoleWeb {
	return &RoleWeb{roleSvc: roleSvc, l: l}
}

func (r *RoleWeb) Register(server *gin.Engine) {
	g := server.Group("/role")
	{
		g.POST("/create", r.CreateRole)
		g.GET("/get/:role_id", r.GetRole)
		g.GET("/list", r.GetAllRoles)
		g.PUT("/update", r.UpdateRole)
		g.DELETE("/delete/:role_id", r.DeleteRole)
		g.POST("/assign-permission", r.AssignPermission)
		g.POST("/remove-permission", r.RemovePermission)
		g.GET("/permissions/:role_id", r.GetRolePermissions)
	}
}

func (r *RoleWeb) CreateRole(ctx *gin.Context) {
	var role domain.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := r.roleSvc.CreateRole(ctx.Request.Context(), role); err != nil {
		r.l.Error("创建角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "创建角色失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "创建成功"})
}

func (r *RoleWeb) GetRole(ctx *gin.Context) {
	roleId, err := strconv.ParseInt(ctx.Param("role_id"), 10, 64)
	if err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	role, err := r.roleSvc.GetRole(ctx.Request.Context(), roleId)
	if err != nil {
		r.l.Error("查询角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询角色失败"})
		return
	}

	ctx.JSON(200, role)
}

func (r *RoleWeb) GetAllRoles(ctx *gin.Context) {
	roles, err := r.roleSvc.GetAllRoles(ctx.Request.Context())
	if err != nil {
		r.l.Error("查询角色列表失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询角色列表失败"})
		return
	}

	ctx.JSON(200, roles)
}

func (r *RoleWeb) UpdateRole(ctx *gin.Context) {
	var role domain.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := r.roleSvc.UpdateRole(ctx.Request.Context(), role); err != nil {
		r.l.Error("更新角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "更新角色失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "更新成功"})
}

func (r *RoleWeb) DeleteRole(ctx *gin.Context) {
	roleId, err := strconv.ParseInt(ctx.Param("role_id"), 10, 64)
	if err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := r.roleSvc.DeleteRole(ctx.Request.Context(), roleId); err != nil {
		r.l.Error("删除角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "删除角色失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "删除成功"})
}

func (r *RoleWeb) AssignPermission(ctx *gin.Context) {
	var req struct {
		RoleId int64 `json:"role_id"`
		PermId int64 `json:"perm_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := r.roleSvc.AssignPermission(ctx.Request.Context(), req.RoleId, req.PermId); err != nil {
		r.l.Error("分配权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "分配权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "分配成功"})
}

func (r *RoleWeb) RemovePermission(ctx *gin.Context) {
	var req struct {
		RoleId int64 `json:"role_id"`
		PermId int64 `json:"perm_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := r.roleSvc.RemovePermission(ctx.Request.Context(), req.RoleId, req.PermId); err != nil {
		r.l.Error("移除权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "移除权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "移除成功"})
}

func (r *RoleWeb) GetRolePermissions(ctx *gin.Context) {
	roleId, err := strconv.ParseInt(ctx.Param("role_id"), 10, 64)
	if err != nil {
		r.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	perms, err := r.roleSvc.GetRolePermissions(ctx.Request.Context(), roleId)
	if err != nil {
		r.l.Error("查询角色权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询角色权限失败"})
		return
	}

	ctx.JSON(200, perms)
}

// PermissionWeb 权限Web接口
type PermissionWeb struct {
	permSvc service.PermissionService
	l       logx.Loggerx
}

// NewPermissionWeb 创建PermissionWeb实例
func NewPermissionWeb(permSvc service.PermissionService, l logx.Loggerx) *PermissionWeb {
	return &PermissionWeb{permSvc: permSvc, l: l}
}

func (p *PermissionWeb) Register(server *gin.Engine) {
	g := server.Group("/permission")
	{
		g.POST("/create", p.CreatePermission)
		g.GET("/get/:perm_id", p.GetPermission)
		g.GET("/list", p.GetAllPermissions)
		g.PUT("/update", p.UpdatePermission)
		g.DELETE("/delete/:perm_id", p.DeletePermission)
	}
}

func (p *PermissionWeb) CreatePermission(ctx *gin.Context) {
	var perm domain.Permission
	if err := ctx.ShouldBindJSON(&perm); err != nil {
		p.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := p.permSvc.CreatePermission(ctx.Request.Context(), perm); err != nil {
		p.l.Error("创建权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "创建权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "创建成功"})
}

func (p *PermissionWeb) GetPermission(ctx *gin.Context) {
	permId, err := strconv.ParseInt(ctx.Param("perm_id"), 10, 64)
	if err != nil {
		p.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	perm, err := p.permSvc.GetPermission(ctx.Request.Context(), permId)
	if err != nil {
		p.l.Error("查询权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询权限失败"})
		return
	}

	ctx.JSON(200, perm)
}

func (p *PermissionWeb) GetAllPermissions(ctx *gin.Context) {
	perms, err := p.permSvc.GetAllPermissions(ctx.Request.Context())
	if err != nil {
		p.l.Error("查询权限列表失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "查询权限列表失败"})
		return
	}

	ctx.JSON(200, perms)
}

func (p *PermissionWeb) UpdatePermission(ctx *gin.Context) {
	var perm domain.Permission
	if err := ctx.ShouldBindJSON(&perm); err != nil {
		p.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := p.permSvc.UpdatePermission(ctx.Request.Context(), perm); err != nil {
		p.l.Error("更新权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "更新权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "更新成功"})
}

func (p *PermissionWeb) DeletePermission(ctx *gin.Context) {
	permId, err := strconv.ParseInt(ctx.Param("perm_id"), 10, 64)
	if err != nil {
		p.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := p.permSvc.DeletePermission(ctx.Request.Context(), permId); err != nil {
		p.l.Error("删除权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "删除权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "删除成功"})
}

// AuthWeb 认证授权Web接口
type AuthWeb struct {
	authSvc service.AuthService
	l       logx.Loggerx
}

// NewAuthWeb 创建AuthWeb实例
func NewAuthWeb(authSvc service.AuthService, l logx.Loggerx) *AuthWeb {
	return &AuthWeb{authSvc: authSvc, l: l}
}

func (a *AuthWeb) Register(server *gin.Engine) {
	g := server.Group("/auth")
	{
		g.POST("/assign-role", a.AssignRoleToUser)
		g.POST("/remove-role", a.RemoveRoleFromUser)
		g.POST("/grant-cron-permission", a.GrantCronPermission)
		g.POST("/revoke-cron-permission", a.RevokeCronPermission)
		g.POST("/check-permission", a.CheckUserPermission)
		g.POST("/check-cron-permission", a.CheckCronPermission)
	}
}

func (a *AuthWeb) AssignRoleToUser(ctx *gin.Context) {
	var req struct {
		UserId int64 `json:"user_id"`
		RoleId int64 `json:"role_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := a.authSvc.AssignRoleToUser(ctx.Request.Context(), req.UserId, req.RoleId); err != nil {
		a.l.Error("分配角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "分配角色失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "分配成功"})
}

func (a *AuthWeb) RemoveRoleFromUser(ctx *gin.Context) {
	var req struct {
		UserId int64 `json:"user_id"`
		RoleId int64 `json:"role_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := a.authSvc.RemoveRoleFromUser(ctx.Request.Context(), req.UserId, req.RoleId); err != nil {
		a.l.Error("移除角色失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "移除角色失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "移除成功"})
}

func (a *AuthWeb) GrantCronPermission(ctx *gin.Context) {
	var req struct {
		CronId int64 `json:"cron_id"`
		DeptId int64 `json:"dept_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := a.authSvc.GrantCronPermission(ctx.Request.Context(), req.CronId, req.DeptId); err != nil {
		a.l.Error("授予任务权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "授予任务权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "授予成功"})
}

func (a *AuthWeb) RevokeCronPermission(ctx *gin.Context) {
	var req struct {
		CronId int64 `json:"cron_id"`
		DeptId int64 `json:"dept_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := a.authSvc.RevokeCronPermission(ctx.Request.Context(), req.CronId, req.DeptId); err != nil {
		a.l.Error("撤销任务权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "撤销任务权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"message": "撤销成功"})
}

func (a *AuthWeb) CheckUserPermission(ctx *gin.Context) {
	var req struct {
		UserId         int64  `json:"user_id"`
		PermissionCode string `json:"permission_code"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	hasPermission, err := a.authSvc.CheckUserPermission(ctx.Request.Context(), req.UserId, req.PermissionCode)
	if err != nil {
		a.l.Error("检查权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "检查权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"has_permission": hasPermission})
}

func (a *AuthWeb) CheckCronPermission(ctx *gin.Context) {
	var req struct {
		UserId int64 `json:"user_id"`
		CronId int64 `json:"cron_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("参数错误", logx.Error(err))
		ctx.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	hasPermission, err := a.authSvc.CheckCronPermission(ctx.Request.Context(), req.UserId, req.CronId)
	if err != nil {
		a.l.Error("检查任务权限失败", logx.Error(err))
		ctx.JSON(500, gin.H{"error": "检查任务权限失败"})
		return
	}

	ctx.JSON(200, gin.H{"has_permission": hasPermission})
}
