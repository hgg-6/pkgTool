package jwtX2

import "github.com/golang-jwt/jwt/v5"

// UserClaims 用于 Access 和 Refresh Token
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64  `json:"uid"`
	Name      string `json:"name"`
	Ssid      string `json:"ssid"` // 会话唯一 ID
	UserAgent string `json:"user_agent"`
	TokenType string `json:"token_type"` // "access" 或 "refresh"
}
