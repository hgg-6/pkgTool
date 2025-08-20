package jwtx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lithammer/shortuuid/v4"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	client        redis.Cmdable // redis.Cmdable是redis的接口
	SigningMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

// NewRedisJWTHandler 基于redis的JWT处理器
func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:        client,
		SigningMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, fmt.Sprintf("user:ssid:%s", ssid)).Result() // 获取token
	if err != nil {
		return err
	} else if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

// ExtractToken 获取token,根据约定，token在Authorization头部
// Bearer XXXXX
func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization") // 获取请求头中的Authorization字段
	if authCode == "" {
		return authCode
	}
	// 因为Authorization: Bearer XXXXX，【Bearer XXXXX中间有空格，需要切割】
	segs := strings.Split(authCode, " ") // 根据空格切割，得到 Bearer 和 token
	if len(segs) != 2 {                  // 一般只有一个空格，切开变成两段，如果切割出来的数组长度不等于2，说明有问题
		return ""
	}
	// 拿到token，拆成两端，token在第2段,token在下标为1的一段
	return segs[1]
}

// SetLoginToken 设置登录Token
func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()              // 生成随机数【长token】
	err := h.SetRefreshToken(ctx, uid, ssid) // 刷新长Token
	if err != nil {
		return err
	}
	return h.SetJWTToken(ctx, uid, ssid) // 设置短JWT Token
}

// ClearToken 删除Token
func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims) // 获取用户信息，断言
	fmt.Println("DEBUG: ", uc)
	return h.client.Set(ctx, fmt.Sprintf("user:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
}

// SetJWTToken 设置JWTToken
func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	//err := h.SetRefreshToken(ctx, uid) // 刷新Token
	//if err != nil {
	//	ctx.String(http.StatusOK, "系统错误")
	//	return
	//}
	uc := UserClaims{
		Uid:       uid,                         // 用户id
		Ssid:      ssid,                        // 登录唯一标识
		UserAgent: ctx.GetHeader("User-Agent"), // 获取用户代理
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 1800))}, //600秒 10分钟 过期
	}
	token := jwt.NewWithClaims(h.SigningMethod, uc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	tokenStr, err := token.SignedString(JWTKey)     // tokenStr是加密后的token字符串
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr) // 设置头信息，返回给前端
	return nil
}

// SetRefreshToken 刷新Token时设置refreshToken
func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		}, // 7天过期
	}
	token := jwt.NewWithClaims(h.SigningMethod, rc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	tokenSrc, err := token.SignedString(RCJWTKey)   // tokenSrc是加密后的token字符串
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenSrc) // 设置头信息，返回给前端
	return nil
}

//// ExtractToken 根据约定，token在Authorization头部
//// Bearer XXXXX
//func ExtractToken(ctx *gin.Context) string { // 获取token
//	authCode := ctx.GetHeader("Authorization") // 获取请求头中的Authorization字段
//	if authCode == "" {
//		return authCode
//	}
//	// 因为Authorization: Bearer XXXXX，【Bearer XXXXX中间有空格，需要切割】
//	segs := strings.Split(authCode, " ") // 根据空格切割，得到 Bearer 和 token
//	if len(segs) != 2 {                  // 一般只有一个空格，切开变成两段，如果切割出来的数组长度不等于2，说明有问题
//		return ""
//	}
//	// 拿到token，因为拆成两端，token在第2段，所以真正的token在下标为1的一段
//	return segs[1]
//}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string // 登录唯一标识
}

// LoginJWT 登录【JWT方式实现：json-web-token】
type UserClaims struct {
	jwt.RegisteredClaims        // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
	Uid                  int64  // 用户id
	Ssid                 string // 登录唯一标识
	UserAgent            string // 用户代理
}

var JWTKey = []byte("rUwYX9LZXU0Vjiizkhzmj8VyBd3GcwrC")   // JWTKey是加密的key，需要保密，不能泄露
var RCJWTKey = []byte("rUwYX9LZXU0Vjiizkhzmj8VyBd3GcwrF") // JWTKey是加密的key，需要保密，不能泄露
