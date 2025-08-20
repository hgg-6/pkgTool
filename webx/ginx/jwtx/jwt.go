package jwtx

//type UserClaimss struct {
//	jwt.RegisteredClaims // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
//	id                   int64
//	client               redis.Cmdable
//}
//
//// TokenCreate jwtKey 加密密钥
//
//func (u *UserClaimss) TokenCreate(jwtKey []byte) (string, error) {
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
//	return token.SignedString(jwtKey)
//}
//
//func (u *UserClaimss) TokenDelete(ctx *gin.Context) (string, error) {
//	ctx.Header("x-jwt-token", "")
//	ctx.Header("x-refresh-token", "")
//	uc := ctx.MustGet("user").(UserClaimss) // 获取用户信息，断言
//	fmt.Println("DEBUG: ", uc)
//	return u.client.Set(ctx, fmt.Sprintf("user:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
//}
