package jwtx

import "github.com/gin-gonic/gin"

// JwtHandlerx 方法
//   - 一般情况下，只用登录、登出、验证、刷新四个token方法
type JwtHandlerx interface {
	// SetToken 生成 JwtToken
	SetToken(ctx *gin.Context, userId int64, name string, ssid string) (*UserClaims, error)
	// ExtractToken 获取 JwtToken
	ExtractToken(ctx *gin.Context) string
	// VerifyToken 验证 JwtToken
	VerifyToken(ctx *gin.Context) (*UserClaims, error)
	// RefreshToken 刷新 JwtToken 过期时间
	RefreshToken(ctx *gin.Context) (*UserClaims, error)
	// DeleteToken 删除 JwtToken
	DeleteToken(ctx *gin.Context) (*UserClaims, error)
}
