package jwtx

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type JwtxMiddlewareGinxConfig struct {
	SigningMethod         jwt.SigningMethod // 默认HS512加密方式jwt.SigningMethodHS512
	DurationExpiresIn     time.Duration     // 默认30分钟
	LongDurationExpiresIn time.Duration     // 默认7天
	JwtKey                []byte            // 【必传】
	LongJwtKey            []byte            // 【必传】
	HeaderJwtTokenKey     string            // 默认jwt-token
	LongHeaderJwtTokenKey string            // 默认long-jwt-token
}

type JwtxMiddlewareGinx struct {
	cache redis.Cmdable
	JwtxMiddlewareGinxConfig
}

// Deprecated: jwtx此包弃用，此方法将在未来版本中删除，请使用jwtX2包【可无缝替换jwtX2包实现】
// NewJwtxMiddlewareGinx 创建JwtxMiddlewareGinx
//   - 【一般情况下，只用设置、验证、刷新、删除四个token方法】
//   - expiresIn: token过期时间
//   - jwtKey: 密钥
func NewJwtxMiddlewareGinx(cache redis.Cmdable, jwtConf *JwtxMiddlewareGinxConfig) JwtHandlerx {
	if jwtConf.SigningMethod == nil {
		jwtConf.SigningMethod = jwt.SigningMethodHS512
	}
	if jwtConf.DurationExpiresIn <= 0 {
		jwtConf.DurationExpiresIn = time.Minute * 30
	}
	if jwtConf.LongDurationExpiresIn <= 0 {
		jwtConf.LongDurationExpiresIn = time.Hour * 24 * 7
	}
	if jwtConf.LongHeaderJwtTokenKey == "" {
		jwtConf.LongHeaderJwtTokenKey = "long-jwt-token"
	}
	if jwtConf.HeaderJwtTokenKey == "" {
		jwtConf.HeaderJwtTokenKey = "jwt-token"
	}

	return &JwtxMiddlewareGinx{
		cache:                    cache,
		JwtxMiddlewareGinxConfig: *jwtConf,
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.DurationExpiresIn))},
	}
	token := jwt.NewWithClaims(j.SigningMethod, uc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	ctx.Set("user", uc)
	tokenStr, err := token.SignedString(j.JwtKey) // tokenStr是加密后的token字符串
	if err != nil {
		var u UserClaims
		return &u, err
	}
	err = j.cache.Set(ctx, "user:token:info:"+fmt.Sprintf("%d", userId), tokenStr, j.DurationExpiresIn).Err()
	if err != nil {
		var u UserClaims
		return &u, err
	}

	reUc := RefreshUserClaims{
		Uid:       userId,
		Name:      name,
		Ssid:      ssid,                        // 登录唯一标识
		UserAgent: ctx.GetHeader("User-Agent"), // 获取用户代理
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.LongDurationExpiresIn))},
	}
	longToken := jwt.NewWithClaims(j.SigningMethod, reUc) // jwt.SigningMethodES512是加密方式，默认是HS256，返回token是结构体
	ctx.Set("userLong", reUc)
	longTokenStr, err := longToken.SignedString(j.LongJwtKey) // tokenStr是加密后的token字符串
	if err != nil {
		var u UserClaims
		return &u, err
	}
	err = j.cache.Set(ctx, "user:longToken:info:"+fmt.Sprintf("%d", userId), longTokenStr, j.LongDurationExpiresIn).Err()
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
	t, err := jwt.ParseWithClaims(tokenStr, uc,
		func(token *jwt.Token) (interface{}, error) {
			return j.JwtKey, nil
		},
	)
	// 验证token，t.Valid是验证token，t.Valid是bool类型，true表示验证成功，false表示验证失败
	if t == nil || err != nil || !t.Valid {
		//ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		//ctx.Abort() // 阻止继续执行
		return uc, fmt.Errorf("invalid token, token无效/伪造的token %v", err)
	}

	// 验证redis中的 token
	rdGet, err := j.cache.Get(ctx, "user:token:info:"+fmt.Sprintf("%d", uc.Uid)).Result()
	if err != nil || tokenStr != rdGet {
		return uc, fmt.Errorf("invalid token, token无效/伪造的token %v", err)
	}
	ctx.Set("user", uc)
	return uc, nil
}

// LongVerifyToken 验证长JwtToken【一般是刷新token时，此方法验证长token，生成新的长短token】
func (j *JwtxMiddlewareGinx) LongVerifyToken(ctx *gin.Context) (*RefreshUserClaims, error) {
	tokenStr := j.ExtractToken(ctx)
	uc := &RefreshUserClaims{}
	t, err := jwt.ParseWithClaims(tokenStr, uc,
		func(token *jwt.Token) (interface{}, error) {
			return j.LongJwtKey, nil
		},
	)
	// 验证token，t.Valid是验证token，t.Valid是bool类型，true表示验证成功，false表示验证失败
	if t == nil || err != nil || !t.Valid {
		//ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		//ctx.Abort() // 阻止继续执行.
		return uc, fmt.Errorf("invalid token, token无效/伪造的token %v", err)
	}
	// 验证redis中的 token
	rdGet, err := j.cache.Get(ctx, "user:longToken:info:"+fmt.Sprintf("%d", uc.Uid)).Result()
	if err != nil || tokenStr != rdGet {
		return uc, fmt.Errorf("invalid token, 长token无效/伪造的token %v", err)
	}
	ctx.Set("userLong", uc)
	return uc, nil
}

// RefreshToken 刷新JwtToken【当用户操作时，直接刷新token，刷新前验证token】
func (j *JwtxMiddlewareGinx) RefreshToken(ctx *gin.Context, ssid string) (*UserClaims, error) {
	// 验证长token，确保本次请求合法【一般gin的middleware中间件校验】
	uc, err := j.LongVerifyToken(ctx)
	if err != nil {
		//return uc, err
		return &UserClaims{}, err
	}

	// 合法请求，删除原token
	ctx.Header(j.HeaderJwtTokenKey, "")
	ctx.Header(j.LongHeaderJwtTokenKey, "")
	// 删除Redis中的用户信息使用
	err = j.cache.Del(ctx, "user:token:info:"+fmt.Sprintf("%d", uc.Uid)).Err()
	if err != nil {
		var u *UserClaims
		return u, fmt.Errorf("delete redis token info error: %v", err)
	}
	err = j.cache.Del(ctx, "user:longToken:info:"+fmt.Sprintf("%d", uc.Uid)).Err()

	// 重新设置token
	//ssid := uuid.New().String()
	return j.SetToken(ctx, uc.Uid, uc.Name, ssid)
}

// DeleteToken 删除JwtToken【多用于退出登录~】
func (j *JwtxMiddlewareGinx) DeleteToken(ctx *gin.Context) (*UserClaims, error) {
	uc, err := j.VerifyToken(ctx)
	if err != nil {
		return uc, err
	}

	ctx.Header(j.HeaderJwtTokenKey, "")
	ctx.Header(j.LongHeaderJwtTokenKey, "")

	// 删除Redis中的用户信息使用
	err = j.cache.Del(ctx, "user:token:info:"+fmt.Sprintf("%d", uc.Uid)).Err()
	if err != nil {
		return uc, fmt.Errorf("delete redis token info error: %v", err)
	}
	return uc, j.cache.Del(ctx, "user:longToken:info:"+fmt.Sprintf("%d", uc.Uid)).Err()
}

// UserClaims  登录【JWT方式实现：json-web-token】
type UserClaims struct {
	jwt.RegisteredClaims        // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
	Uid                  int64  // 用户id
	Name                 string // 用户名
	Ssid                 string // 登录唯一标识
	UserAgent            string // 用户代理
}
type RefreshUserClaims struct {
	jwt.RegisteredClaims        // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
	Uid                  int64  // 用户id
	Name                 string // 用户名
	Ssid                 string // 登录唯一标识
	UserAgent            string // 用户代理
}
