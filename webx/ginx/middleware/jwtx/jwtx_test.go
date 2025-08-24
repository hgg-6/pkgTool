package jwtx

import (
	"testing"
)

func TestJwtxToken(t *testing.T) {
	//j := NewJwtxMiddlewareGinx(time.Minute * 10)
	//s, err := j.SetToken(nil, 1, JWTKey, "User-Agent", "112211221122")
	//assert.NoError(t, err)
	//t.Log(s)
	maps := map[string]any{
		"users": map[string]any{
			"name1": "hgg1",
			"name2": "hgg2",
			"users2": map[string]any{
				"name3": "hgg3",
			},
		},
	}
	s, ok := MapKeyFunctional(maps, "users.*.name3", true)
	if ok {
		t.Log(s)
	}
}
