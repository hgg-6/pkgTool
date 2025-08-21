package jwtx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var JwtKey = []byte("rUwYX9LZXU0Vjiizkhzmj8VyBd3GcwrC") // JWTKey是加密的key，需要保密，不能泄露

func TestJwt(t *testing.T) {
	uid := int64(1)
	u := NewUserClaimss(uid, "my-app", time.Minute*24)
	token, err := u.TokenCreate(JwtKey)
	assert.NoError(t, err) // 断言，和require区别是：require会panic
	t.Log("token: ", token)
	err = u.TokenCheck(token, JwtKey)
	assert.NoError(t, err)
}
