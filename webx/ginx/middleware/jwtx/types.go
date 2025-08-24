package jwtx

import "github.com/gin-gonic/gin"

type JwtHandlerx interface {
	// SetToken 生成 JwtToken【其他有关业务，都可以放入这个 biz []string】
	SetToken(ctx *gin.Context, userId int64, jwtKey []byte, biz ...map[string]any) (string, error)
	// VerifyToken 验证 JwtToken
	VerifyToken(ctx *gin.Context, token string) (int64, error)
	// RefreshToken 刷新 JwtToken 过期时间
	RefreshToken(ctx *gin.Context, token string) (string, error)
	// DeleteToken 删除 JwtToken
	DeleteToken(ctx *gin.Context, token string) error
}
