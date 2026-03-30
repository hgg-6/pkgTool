package middleware

import (
	"net/http"
	"strconv"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"

	"github.com/gin-gonic/gin"

	jwtX2 "gitee.com/hgg_test/pkg_tool/v2/webx/ginx/middleware/jwtX2"
)

// AuthMiddleware 权限验证中间件
type AuthMiddleware struct {
	authSvc    service.AuthService
	jwtHandler jwtX2.JwtHandlerx
	l          logx.Loggerx
}

// NewAuthMiddleware 创建AuthMiddleware实例
func NewAuthMiddleware(authSvc service.AuthService, jwtHandler jwtX2.JwtHandlerx, l logx.Loggerx) *AuthMiddleware {
	return &AuthMiddleware{
		authSvc:    authSvc,
		jwtHandler: jwtHandler,
		l:          l,
	}
}

// RequireLogin 要求用户已登录（验证JWT Token）
func (a *AuthMiddleware) RequireLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 通过JWT验证Token
		claims, err := a.jwtHandler.VerifyToken(ctx)
		if err != nil {
			a.l.Warn("用户未登录或Token无效", logx.Error(err))
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			ctx.Abort()
			return
		}

		// 将用户信息写入context
		ctx.Set("user_id", claims.Uid)
		ctx.Set("username", claims.Name)
		ctx.Set("ssid", claims.Ssid)
		ctx.Next()
	}
}

// RequirePermission 要求用户拥有指定权限
func (a *AuthMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从上下文中获取用户ID（由RequireLogin中间件设置）
		userIdValue, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("未找到用户ID")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		// 检查权限
		hasPermission, err := a.authSvc.CheckUserPermission(ctx.Request.Context(), userId, permissionCode)
		if err != nil {
			a.l.Error("检查权限失败", logx.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "内部错误"})
			ctx.Abort()
			return
		}

		if !hasPermission {
			a.l.Warn("用户权限不足", logx.Int64("user_id", userId), logx.String("permission", permissionCode))
			ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
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
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
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
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
				ctx.Abort()
				return
			}
			cronIdStr = strconv.FormatInt(req.CronId, 10)
		}

		cronId, err := strconv.ParseInt(cronIdStr, 10, 64)
		if err != nil {
			a.l.Error("任务ID格式错误", logx.Error(err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			ctx.Abort()
			return
		}

		// 检查用户所在部门是否有权限管理该任务
		hasPermission, err := a.authSvc.CheckCronPermission(ctx.Request.Context(), userId, cronId)
		if err != nil {
			a.l.Error("检查任务权限失败", logx.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "内部错误"})
			ctx.Abort()
			return
		}

		if !hasPermission {
			a.l.Warn("用户无权操作此任务", logx.Int64("user_id", userId), logx.Int64("cron_id", cronId))
			ctx.JSON(http.StatusForbidden, gin.H{"error": "无权操作此任务"})
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
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
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
		ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		ctx.Abort()
	}
}

// RequireAllPermissions 要求用户拥有所有指定权限
func (a *AuthMiddleware) RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIdValue, exists := ctx.Get("user_id")
		if !exists {
			a.l.Warn("未找到用户ID")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		userId, ok := userIdValue.(int64)
		if !ok {
			a.l.Error("用户ID类型错误")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
			return
		}

		// 检查是否拥有所有权限
		for _, permCode := range permissionCodes {
			hasPermission, err := a.authSvc.CheckUserPermission(ctx.Request.Context(), userId, permCode)
			if err != nil {
				a.l.Error("检查权限失败", logx.Error(err))
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "内部错误"})
				ctx.Abort()
				return
			}
			if !hasPermission {
				a.l.Warn("用户缺少必要权限", logx.Int64("user_id", userId), logx.String("permission", permCode))
				ctx.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
