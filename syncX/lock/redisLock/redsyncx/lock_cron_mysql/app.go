package lock_cron_mysql

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/executor"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/middleware"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/scheduler"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CronMysql 定时任务系统主类
type CronMysql struct {
	web     *gin.Engine
	db      *gorm.DB
	redSync redsyncx.RedSyncIn
	l       logx.Loggerx

	// 各层组件
	cronWeb        *web.CronWeb
	departmentWeb  *web.DepartmentWeb
	userWeb        *web.UserWeb
	roleWeb        *web.RoleWeb
	permissionWeb  *web.PermissionWeb
	authWeb        *web.AuthWeb
	authMiddleware *middleware.AuthMiddleware
	jobHistoryWeb  *web.JobHistoryWeb // 添加任务历史Web处理器

	// 任务执行引擎
	scheduler       *scheduler.CronScheduler
	executorFactory executor.ExecutorFactory
}

// NewCronMysql 创建CronMysql实例（带完整依赖注入）
func NewCronMysql(engine *gin.Engine, db *gorm.DB, redSync redsyncx.RedSyncIn, l logx.Loggerx) *CronMysql {
	// DAO层
	cronDb := dao.NewCronDb(db)
	jobHistoryDao := dao.NewJobHistoryDAO(db) // 添加任务历史DAO
	deptDb := dao.NewDepartmentDb(db)
	userDb := dao.NewUserDb(db)
	roleDb := dao.NewRoleDb(db)
	permDb := dao.NewPermissionDb(db)
	userRoleDb := dao.NewUserRoleDb(db)
	rolePermDb := dao.NewRolePermissionDb(db)
	cronPermDb := dao.NewCronPermissionDb(db)

	// Repository层
	cronRepo := repository.NewCronRepository(cronDb)
	jobHistoryRepo := repository.NewJobHistoryRepository(jobHistoryDao) // 添加任务历史Repository
	deptRepo := repository.NewDepartmentRepository(deptDb)
	userRepo := repository.NewUserRepository(userDb, userRoleDb, permDb)
	roleRepo := repository.NewRoleRepository(roleDb, rolePermDb)
	permRepo := repository.NewPermissionRepository(permDb)
	authRepo := repository.NewAuthRepository(userRoleDb, cronPermDb)

	// Service层
	cronSvc := service.NewCronService(cronRepo)
	jobHistorySvc := service.NewJobHistoryService(jobHistoryRepo) // 添加任务历史Service
	deptSvc := service.NewDepartmentService(deptRepo)
	userSvc := service.NewUserService(userRepo)
	roleSvc := service.NewRoleService(roleRepo)
	permSvc := service.NewPermissionService(permRepo)
	authSvc := service.NewAuthService(authRepo, userRepo)

	// Web层
	cronWeb := web.NewCronWeb(cronSvc, l)
	jobHistoryWeb := web.NewJobHistoryWeb(jobHistorySvc, l) // 添加任务历史Web
	deptWeb := web.NewDepartmentWeb(deptSvc, l)
	userWeb := web.NewUserWeb(userSvc, l)
	roleWeb := web.NewRoleWeb(roleSvc, l)
	permWeb := web.NewPermissionWeb(permSvc, l)
	authWebInst := web.NewAuthWeb(authSvc, l)

	// 中间件
	authMiddleware := middleware.NewAuthMiddleware(authSvc, l)

	// 任务执行引擎
	executorFactory := executor.NewExecutorFactoryWithHistory(jobHistorySvc, l).(*executor.DefaultExecutorFactory) // 使用带历史记录的工厂
	// 注册各类执行器
	funcExec := executor.NewFunctionExecutor(l)
	httpExec := executor.NewHTTPExecutor(l)
	grpcExec := executor.NewGRPCExecutor(l)
	executorFactory.RegisterExecutor(funcExec)
	executorFactory.RegisterExecutor(httpExec)
	executorFactory.RegisterExecutor(grpcExec)

	// 创建调度器
	scheduler := scheduler.NewCronScheduler(cronSvc, executorFactory, redSync, l)

	return &CronMysql{
		web:             engine,
		db:              db,
		redSync:         redSync,
		l:               l,
		cronWeb:         cronWeb,
		departmentWeb:   deptWeb,
		userWeb:         userWeb,
		roleWeb:         roleWeb,
		permissionWeb:   permWeb,
		authWeb:         authWebInst,
		authMiddleware:  authMiddleware,
		jobHistoryWeb:   jobHistoryWeb, // 添加任务历史Web
		scheduler:       scheduler,
		executorFactory: executorFactory,
	}
}

// RegisterRoutes 注册所有路由
func (c *CronMysql) RegisterRoutes() {
	// 注册公开路由（不需要认证）
	c.userWeb.Register(c.web) // 包含登录接口

	// 需要登录的路由
	authorized := c.web.Group("")
	authorized.Use(c.authMiddleware.RequireLogin())
	{
		// 任务管理（需要cron权限）
		cronGroup := authorized.Group("/cron")
		cronGroup.Use(c.authMiddleware.RequirePermission("cron:read"))
		{
			cronGroup.GET("/find/:cron_id", c.cronWeb.FindId)
			cronGroup.GET("/profile", c.cronWeb.FindAll)
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

		// 任务执行历史（需要cron权限）
		historyGroup := authorized.Group("/job-history")
		historyGroup.Use(c.authMiddleware.RequirePermission("cron:read"))
		{
			c.jobHistoryWeb.Register(historyGroup)
		}

		// 部门管理（需要dept权限）
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

		// 角色和权限管理（需要admin权限）
		adminGroup := authorized.Group("")
		adminGroup.Use(c.authMiddleware.RequirePermission("admin"))
		{
			c.roleWeb.Register(c.web)
			c.permissionWeb.Register(c.web)
			c.authWeb.Register(c.web)
		}
	}
}

// AutoMigrate 自动迁移数据库表
func (c *CronMysql) AutoMigrate() error {
	return c.db.AutoMigrate(
		&dao.CronJob{},
		&dao.JobHistory{}, // 添加任务执行历史表
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
