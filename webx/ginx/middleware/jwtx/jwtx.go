package jwtx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JwtxMiddlewareGinx struct {
	SigningMethod jwt.SigningMethod
	du            time.Duration
}

func NewJwtxMiddlewareGinx(du time.Duration) JwtHandlerx {
	return &JwtxMiddlewareGinx{
		SigningMethod: jwt.SigningMethodHS512,
		du:            du,
	}
}

// SetToken 设置JwtToken
func (j *JwtxMiddlewareGinx) SetToken(ctx *gin.Context, userId int64, jwtKey []byte, biz ...map[string]any) (string, error) {
	//ok := slices.Contains(biz, "User-Agent")
	//if ok {
	//	ctx.GetHeader("User-Agent") // 获取用户代理
	//}

	uc := UserClaims{
		Uid: userId,
		//Ssid:      ssid,                        // 登录唯一标识
		//UserAgent: ctx.GetHeader("User-Agent"), // 获取用户代理
		biz: biz,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.du))},
	}
	token := jwt.NewWithClaims(j.SigningMethod, uc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	return token.SignedString(jwtKey)               // tokenStr是加密后的token字符串
}

func (j *JwtxMiddlewareGinx) VerifyToken(ctx *gin.Context, token string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (j *JwtxMiddlewareGinx) RefreshToken(ctx *gin.Context, token string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (j *JwtxMiddlewareGinx) DeleteToken(ctx *gin.Context, token string) error {
	//TODO implement me
	panic("implement me")
}

// LoginJWT 登录【JWT方式实现：json-web-token】
type UserClaims struct {
	jwt.RegisteredClaims       // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
	Uid                  int64 // 用户id
	//Ssid                 string // 登录唯一标识
	//UserAgent            string // 用户代理
	biz []map[string]any // 业务标识、登录唯一标识、用户代理
}

var JWTKey = []byte("rUwYX9LZXU0Vjiizkhzmj8VyBd3GcwrC")   // JWTKey是加密的key，需要保密，不能泄露
var RCJWTKey = []byte("rUwYX9LZXU0Vjiizkhzmj8VyBd3GcwrF") // JWTKey是加密的key，需要保密，不能泄露
