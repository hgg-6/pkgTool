package jwtX2

import (
	"github.com/redis/go-redis/v9"
	"testing"
)

func TestNewJwtxMiddlewareGinx(t *testing.T) {
	NewJwtxMiddlewareGinx(InitCache(), &JwtxMiddlewareGinxConfig{
		SigningMethod:         nil,
		DurationExpiresIn:     0,
		LongDurationExpiresIn: 0,
		JwtKey:                []byte("qwerrewq"),
		LongJwtKey:            []byte("qwerrewq123"),
		HeaderJwtTokenKey:     "",
		LongHeaderJwtTokenKey: "",
	})
}

func InitCache() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
}
