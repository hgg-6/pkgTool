package jwtx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"strings"
	"time"
)

type JwtxMiddlewareGinxConfig struct {
	SigningMethod         jwt.SigningMethod // 默认HS512加密方式jwt.SigningMethodHS512
	ExpiresIn             time.Duration     // 默认10分钟
	LongExpiresIn         time.Duration     // 默认7天
	JwtKey                []byte            // 【必传】
	LongJwtKey            []byte            // 【必传】
	HeaderJwtTokenKey     string            // 默认jwt-token
	LongHeaderJwtTokenKey string            // 默认long-jwt-token
}

type JwtxMiddlewareGinx struct {
	JwtxMiddlewareGinxConfig
}

// NewJwtxMiddlewareGinx 创建JwtxMiddlewareGinx
//   - 【一般情况下，只用设置、验证、刷新、删除四个token方法】
//   - expiresIn: token过期时间
//   - jwtKey: 密钥
func NewJwtxMiddlewareGinx(jwtConf JwtxMiddlewareGinxConfig) JwtHandlerx {
	jwtConf.SigningMethod = jwt.SigningMethodHS512
	jwtConf.HeaderJwtTokenKey = "jwt-token"
	jwtConf.LongHeaderJwtTokenKey = "long-jwt-token"
	jwtConf.ExpiresIn = time.Minute * 20
	jwtConf.LongExpiresIn = time.Hour * 24 * 7
	return &JwtxMiddlewareGinx{
		JwtxMiddlewareGinxConfig: JwtxMiddlewareGinxConfig{
			SigningMethod:         jwtConf.SigningMethod,
			ExpiresIn:             jwtConf.ExpiresIn,
			JwtKey:                jwtConf.JwtKey,
			LongJwtKey:            jwtConf.LongJwtKey,
			HeaderJwtTokenKey:     jwtConf.HeaderJwtTokenKey,
			LongHeaderJwtTokenKey: jwtConf.LongHeaderJwtTokenKey,
		},
		//SigningMethod:     jwt.SigningMethodHS512,
		//ExpiresIn:         expiresIn,
		//JwtKey:            jwtKey,
		//LongJwtKey:        longJwtKey,
		//HeaderJwtTokenKey: "jwt-token",
	}
}

// SetToken 设置JwtToken【ssid构造一般可以 ssid := uuid.New().String() 来生成随机数【长token】】
//   - SetToken一般登录时设置调用
//   - 一般情况--登录设置长短token--》验证token【会获取token验证】--》刷新token--》删除token--》退出登录
func (j *JwtxMiddlewareGinx) SetToken(ctx *gin.Context, userId int64, name string, ssid string) (*UserClaims, error) {
	//ok := slices.Contains(biz, "User-Agent")
	//if ok {
	//	ctx.GetHeader("User-Agent") // 获取用户代理
	//}

	uc := UserClaims{
		Uid:       userId,
		Name:      name,
		Ssid:      ssid,                        // 登录唯一标识
		UserAgent: ctx.GetHeader("User-Agent"), // 获取用户代理
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ExpiresIn))},
	}
	token := jwt.NewWithClaims(j.SigningMethod, uc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	ctx.Set("user", uc)
	tokenStr, err := token.SignedString(j.JwtKey) // tokenStr是加密后的token字符串
	if err != nil {
		var u UserClaims
		return &u, err
	}

	uc = UserClaims{
		Uid:       userId,
		Name:      name,
		Ssid:      ssid,                        // 登录唯一标识
		UserAgent: ctx.GetHeader("User-Agent"), // 获取用户代理
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.LongExpiresIn))},
	}
	longToken := jwt.NewWithClaims(j.SigningMethod, uc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	ctx.Set("userLong", uc)
	longTokenStr, err := longToken.SignedString(j.LongJwtKey) // tokenStr是加密后的token字符串
	if err != nil {
		var u UserClaims
		return &u, err
	}

	ctx.Header(j.HeaderJwtTokenKey, tokenStr)
	ctx.Header(j.LongHeaderJwtTokenKey, longTokenStr)
	return &uc, nil
}

// ExtractToken 获取JwtToken
func (j *JwtxMiddlewareGinx) ExtractToken(ctx *gin.Context) string {
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

// VerifyToken 验证JwtToken
func (j *JwtxMiddlewareGinx) VerifyToken(ctx *gin.Context) (*UserClaims, error) {
	tokenStr := j.ExtractToken(ctx)
	// 解析token
	//var uc *UserClaims
	//uc = &UserClaims{}
	uc := &UserClaims{}
	t, err := jwt.ParseWithClaims(tokenStr, uc, func(token *jwt.Token) (interface{}, error) {
		return j.JwtKey, nil
	})
	// 验证token，t.Valid是验证token，t.Valid是bool类型，true表示验证成功，false表示验证失败
	if t == nil || err != nil || !t.Valid {
		//ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		//ctx.Abort() // 阻止继续执行
		return uc, errors.New("invalid token, token无效/伪造的token")
	}
	ctx.Set("user", uc)
	return uc, nil
}

// RefreshToken 刷新JwtToken【当用户操作时，直接刷新token，刷新前验证token】
func (j *JwtxMiddlewareGinx) RefreshToken(ctx *gin.Context) (*UserClaims, error) {
	// 验证token，确保本次请求合法
	uc, err := j.VerifyToken(ctx)
	if err != nil {
		return uc, err
	}
	// 重新设置token
	ssid := uuid.New().String()
	return j.SetToken(ctx, uc.Uid, uc.Name, ssid)
}

// DeleteToken 删除JwtToken
func (j *JwtxMiddlewareGinx) DeleteToken(ctx *gin.Context) (*UserClaims, error) {
	ctx.Header(j.HeaderJwtTokenKey, "")
	uc, ok := ctx.MustGet("user").(UserClaims) // 获取用户信息，断言
	if !ok {
		return &UserClaims{}, fmt.Errorf("user claims not found, 请求头中没有找到用户信息")
	}
	// 【uc】是删除Redis中的用户信息使用
	//return h.client.Set(ctx, fmt.Sprintf("user:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
	return &uc, nil
}

// UserClaims  登录【JWT方式实现：json-web-token】
type UserClaims struct {
	jwt.RegisteredClaims        // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
	Uid                  int64  // 用户id
	Name                 string // 用户名
	Ssid                 string // 登录唯一标识
	UserAgent            string // 用户代理
	//biz map[string]any // 业务标识、登录唯一标识、用户代理
}
