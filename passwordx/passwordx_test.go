package passwordx

import "testing"

func TestPasswordx(t *testing.T) {
	p, err := JiaMi[string]("123123", 10)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(p))

	err = Check(string(p), "123123")
	if err != nil {
		t.Error(err)
	}
	t.Log("密码验证成功")
}
