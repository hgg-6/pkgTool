package jwtx

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Jwttclient interface {
	TokenCreate(jwtKey []byte) (string, error)
	TokenCheck(token string, jwtKey []byte) error
}

type UserClaimss struct {
	jwt.RegisteredClaims        // jwt.RegisteredClaims是jwt的默认结构体，里面有字段：Issuer, Subject, Audience, ExpiresAt, NotBefore, ID
	id                   int64  // 用户id
	role                 string // 角色/权限/id/归属【不定参数可拼接str】
}

// NewUserClaimss 创建用户jwt
func NewUserClaimss(id int64, role string, duration time.Duration) Jwttclient {
	return &UserClaimss{
		id:   id,
		role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)), // 过期时间
		},
	}
}

// TokenCreate jwtKey 加密密钥
func (u *UserClaimss) TokenCreate(jwtKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	return token.SignedString(jwtKey)
}

// TokenCheck jwtKey 解密密钥
func (u *UserClaimss) TokenCheck(token string, jwtKey []byte) error {
	t, err := jwt.ParseWithClaims(token, u, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil || t == nil || !t.Valid {
		return err
	}
	return nil
}
