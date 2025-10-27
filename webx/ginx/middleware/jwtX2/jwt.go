package jwtX2

import (
	"crypto/sha256"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// JwtxMiddlewareGinxConfig 配置
//   - SigningMethod: 默认HS512加密方式jwt.SigningMethodHS512
//   - DurationExpiresIn: Access Token 过期时间，默认30分钟
//   - LongDurationExpiresIn: Refresh Token 过期时间，默认7天
//   - JwtKey: 密钥【必传】
//   - LongJwtKey: 密钥【必传】
//   - HeaderJwtTokenKey: Access Token Header Key，默认jwt-token
//   - LongHeaderJwtTokenKey: Refresh Token Header Key，默认long-jwt-token
type JwtxMiddlewareGinxConfig struct {
	SigningMethod         jwt.SigningMethod // 默认HS512加密方式jwt.SigningMethodHS512
	DurationExpiresIn     time.Duration     // Access Token 过期时间
	LongDurationExpiresIn time.Duration     // Refresh Token 过期时间
	JwtKey                []byte            // Access Token 密钥
	LongJwtKey            []byte            // Refresh Token 密钥
	HeaderJwtTokenKey     string            // Access Token Header Key，默认jwt-token
	LongHeaderJwtTokenKey string            // Refresh Token Header Key，默认long-jwt-token
}

type JwtxMiddlewareGinx struct {
	cache redis.Cmdable
	cfg   JwtxMiddlewareGinxConfig
}

func NewJwtxMiddlewareGinx(cache redis.Cmdable, jwtConf *JwtxMiddlewareGinxConfig) JwtHandlerx {
	if jwtConf == nil {
		panic("jwtConf must not be nil")
	}
	if len(jwtConf.JwtKey) == 0 || len(jwtConf.LongJwtKey) == 0 {
		panic("JwtKey and LongJwtKey must not be empty")
	}
	if jwtConf.SigningMethod == nil {
		jwtConf.SigningMethod = jwt.SigningMethodHS512
	}
	if jwtConf.DurationExpiresIn <= 0 {
		jwtConf.DurationExpiresIn = 30 * time.Minute
	}
	if jwtConf.LongDurationExpiresIn <= 0 {
		jwtConf.LongDurationExpiresIn = 7 * 24 * time.Hour
	}
	if jwtConf.HeaderJwtTokenKey == "" {
		jwtConf.HeaderJwtTokenKey = "jwt-token"
	}
	if jwtConf.LongHeaderJwtTokenKey == "" {
		jwtConf.LongHeaderJwtTokenKey = "long-jwt-token"
	}

	return &JwtxMiddlewareGinx{
		cache: cache,
		cfg:   *jwtConf,
	}
}

// deviceKey 返回用户-设备映射 key
func (j *JwtxMiddlewareGinx) deviceKey(userId int64, deviceID string) string {
	return fmt.Sprintf("user:device:%d:%s", userId, deviceID)
}

// devicesSetKey 返回用户所有设备集合 key
func (j *JwtxMiddlewareGinx) devicesSetKey(userId int64) string {
	return fmt.Sprintf("user:devices:%d", userId)
}

// sessionKey 返回 Redis session key
func (j *JwtxMiddlewareGinx) sessionKey(ssid, tokenType string) string {
	return fmt.Sprintf("user:session:%s:%s", ssid, tokenType)
}

// getDeviceID 从请求中提取或生成设备 ID
func (j *JwtxMiddlewareGinx) getDeviceID(ctx *gin.Context) string {
	// 优先使用前端传的 X-Device-ID
	if deviceID := ctx.GetHeader("X-Device-ID"); deviceID != "" {
		return deviceID
	}

	// 兜底：用 User-Agent 哈希（同一 UA 视为同一设备）
	ua := ctx.GetHeader("User-Agent")
	if ua == "" {
		ua = "unknown"
	}
	h := sha256.Sum256([]byte(ua))
	return fmt.Sprintf("%x", h)[:16] // 取前16字符作为设备ID
}

// SetToken 登录：自动按设备管理会话（同设备只保留最新）
func (j *JwtxMiddlewareGinx) SetToken(ctx *gin.Context, userId int64, name string, ssid string) (*UserClaims, error) {
	// 注意：ssid 参数已废弃，但为兼容接口保留（不使用）
	deviceID := j.getDeviceID(ctx)
	ssId := uuid.New().String()
	if ssid != "" {
		ssId = ssid
	}

	now := time.Now()
	userAgent := ctx.GetHeader("User-Agent")

	// === 踢掉该用户在此设备上的旧会话 ===
	oldSsidKey := j.deviceKey(userId, deviceID)
	oldSsid, err := j.cache.Get(ctx, oldSsidKey).Result()
	if err == nil && oldSsid != "" {
		// 删除旧 Token
		j.cache.Del(ctx,
			j.sessionKey(oldSsid, "access"),
			j.sessionKey(oldSsid, "refresh"),
		)
	}

	// --- 生成 Access Token ---
	accessClaims := &UserClaims{
		Uid:       userId,
		Name:      name,
		Ssid:      ssId,
		UserAgent: userAgent,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.cfg.DurationExpiresIn)),
		},
	}
	accessToken := jwt.NewWithClaims(j.cfg.SigningMethod, accessClaims)
	accessTokenStr, err := accessToken.SignedString(j.cfg.JwtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// --- 生成 Refresh Token ---
	refreshClaims := &UserClaims{
		Uid:       userId,
		Name:      name,
		Ssid:      ssId,
		UserAgent: userAgent,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.cfg.LongDurationExpiresIn)),
		},
	}
	refreshToken := jwt.NewWithClaims(j.cfg.SigningMethod, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(j.cfg.LongJwtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// --- 存入 Redis ---
	accessKey := j.sessionKey(ssId, "access")
	refreshKey := j.sessionKey(ssId, "refresh")

	err = j.cache.Set(ctx, accessKey, accessTokenStr, j.cfg.DurationExpiresIn).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store access token in redis: %w", err)
	}
	err = j.cache.Set(ctx, refreshKey, refreshTokenStr, j.cfg.LongDurationExpiresIn).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token in redis: %w", err)
	}

	// --- 更新设备会话映射 & 设备列表 ---
	j.cache.Set(ctx, oldSsidKey, ssId, j.cfg.LongDurationExpiresIn)
	j.cache.SAdd(ctx, j.devicesSetKey(userId), deviceID)
	j.cache.Expire(ctx, j.devicesSetKey(userId), 30*24*time.Hour) // 保留30天设备记录

	// --- 返回 Headers ---
	ctx.Header(j.cfg.HeaderJwtTokenKey, accessTokenStr)
	ctx.Header(j.cfg.LongHeaderJwtTokenKey, refreshTokenStr)
	ctx.Set("user", accessClaims)

	return accessClaims, nil
}

// ExtractToken 从 Authorization: Bearer <token> 提取
func (j *JwtxMiddlewareGinx) ExtractToken(ctx *gin.Context) string {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

// VerifyToken 验证 Access Token
func (j *JwtxMiddlewareGinx) VerifyToken(ctx *gin.Context) (*UserClaims, error) {
	tokenStr := j.ExtractToken(ctx)
	if tokenStr == "" {
		return nil, fmt.Errorf("missing access token")
	}

	claims := &UserClaims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return j.cfg.JwtKey, nil
	})

	if err != nil || !t.Valid || claims.TokenType != "access" {
		return claims, fmt.Errorf("invalid access token: %w", err)
	}

	key := j.sessionKey(claims.Ssid, "access")
	stored, err := j.cache.Get(ctx, key).Result()
	if err != nil || stored != tokenStr {
		return claims, fmt.Errorf("access token revoked or not found")
	}

	ctx.Set("user", claims)
	return claims, nil
}

// LongVerifyToken 验证 Refresh Token
func (j *JwtxMiddlewareGinx) LongVerifyToken(ctx *gin.Context) (*UserClaims, error) {
	tokenStr := j.ExtractToken(ctx)
	if tokenStr == "" {
		return nil, fmt.Errorf("missing refresh token")
	}

	claims := &UserClaims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return j.cfg.LongJwtKey, nil
	})

	if err != nil || !t.Valid || claims.TokenType != "refresh" {
		return claims, fmt.Errorf("invalid refresh token: %w", err)
	}

	key := j.sessionKey(claims.Ssid, "refresh")
	stored, err := j.cache.Get(ctx, key).Result()
	if err != nil || stored != tokenStr {
		return claims, fmt.Errorf("refresh token revoked")
	}

	ctx.Set("userLong", claims)
	return claims, nil
}

// RefreshToken 刷新 Token（可选新设备）
func (j *JwtxMiddlewareGinx) RefreshToken(ctx *gin.Context, newSsid string) (*UserClaims, error) {
	oldClaims, err := j.LongVerifyToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh failed: %w", err)
	}

	ssid := oldClaims.Ssid
	if newSsid != "" {
		ssid = newSsid
	}

	// 删除旧 Token（安全最佳实践）
	oldAccessKey := j.sessionKey(oldClaims.Ssid, "access")
	oldRefreshKey := j.sessionKey(oldClaims.Ssid, "refresh")
	j.cache.Del(ctx, oldAccessKey, oldRefreshKey)

	// 注意：这里复用 SetToken，会按新 ssid 创建会话（可能新设备）
	// 如果希望保持设备绑定，应传入原 deviceID，但当前设计以 ssid 为单位
	return j.SetToken(ctx, oldClaims.Uid, oldClaims.Name, ssid)
}

// DeleteToken 退出当前会话
func (j *JwtxMiddlewareGinx) DeleteToken(ctx *gin.Context) (*UserClaims, error) {
	claims, err := j.VerifyToken(ctx)
	if err != nil {
		return claims, fmt.Errorf("logout failed: %w", err)
	}

	accessKey := j.sessionKey(claims.Ssid, "access")
	refreshKey := j.sessionKey(claims.Ssid, "refresh")

	ctx.Header(j.cfg.HeaderJwtTokenKey, "")
	ctx.Header(j.cfg.LongHeaderJwtTokenKey, "")

	err1 := j.cache.Del(ctx, accessKey).Err()
	err2 := j.cache.Del(ctx, refreshKey).Err()

	if err1 != nil || err2 != nil {
		return claims, fmt.Errorf("failed to delete session tokens")
	}

	return claims, nil
}
