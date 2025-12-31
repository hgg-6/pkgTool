package middleware

import (
	"strconv"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 权限验证中间件
type AuthMiddleware struct {
	authSvc service.AuthService
	l       logx.Loggerx
}

// NewAuthMiddleware 创建AuthMiddleware实例
func NewAuthMiddleware(authSvc service.AuthService, l logx.Loggerx) *AuthMiddleware {
	return &AuthMiddleware{
		authSvc: authSvc,
		l:       l,
	}
}

// RequirePermission 要求用户拥有指定权限
func (a *AuthMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从上下文中获取用户ID（假设前面有JWT中间件设置了user_id）
		userIdValue, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("未找到用户ID")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		// 检查权限
		hasPermission, err := a.authSvc.CheckUserPermission(ctx.Request.Context(), userId, permissionCode)
		if err != nil {
			a.l.Error("检查权限失败", logx.Error(err))
			ctx.JSON(500, gin.H{"error": "内部错误"})
			ctx.Abort()
			return
		}

		if !hasPermission {
			a.l.Warn("用户权限不足", logx.Int64("user_id", userId), logx.String("permission", permissionCode))
			ctx.JSON(403, gin.H{"error": "权限不足"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireCronPermission 要求用户拥有指定任务的权限（基于部门）
func (a *AuthMiddleware) RequireCronPermission() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从上下文中获取用户ID
		userIdValue, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("未找到用户ID")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		// 从路径参数或请求体中获取cron_id
		cronIdStr := ctx.Param("cron_id")
		if cronIdStr == "" {
			// 尝试从JSON中获取
			var req struct {
				CronId int64 `json:"cron_id"`
			}
			if err := ctx.ShouldBindJSON(&req); err != nil {
				a.l.Error("获取任务ID失败", logx.Error(err))
				ctx.JSON(400, gin.H{"error": "参数错误"})
				ctx.Abort()
				return
			}
			// 将请求体回写到context，以便后续handler使用
			ctx.Set("cron_id", req.CronId)
			cronIdStr = strconv.FormatInt(req.CronId, 10)
		}

		cronId, err := strconv.ParseInt(cronIdStr, 10, 64)
		if err != nil {
			a.l.Error("任务ID格式错误", logx.Error(err))
			ctx.JSON(400, gin.H{"error": "参数错误"})
			ctx.Abort()
			return
		}

		// 检查用户所在部门是否有权限管理该任务
		hasPermission, err := a.authSvc.CheckCronPermission(ctx.Request.Context(), userId, cronId)
		if err != nil {
			a.l.Error("检查任务权限失败", logx.Error(err))
			ctx.JSON(500, gin.H{"error": "内部错误"})
			ctx.Abort()
			return
		}

		if !hasPermission {
			a.l.Warn("用户无权操作此任务", logx.Int64("user_id", userId), logx.Int64("cron_id", cronId))
			ctx.JSON(403, gin.H{"error": "无权操作此任务"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireLogin 要求用户已登录
func (a *AuthMiddleware) RequireLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 检查是否存在user_id（假设JWT中间件已设置）
		_, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("用户未登录")
			ctx.JSON(401, gin.H{"error": "请先登录"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireAnyPermission 要求用户拥有任意一个指定权限
func (a *AuthMiddleware) RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIdValue, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("未找到用户ID")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		// 检查是否拥有任意一个权限
		for _, permCode := range permissionCodes {
			hasPermission, err := a.authSvc.CheckUserPermission(ctx.Request.Context(), userId, permCode)
			if err != nil {
				a.l.Error("检查权限失败", logx.Error(err))
				continue
			}
			if hasPermission {
				ctx.Next()
				return
			}
		}

		a.l.Warn("用户权限不足", logx.Int64("user_id", userId))
		ctx.JSON(403, gin.H{"error": "权限不足"})
		ctx.Abort()
	}
}

// RequireAllPermissions 要求用户拥有所有指定权限
func (a *AuthMiddleware) RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIdValue, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("未找到用户ID")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(401, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		// 检查是否拥有所有权限
		for _, permCode := range permissionCodes {
			hasPermission, err := a.authSvc.CheckUserPermission(ctx.Request.Context(), userId, permCode)
			if err != nil {
				a.l.Error("检查权限失败", logx.Error(err))
				ctx.JSON(500, gin.H{"error": "内部错误"})
				ctx.Abort()
				return
			}
			if !hasPermission {
				a.l.Warn("用户缺少必要权限", logx.Int64("user_id", userId), logx.String("permission", permCode))
				ctx.JSON(403, gin.H{"error": "权限不足"})
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
