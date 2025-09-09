package configx

import "github.com/spf13/viper"

type ViperConfigIn interface {
	// GetViper 获取viper的实例
	GetViper() *viper.Viper
	// InitViperLocal 配置单个文件
	//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
	//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
	InitViperLocal(filePath string, defaultConfig ...DefaultConfig) error
	// InitViperLocals 配置多个文件
	//   - 读取多个配置文件,fileName是文件名 精确文件名不带后缀，filePath是文件路径 精确到文件夹名，
	//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
	InitViperLocals(fileName, filePath string, defaultConfig ...DefaultConfig) error
	// InitViperLocalWatch  配置本地文件并监听文件变化
	//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
	//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
	InitViperLocalWatch(filePath string, defaultConfig ...DefaultConfig) error
	// InitViperRemote 配置远程文件
	//   - provider 是远程配置的提供者，这里使用的是etcd3
	//   - endpoint 是远程配置的访问地址
	//   - path 是远程配置的存储路径
	InitViperRemote(provider, endpoint, path string) error
}
