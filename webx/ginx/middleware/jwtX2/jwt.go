// Package jwtX2 提供基于 JWT 的认证中间件，支持多设备管理和会话控制。
package jwtX2

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

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
	cfg   *JwtxMiddlewareGinxConfig
}

func NewJwtxMiddlewareGinx(cache redis.Cmdable, jwtConf *JwtxMiddlewareGinxConfig) (JwtHandlerx, error) {
	if jwtConf == nil {
		return nil, errors.New("jwtConf must not be nil")
	}
	if len(jwtConf.JwtKey) == 0 || len(jwtConf.LongJwtKey) == 0 {
		return nil, errors.New("JwtKey and LongJwtKey must not be empty")
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
		cfg:   jwtConf,
	}, nil
}

// MustNewJwtxMiddlewareGinx 创建 JWT 中间件，配置错误时 panic
// Deprecated: 推荐使用 NewJwtxMiddlewareGinx 并处理返回的 error
func MustNewJwtxMiddlewareGinx(cache redis.Cmdable, jwtConf *JwtxMiddlewareGinxConfig) JwtHandlerx {
	h, err := NewJwtxMiddlewareGinx(cache, jwtConf)
	if err != nil {
		panic(err)
	}
	return h
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
	deviceID := fmt.Sprintf("%x", h)[:16] // 取前16字符作为设备ID
	// 确保 deviceID 非空（理论上不会发生，但安全起见）
	if deviceID == "" {
		deviceID = uuid.New().String()
	}
	return deviceID
}

// SetToken 登录：自动按设备管理会话（同设备只保留最新）
func (j *JwtxMiddlewareGinx) SetToken(ctx *gin.Context, userId int64, name string, ssid string) (*UserClaims, error) {
	// 注意：ssid 参数已废弃，但为兼容接口保留（不使用）
	deviceID := j.getDeviceID(ctx)
	if deviceID == "" {
		return nil, fmt.Errorf("device ID cannot be empty")
	}
	ssId := uuid.New().String()
	if ssid != "" {
		ssId = ssid
	}
	if ssId == "" {
		ssId = uuid.New().String()
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
		DeviceID:  deviceID, // P0-14: 记录会话绑定设备，供 RefreshToken/DeleteToken 精确清理
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
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.cfg.LongDurationExpiresIn)),
		},
	}
	refreshToken := jwt.NewWithClaims(j.cfg.SigningMethod, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(j.cfg.LongJwtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// --- 存入 Redis (使用Pipeline保证原子性) ---
	accessKey := j.sessionKey(ssId, "access")
	refreshKey := j.sessionKey(ssId, "refresh")

	pipe := j.cache.Pipeline()
	pipe.Set(ctx, accessKey, accessTokenStr, j.cfg.DurationExpiresIn)
	pipe.Set(ctx, refreshKey, refreshTokenStr, j.cfg.LongDurationExpiresIn)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to store tokens in redis: %w", err)
	}

	// --- 更新设备会话映射 & 设备列表 ---
	// P0-14 修复：devicesSetKey 的 TTL 与 deviceKey 一致（均用 LongDurationExpiresIn），
	// 旧实现硬编码 30 天，与 deviceKey 的 LongDurationExpiresIn（默认 7 天）不一致，
	// 导致 deviceKey 过期后 devicesSet 里残留幽灵设备最长 29 天。
	devicePipe := j.cache.Pipeline()
	devicePipe.Set(ctx, oldSsidKey, ssId, j.cfg.LongDurationExpiresIn)
	devicePipe.SAdd(ctx, j.devicesSetKey(userId), deviceID)
	devicePipe.Expire(ctx, j.devicesSetKey(userId), j.cfg.LongDurationExpiresIn)
	_, _ = devicePipe.Exec(ctx)

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

// revokeSessionScript 原子地撤销一个 ssid 对应的会话及其设备映射。
// P0-14 修复：用 Lua CAS 保证 refresh 并发安全 —— 只有旧 refresh token 仍是当前
// session 值时才执行删除（Compare-And-Delete），避免两个并发 refresh 请求互相覆盖。
// KEYS[1] = refresh session key, KEYS[2] = access session key,
// KEYS[3] = deviceKey(uid:deviceID), KEYS[4] = devicesSetKey(uid)
// ARGV[1] = 期望的 refresh token 值, ARGV[2] = deviceID
// 返回 1 表示已撤销，0 表示已被其他请求抢先撤销（调用方应中止）。
const revokeSessionScript = `
local cur = redis.call('GET', KEYS[1])
if cur == false or cur ~= ARGV[1] then
  return 0
end
redis.call('DEL', KEYS[1], KEYS[2], KEYS[3])
redis.call('SREM', KEYS[4], ARGV[2])
return 1
`

// RefreshToken 刷新 Token（可选新 ssid）
func (j *JwtxMiddlewareGinx) RefreshToken(ctx *gin.Context, newSsid string) (*UserClaims, error) {
	oldClaims, err := j.LongVerifyToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh failed: %w", err)
	}

	ssid := oldClaims.Ssid
	if newSsid != "" {
		ssid = newSsid
	}

	// P0-14 修复：旧实现用'当前请求'的 deviceID 删 deviceKey，但旧会话可能来自别的设备，
	// 导致删除无意义、旧设备 deviceKey 残留形成幽灵映射；并发 refresh 是 TOCTOU。
	// 现在用 oldClaims.DeviceID（旧会话绑定的真实设备）精确清理，并用 Lua CAS 保证
	// 只有旧 refresh token 仍是当前 session 值时才撤销，消除并发竞态。
	deviceID := oldClaims.DeviceID
	if deviceID == "" {
		// 兼容旧 token（无 DeviceID 字段）：退回到当前请求 deviceID，至少尽力清理。
		deviceID = j.getDeviceID(ctx)
	}
	oldRefreshToken := j.ExtractToken(ctx)

	keys := []string{
		j.sessionKey(oldClaims.Ssid, "refresh"),
		j.sessionKey(oldClaims.Ssid, "access"),
		j.deviceKey(oldClaims.Uid, deviceID),
		j.devicesSetKey(oldClaims.Uid),
	}
	revoked, err := j.cache.Eval(ctx, revokeSessionScript, keys, oldRefreshToken, deviceID).Int()
	if err != nil {
		return nil, fmt.Errorf("refresh: revoke old session failed: %w", err)
	}
	if revoked == 0 {
		// 旧 session 已被另一个并发 refresh 撤销并替换，本次 refresh 应中止，
		// 否则会用同一个 oldClaims 再签发一份，造成重复 session。
		return nil, fmt.Errorf("refresh: session already refreshed by a concurrent request")
	}

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

	// P0-14 修复：优先用 claims 里记录的会话绑定设备精确清理 deviceKey。
	// 旧实现用'当前请求'的 deviceID，与原登录设备可能不一致（如换 UA 调用登出），
	// 会导致删错设备或残留原设备映射。
	deviceID := claims.DeviceID
	if deviceID == "" {
		// 兼容旧 token（无 DeviceID 字段）。
		deviceID = j.getDeviceID(ctx)
	}
	deviceKey := j.deviceKey(claims.Uid, deviceID)

	err1 := j.cache.Del(ctx, accessKey).Err()
	err2 := j.cache.Del(ctx, refreshKey).Err()
	err3 := j.cache.Del(ctx, deviceKey).Err()

	if err1 != nil || err2 != nil {
		return claims, fmt.Errorf("failed to delete session tokens")
	}

	// 从设备集合中移除当前设备
	if err3 == nil {
		j.cache.SRem(ctx, j.devicesSetKey(claims.Uid), deviceID)
	}

	return claims, nil
}
