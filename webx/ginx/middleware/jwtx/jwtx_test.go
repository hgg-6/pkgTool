package jwtx

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestJwtxToken(t *testing.T) {
	//conf := JwtxMiddlewareGinxConfig{
	//	SigningMethod:         jwt.SigningMethodHS512,
	//	ExpiresIn:             time.Second * 60,
	//	LongExpiresIn:         time.Minute * 24,
	//	JwtKey:                []byte("123123123qwe"),
	//	LongJwtKey:            []byte("qweqwewqdsads21"),
	//	HeaderJwtTokenKey:     "",
	//	LongHeaderJwtTokenKey: "",
	//}
	j := NewJwtxMiddlewareGinx(&JwtxMiddlewareGinxConfig{
		SigningMethod:         jwt.SigningMethodHS512,
		ExpiresIn:             time.Second * 60,
		LongExpiresIn:         time.Second * 60 * 2,
		JwtKey:                []byte("123123123qwe"),
		LongJwtKey:            []byte("qweqwewqdsads21"),
		HeaderJwtTokenKey:     "duan-jwt-token",
		LongHeaderJwtTokenKey: "",
	})
	t.Log(j)

}
