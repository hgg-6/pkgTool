package jwtx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwtxToken(t *testing.T) {
	j := NewJwtxMiddlewareGinx(time.Minute*10, []byte("jfgkdfsjfjsarew12321edsa"))
	s, err := j.SetToken(nil, 1, "user", "112211221122")
	assert.NoError(t, err)
	t.Log(s)

}
