package hashx

import "testing"

func TestPasswordx(t *testing.T) {
	p, err := JiaMi("123123", 10)
	if err != nil {
		t.Error(err)
	}
	t.Log(p)

	err = Check(string(p), "123123")
	if err != nil {
		t.Error(err)
	}
	t.Log("密码验证成功")
}
