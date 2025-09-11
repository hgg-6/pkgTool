package configx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	DbConfFile    = "db.yaml"
	RedisConfFile = "redis.yaml"
)

// TestInitViperLocals 测试读取多个配置文件
func TestInitViperLocals(t *testing.T) {

	// 初始化viper
	conf := NewViperConfigStr()
	// 初始化读取配置文件1
	err := conf.InitViperLocals("db", "yaml", ".")
	assert.NoError(t, err)
	// 初始化读取配置文件2
	err = conf.InitViperLocals("redis", "yaml", ".")
	assert.NoError(t, err)
	// 正常项目已经可在此返回ViperConfigIn接口了
	//return conf

	// 获取配置文件信息
	//	- 调用configx的单独Get方法，基于泛型约束，自动匹配返回值类型
	//	- conf参数为configx.ConfigIn接口，初始化配置文件时返回
	dbConf := Get[string](conf, "mysql.dsn", DbConfFile)
	redisConf := Get[string](conf, "redis.addr", RedisConfFile)
	t.Logf("dbConf: %s, redisConf: %s", dbConf, redisConf)
}

// TestInitViperLocalsWatchs 测试读取多个配置文件并监听文件变化
func TestInitViperLocalsWatchs(t *testing.T) {
	conf := NewViperConfigStr()
	err := conf.InitViperLocalsWatchs("db", "yaml", ".")
	assert.NoError(t, err)
	err = conf.InitViperLocalsWatchs("redis", "yaml", ".")
	assert.NoError(t, err)

	// 正常项目已经可在此返回ViperConfigIn接口了
	//return conf

	// 获取配置文件信息
	//	- 调用configx的单独Get方法，基于泛型约束，自动匹配返回值类型
	//	- conf参数为configx.ConfigIn接口，初始化配置文件时返回
	dbConf := Get[string](conf, "mysql.dsn", DbConfFile)
	redisConf := Get[string](conf, "redis.addr", RedisConfFile)
	t.Logf("dbConf: %s, redisConf: %s", dbConf, redisConf)

	time.Sleep(time.Minute * 5)
}
