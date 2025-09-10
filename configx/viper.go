package configx

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"sync"
	"time"
)

type ViperConfigStr struct {
	Config   *viper.Viper
	Configs  map[string]*viper.Viper
	mutex    sync.RWMutex
	interval time.Duration // 远程配置中心监听文件变更的间隔时间
}

func NewViperConfigStr() ViperConfigIn {
	return &ViperConfigStr{
		Config:   viper.New(),
		Configs:  make(map[string]*viper.Viper),
		interval: time.Second * 5,
	}
}

// GetViper 获取viper的实例
func (v *ViperConfigStr) GetViper() *viper.Viper {
	return v.Config
}

// GetNamedViper 获取指定名称的viper实例【用于配置多个配置文件时使用】
// - name是配置文件名称，如：dev.yaml
func (v *ViperConfigStr) GetNamedViper(name string) (*viper.Viper, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if v, ok := v.Configs[name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("config %s not found", name)
}

// InitViperLocal 配置单个文件
//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
func (v *ViperConfigStr) InitViperLocal(filePath string, defaultConfig ...DefaultConfig) error {
	//cfilg := pflag.String("config", "config/dev.yaml", "配置文件路径") // pflag.String是设置命令行参数，用于指定配置文件路径
	cfilg := pflag.String("config", filePath, "配置文件路径") // pflag.String是设置命令行参数，用于指定配置文件路径
	pflag.Parse()                                             // 解析命令行参数，pflag.String时cfilg还没有值，需要调一下pflag.Parse()，cfilg才有值config/config.yaml

	v.Config.SetConfigFile(*cfilg) // 配置文件名称【pflag.String时cfilg指定配置文件路径】

	if len(defaultConfig) != 0 {
		for _, s := range defaultConfig {
			v.Config.SetDefault(s.Key, s.Val)
		}
	}

	err := v.Config.ReadInConfig() // 读取配置文件
	if err != nil {
		return err
	}
	return nil
}

// InitViperLocals 配置多个文件
//   - 读取多个配置文件,fileName是文件名 精确文件名不带后缀，fileType是文件得类型eg: yaml、json....，filePath是文件路径 精确到文件夹名，
//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
func (v *ViperConfigStr) InitViperLocals(fileName, fileType, filePath string, defaultConfig ...DefaultConfig) error {
	v.Config = viper.New()
	v.Config.SetConfigName(fileName) // 配置文件名称(无扩展名)
	v.Config.SetConfigType(fileType) // 配置文件类型
	v.Config.AddConfigPath(filePath) // 添加配置文件路径，当前目录的config下【可以反复读多次，可以设置多个】

	if len(defaultConfig) != 0 {
		for _, s := range defaultConfig {
			v.Config.SetDefault(s.Key, s.Val)
		}
	}

	err := v.Config.ReadInConfig() // 读取配置文件
	if err != nil {
		return err
	}

	v.Configs[fileName+"."+fileType] = v.Config
	return nil
}

// InitViperRemote 配置远程文件
//   - provider 是远程配置的提供者，这里使用的是etcd3
//   - endpoint 是远程配置的访问地址
//   - path 是远程配置的存储路径
func (v *ViperConfigStr) InitViperRemote(provider, endpoint, path string) error {
	// AddRemoteProvider参数 provider 是远程配置的提供者，这里使用的是etcd, endpoint 是远程配置的访问地址，path 是远程配置的存储路径
	err := v.Config.AddRemoteProvider(provider, endpoint, path)
	if err != nil {
		panic(err)
	}
	err = v.Config.ReadRemoteConfig() // 读取远程配置文件
	if err != nil {
		return err
	}
	return nil
}

// InitViperLocalWatch  配置本地文件并监听文件变化
//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
func (v *ViperConfigStr) InitViperLocalWatch(filePath string, defaultConfig ...DefaultConfig) error {
	//cfilg := pflag.String("config", "config/dev.yaml", "配置文件路径") // pflag.String是设置命令行参数，用于指定配置文件路径
	cfilg := pflag.String("config", filePath, "配置文件路径") // pflag.String是设置命令行参数，用于指定配置文件路径
	pflag.Parse()                                             // 解析命令行参数，pflag.String时cfilg还没有值，需要调一下pflag.Parse()，cfilg才有值config/config.yaml

	v.Config.SetConfigFile(*cfilg) // 配置文件名称【pflag.String时cfilg指定配置文件路径】

	if len(defaultConfig) != 0 {
		for _, s := range defaultConfig {
			v.Config.SetDefault(s.Key, s.Val)
		}
	}

	// 开始监听配置文件变更
	v.Config.WatchConfig()
	// 配置文件变更时，执行回调函数【可自定义变更逻辑】
	v.Config.OnConfigChange(func(in fsnotify.Event) {
		log.Println("本地配置文件发生变更: ", in.Name, in.Op)
	})

	err := v.Config.ReadInConfig() // 读取配置文件
	if err != nil {
		return err
	}
	return nil
}

func (v *ViperConfigStr) InitViperRemoteWatch(provider, endpoint, path string) error {
	// AddRemoteProvider参数 provider 是远程配置的提供者，这里使用的是etcd, endpoint 是远程配置的访问地址，path 是远程配置的存储路径
	err := v.Config.AddRemoteProvider(provider, endpoint, path)
	if err != nil {
		panic(err)
	}
	err = v.Config.ReadRemoteConfig() // 读取远程配置文件
	if err != nil {
		return err
	}

	// 启动监听远程配置变更
	go func() {
		for {
			// 控制检查频率，默认每5秒检查一次
			time.Sleep(v.interval)
			// 监听远程配置变更，如果有变更，则自动重新加载配置
			er := viper.WatchRemoteConfig()
			if er != nil {
				fmt.Printf("WatchRemoteConfig error: %v\n", er)
				continue
			}
			fmt.Println("远程配置中心监听变更，成功拉取并加载数据")
		}
	}()

	return nil
}

type DefaultConfig struct {
	Key string
	Val any
}
