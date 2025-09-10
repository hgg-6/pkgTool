package configx

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// TestViperConfigs 测试读取多个配置文件
func TestViperConfigs(t *testing.T) {
	// 初始化viper
	conf := NewViperConfigStr()
	// 读取配置文件1
	err := conf.InitViperLocals("db", "yaml", ".")
	assert.NoError(t, err)
	// 读取配置文件2
	err = conf.InitViperLocals("redis", "yaml", ".")
	assert.NoError(t, err)

	// 正常项目已经可在此返回了
	//return conf

	// 获取配置文件1 的viper实例
	dbConf, err := conf.GetNamedViper("db.yaml")
	assert.NoError(t, err)
	// 获取配置文件2 的viper实例
	redisConf, err := conf.GetNamedViper("redis.yaml")
	assert.NoError(t, err)
	log.Println(dbConf.GetString("mysql.dsn"))
	log.Println(redisConf.GetString("redis.addr"))
}
