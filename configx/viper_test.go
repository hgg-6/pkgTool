package configx

import (
	"github.com/spf13/viper"
	"testing"
)

func TestViper(t *testing.T) {
	conf := viper.New()
	// 配置文件类型，初始化需添加
	conf.SetConfigType("yaml")
	v := NewViperConfigStr(conf)
	v.InitViperLocal("config/config.yaml", DefaultConfig{Key: "mysql.dsn", Val: "root:root@tcp(localhost:3306)/hgg"})
}
