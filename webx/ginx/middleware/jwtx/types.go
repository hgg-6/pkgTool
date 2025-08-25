package jwtx

import "github.com/gin-gonic/gin"

type JwtHandlerx interface {
	// SetToken 生成 JwtToken【其他有关业务，都可以放入这个 biz []string，可以为nil】
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
