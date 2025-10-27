package jwtX2

import "github.com/gin-gonic/gin"

//go:generate mockgen -source=./types.go -package=jwtxmocks -destination=mocks/jwtHdl.mock.go JwtHandlerx
type JwtHandlerx interface {
	// SetToken 登录时设置长短 Token
	SetToken(ctx *gin.Context, userId int64, name string, ssid string) (*UserClaims, error)
	// ExtractToken 从 Authorization: Bearer <token> 中提取 token
	ExtractToken(ctx *gin.Context) string
	// VerifyToken 验证 Access Token（短 Token）
	VerifyToken(ctx *gin.Context) (*UserClaims, error)
	// LongVerifyToken 验证 Refresh Token（长 Token）
	LongVerifyToken(ctx *gin.Context) (*UserClaims, error)
	// RefreshToken 刷新 Token，可选新 ssid（若为空则复用原 ssid）
	RefreshToken(ctx *gin.Context, newSsid string) (*UserClaims, error)
	// DeleteToken 退出登录：仅删除当前会话 Token
	DeleteToken(ctx *gin.Context) (*UserClaims, error)
}

//type JwtHandlerx interface {
//	// SetToken 登录时设置长短 Token
//	SetToken(ctx *gin.Context, userId int64, name string, ssid string) (*UserClaims, error)
//	// ExtractToken 从 Authorization: Bearer <token> 中提取 token
//	ExtractToken(ctx *gin.Context) string
//	// VerifyToken 验证 Access Token（短 Token）
//	VerifyToken(ctx *gin.Context) (*UserClaims, error)
//	// LongVerifyToken 验证 Refresh Token（长 Token）
//	LongVerifyToken(ctx *gin.Context) (*UserClaims, error)
//	// RefreshToken 刷新 Token，可选新 ssid（若为空则复用原 ssid）
//	RefreshToken(ctx *gin.Context, newSsid string) (*UserClaims, error)
//	// DeleteToken 退出登录：仅删除当前会话 Token
//	DeleteToken(ctx *gin.Context) (*UserClaims, error)
//}
