package configx

import (
	"github.com/spf13/viper"
)

type ViperConfigIn interface {
	// GetViper 获取viper的实例【仅用于整个项目单个配置文件】
	GetViper() *viper.Viper
	// GetNamedViper 获取指定名称的viper实例【用于配置多个配置文件时使用】
	GetNamedViper(name string) (*viper.Viper, error)
	// InitViperLocal 配置单个文件
	//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
	//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
	InitViperLocal(filePath string, defaultConfig ...DefaultConfig) error
	// InitViperLocals 配置多个文件
	//   - 读取多个配置文件,fileName是文件名 精确文件名不带后缀，fileType是文件得类型eg: yaml、json....，filePath是文件路径 精确到文件夹名，
	//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
	InitViperLocals(fileName, fileType, filePath string, defaultConfig ...DefaultConfig) error
	// InitViperLocalWatch  配置单个本地文件并监听文件变化
	//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
	//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
	InitViperLocalWatch(filePath string, defaultConfig ...DefaultConfig) error
	// InitViperRemote 配置远程文件
	//   - provider 是远程配置的提供者，这里使用的是etcd3
	//   - endpoint 是远程配置的访问地址
	//   - path 是远程配置的存储路径
	InitViperRemote(provider, endpoint, path string) error
	// InitViperRemoteWatch 配置远程文件并监听文件变化
	//   - provider 是远程配置的提供者，这里使用的是etcd3
	//   - endpoint 是远程配置的访问地址
	//   - path 是远程配置的存储路径
	//   - interval 是远程配置的监听间隔频率【几秒监听一次...】
	InitViperRemoteWatch(provider, endpoint, path string) error
}

//func InitConfigViper() configx.ViperConfigIn {
//	conf := configx.NewViperConfigStr(viper.New())
//	err := conf.InitViperLocalWatch("./config/dev.yaml",
//		// 默认配置，当配置文件读取失败时使用
//		configx.DefaultConfig{
//			Key: "mysql.dsn",
//			Val: "root:root@tcp(localhost:3306)/hgg",
//		},
//		// 默认配置，当配置文件读取失败时使用
//		configx.DefaultConfig{
//			Key: "redis.addr",
//			Val: "localhost:6379",
//		},
//	)
//
//	if err != nil {
//		panic(err)
//	}
//	return conf
//}

/*
	如果配置多个配置文件
	以InitViperLocal()为例
	func InitViperLocal() configx.ViperConfigIn {
	conf := configx.NewViperConfigStr(viper.New())
	err := conf.InitViperLocalWatch("./config/dev.yaml",
		// 默认配置，当配置文件读取失败时使用
		configx.DefaultConfig{
			Key: "mysql.dsn",
			Val: "root:root@tcp(localhost:3306)/hgg",
		},
		// 默认配置，当配置文件读取失败时使用
		configx.DefaultConfig{
			Key: "redis.addr",
			Val: "localhost:6379",
		},
	)

	if err != nil {
		panic(err)
	}
	return conf
}
*/
