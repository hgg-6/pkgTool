package lock_cron_mysql

import (
	"context"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/config"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/executor"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/middleware"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/scheduler"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/web"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	jwtX2 "gitee.com/hgg_test/pkg_tool/v2/webx/ginx/middleware/jwtX2"
)

// CronMysql 定时任务系统主类
type CronMysql struct {
	web     *gin.Engine
	db      *gorm.DB
	redSync redsyncx.RedSyncIn
	l       logx.Loggerx
	cfg     *config.Config
	rdb     redis.Cmdable

	// 各层组件
	cronWeb        *web.CronWeb
	departmentWeb  *web.DepartmentWeb
	userWeb        *web.UserWeb
	roleWeb        *web.RoleWeb
	permissionWeb  *web.PermissionWeb
	authWeb        *web.AuthWeb
	authMiddleware *middleware.AuthMiddleware
	jobHistoryWeb  *web.JobHistoryWeb
	jwtHandler     jwtX2.JwtHandlerx

	// 任务执行引擎
	scheduler       *scheduler.CronScheduler
	executorFactory executor.ExecutorFactory
	funcExecutor    *executor.FunctionExecutor
}

// NewCronMysql 创建CronMysql实例（带完整依赖注入）
func NewCronMysql(engine *gin.Engine, db *gorm.DB, redSync redsyncx.RedSyncIn, rdb redis.Cmdable, l logx.Loggerx, cfg *config.Config) *CronMysql {
	// JWT handler
	jwtHandler, err := jwtX2.NewJwtxMiddlewareGinx(rdb, &jwtX2.JwtxMiddlewareGinxConfig{
		JwtKey:                []byte(cfg.JWT.Secret),
		LongJwtKey:            []byte(cfg.JWT.LongSecret),
		DurationExpiresIn:     cfg.JWT.AccessTTL,
		LongDurationExpiresIn: cfg.JWT.RefreshTTL,
	})
	if err != nil {
		panic("初始化JWT失败: " + err.Error())
	}

	// DAO层
	cronDb := dao.NewCronDb(db)
	jobHistoryDao := dao.NewJobHistoryDAO(db)
	deptDb := dao.NewDepartmentDb(db)
	userDb := dao.NewUserDb(db)
	roleDb := dao.NewRoleDb(db)
	permDb := dao.NewPermissionDb(db)
	userRoleDb := dao.NewUserRoleDb(db)
	rolePermDb := dao.NewRolePermissionDb(db)
	cronPermDb := dao.NewCronPermissionDb(db)

	// Repository层
	cronRepo := repository.NewCronRepository(cronDb)
	jobHistoryRepo := repository.NewJobHistoryRepository(jobHistoryDao)
	deptRepo := repository.NewDepartmentRepository(deptDb)
	userRepo := repository.NewUserRepository(userDb, userRoleDb, permDb)
	roleRepo := repository.NewRoleRepository(roleDb, rolePermDb)
	permRepo := repository.NewPermissionRepository(permDb)
	authRepo := repository.NewAuthRepository(userRoleDb, cronPermDb)

	// Service层
	cronSvc := service.NewCronService(cronRepo, nil)
	jobHistorySvc := service.NewJobHistoryService(jobHistoryRepo)
	deptSvc := service.NewDepartmentService(deptRepo)
	userSvc := service.NewUserService(userRepo)
	roleSvc := service.NewRoleService(roleRepo)
	permSvc := service.NewPermissionService(permRepo)
	authSvc := service.NewAuthService(authRepo, userRepo)

	// Web层
	cronWeb := web.NewCronWeb(cronSvc, l)
	jobHistoryWeb := web.NewJobHistoryWeb(jobHistorySvc, l)
	deptWeb := web.NewDepartmentWeb(deptSvc, l)
	userWeb := web.NewUserWeb(userSvc, jwtHandler, l)
	roleWeb := web.NewRoleWeb(roleSvc, l)
	permWeb := web.NewPermissionWeb(permSvc, l)
	authWebInst := web.NewAuthWeb(authSvc, l)

	// 中间件
	authMiddleware := middleware.NewAuthMiddleware(authSvc, jwtHandler, l)

	// 任务执行引擎
	executorFactory := executor.NewExecutorFactoryWithHistory(jobHistorySvc, l).(*executor.DefaultExecutorFactory)
	funcExec := executor.NewFunctionExecutor(l)
	httpExec := executor.NewHTTPExecutor(l)
	grpcExec := executor.NewGRPCExecutor(l)
	executorFactory.RegisterExecutor(funcExec)
	executorFactory.RegisterExecutor(httpExec)
	executorFactory.RegisterExecutor(grpcExec)
	// 注入执行器工厂到Service（用于创建任务时校验）
	cronSvc.SetTaskValidator(executorFactory)


	// 创建调度器
	sched := scheduler.NewCronScheduler(cronSvc, executorFactory, redSync, l)
	cronSvc.SetScheduler(sched)

	return &CronMysql{
		web:             engine,
		db:              db,
		redSync:         redSync,
		rdb:             rdb,
		l:               l,
		cfg:             cfg,
		cronWeb:         cronWeb,
		departmentWeb:   deptWeb,
		userWeb:         userWeb,
		roleWeb:         roleWeb,
		permissionWeb:   permWeb,
		authWeb:         authWebInst,
		authMiddleware:  authMiddleware,
		jobHistoryWeb:   jobHistoryWeb,
		jwtHandler:      jwtHandler,
		scheduler:       sched,
		executorFactory: executorFactory,
		funcExecutor:    funcExec,
	}
}

// RegisterRoutes 注册所有路由
func (c *CronMysql) RegisterRoutes() {
	// 公开路由（不需要认证）—— 仅限登录接口
	publicGroup := c.web.Group("/user")
	{
		publicGroup.POST("/login", c.userWeb.Login)
	}

	// 需要登录的路由
	authorized := c.web.Group("")
	authorized.Use(c.authMiddleware.RequireLogin())
	{
		// 任务管理（需要cron:read权限）
		cronReadGroup := authorized.Group("/cron")
		cronReadGroup.Use(c.authMiddleware.RequirePermission("cron:read"))
		{
			cronReadGroup.GET("/find/:cron_id", c.cronWeb.FindId)
			cronReadGroup.GET("/profile", c.cronWeb.FindAll)
		}

		cronCreateGroup := authorized.Group("/cron")
		cronCreateGroup.Use(c.authMiddleware.RequirePermission("cron:create"))
		{
			cronCreateGroup.POST("/add", c.cronWeb.Add)
			cronCreateGroup.POST("/adds", c.cronWeb.Adds)
		}

		cronDeleteGroup := authorized.Group("/cron")
		cronDeleteGroup.Use(c.authMiddleware.RequirePermission("cron:delete"))
		{
			cronDeleteGroup.DELETE("/delete/:cron_id", c.cronWeb.Delete)
			cronDeleteGroup.DELETE("/deletes/", c.cronWeb.Deletes)
		}

		// 任务状态管理（需要cron:manage权限）
		cronManageGroup := authorized.Group("/cron")
		cronManageGroup.Use(c.authMiddleware.RequirePermission("cron:manage"))
		{
			cronManageGroup.PUT("/start/:cron_id", c.cronWeb.StartJob)
			cronManageGroup.PUT("/pause/:cron_id", c.cronWeb.PauseJob)
			cronManageGroup.PUT("/resume/:cron_id", c.cronWeb.ResumeJob)
		}

		// 任务执行历史（需要cron:read权限）
		historyGroup := authorized.Group("/job-history")
		historyGroup.Use(c.authMiddleware.RequirePermission("cron:read"))
		{
			c.jobHistoryWeb.Register(historyGroup)
		}

		// 部门管理（需要dept:read权限）
		deptGroup := authorized.Group("/department")
		deptGroup.Use(c.authMiddleware.RequirePermission("dept:read"))
		{
			deptGroup.GET("/get/:dept_id", c.departmentWeb.GetDepartment)
			deptGroup.GET("/list", c.departmentWeb.GetAllDepartments)
			deptGroup.GET("/sub/:parent_id", c.departmentWeb.GetSubDepartments)
		}

		deptManageGroup := authorized.Group("/department")
		deptManageGroup.Use(c.authMiddleware.RequirePermission("dept:manage"))
		{
			deptManageGroup.POST("/create", c.departmentWeb.CreateDepartment)
			deptManageGroup.PUT("/update", c.departmentWeb.UpdateDepartment)
			deptManageGroup.DELETE("/delete/:dept_id", c.departmentWeb.DeleteDepartment)
		}

		// 角色和权限管理（需要admin权限）—— 注册到adminGroup，而非root engine
		adminGroup := authorized.Group("")
		adminGroup.Use(c.authMiddleware.RequirePermission("admin"))
		{
			c.roleWeb.Register(adminGroup)
			c.permissionWeb.Register(adminGroup)
			c.authWeb.Register(adminGroup)
		}

		// 用户自身操作（仅需登录）
		userGroup := authorized.Group("/user")
		{
			userGroup.POST("/change-password", c.userWeb.ChangePassword)
			userGroup.POST("/logout", c.userWeb.Logout)
			userGroup.GET("/profile", c.userWeb.GetProfile)
		}

		// 用户管理操作（需要user:manage权限）
		userManageGroup := authorized.Group("/user")
		userManageGroup.Use(c.authMiddleware.RequirePermission("user:manage"))
		{
			userManageGroup.POST("/create", c.userWeb.CreateUser)
			userManageGroup.GET("/get/:user_id", c.userWeb.GetUser)
			userManageGroup.GET("/list", c.userWeb.GetAllUsers)
			userManageGroup.GET("/dept/:dept_id", c.userWeb.GetUsersByDepartment)
			userManageGroup.PUT("/update", c.userWeb.UpdateUser)
			userManageGroup.DELETE("/delete/:user_id", c.userWeb.DeleteUser)
			userManageGroup.GET("/roles/:user_id", c.userWeb.GetUserRoles)
			userManageGroup.GET("/permissions/:user_id", c.userWeb.GetUserPermissions)
		}
	}
}

// AutoMigrate 自动迁移数据库表
func (c *CronMysql) AutoMigrate() error {
	return c.db.AutoMigrate(
		&dao.CronJob{},
		&dao.JobHistory{},
		&dao.Department{},
		&dao.User{},
		&dao.Role{},
		&dao.Permission{},
		&dao.UserRole{},
		&dao.RolePermission{},
		&dao.CronPermission{},
	)
}

// Start 启动系统（执行初始化任务）
func (c *CronMysql) Start() error {
	c.l.Info("CronMysql系统正在启动...")

	// 自动迁移数据库
	if err := c.AutoMigrate(); err != nil {
		c.l.Error("数据库迁移失败", logx.Error(err))
		return err
	}

	// 注册路由
	c.RegisterRoutes()

	// 启动任务调度器
	if err := c.scheduler.Start(); err != nil {
		c.l.Error("启动任务调度器失败", logx.Error(err))
		return err
	}

	c.l.Info("CronMysql系统启动完成")
	return nil
}

// Stop 停止系统
func (c *CronMysql) Stop() {
	c.l.Info("CronMysql系统正在停止...")
	c.scheduler.Stop()
	c.l.Info("CronMysql系统已停止")
}

// GetScheduler 获取调度器（用于动态管理任务）
func (c *CronMysql) GetScheduler() *scheduler.CronScheduler {
	return c.scheduler
}

// RegisterFunction 注册业务函数（供外部调用，在 Start 前注册）
// name: 函数名，与创建任务时 description 中的 function_name 对应
// fn: 函数实现，接收参数 map，返回结果和错误
func (c *CronMysql) RegisterFunction(name string, fn func(context.Context, map[string]interface{}) (interface{}, error)) {
	c.funcExecutor.RegisterFunction(name, fn)
}
