package hashx

import (
	"testing"
)

func TestPasswordx(t *testing.T) {
	p, err := PasswdJiaMi("123123", 10)
	if err != nil {
		t.Error(err)
	}
	t.Log(p)

	err = PasswdJieMi(string(p), "123123")
	if err != nil {
		t.Error(err)
	}
	t.Log("密码验证成功")
}
