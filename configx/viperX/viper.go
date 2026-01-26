package viperX

import (
	"context"
	"fmt"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/configx"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type ViperConfigStr struct {
	Config   *viper.Viper
	Configs  *syncX.Map[string, *viper.Viper] // 使用并发安全的Map，无需额外锁
	interval time.Duration                    // 远程配置中心监听文件变更的间隔时间,默认5秒

	l                 logx.Loggerx
	remoteWatchCtx    context.Context
	remoteWatchCancel context.CancelFunc
}

func NewViperConfigStr(l logx.Loggerx) configx.ConfigIn {
	return &ViperConfigStr{
		Config: viper.New(),
		//Configs:  make(map[string]*viper.Viper),
		Configs:  &syncX.Map[string, *viper.Viper]{},
		interval: time.Second * 5,
		l:        l,
	}
}

// GetViper 获取viper的实例
func (v *ViperConfigStr) GetViper() *viper.Viper {
	return v.Config
}

// GetNamedViper 获取指定名称的viper实例【用于配置多个配置文件时使用】
// - name是配置文件名称，如：dev.yaml
func (v *ViperConfigStr) GetNamedViper(name string) (*viper.Viper, error) {
	// 使用并发安全的Map，无需加锁
	if val, ok := v.Configs.Load(name); ok {
		return val, nil
	}
	return nil, fmt.Errorf("config %s not found", name)
}

// InitViperLocal 配置单个文件
//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
func (v *ViperConfigStr) InitViperLocal(filePath string, defaultConfig ...configx.DefaultConfig) error {
	v.Config.SetConfigFile(filePath)

	if len(defaultConfig) != 0 {
		for _, s := range defaultConfig {
			v.Config.SetDefault(s.Key, s.Val)
		}
	}

	err := v.Config.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// InitViperLocals 配置多个文件
//   - 读取多个配置文件,fileName是文件名 精确文件名不带后缀，fileType是文件得类型eg: yaml、json....，filePath是文件路径 精确到文件夹名，
//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
func (v *ViperConfigStr) InitViperLocals(fileName, fileType, filePath string, defaultConfig ...configx.DefaultConfig) error {
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

	//v.Configs[fileName+"."+fileType] = v.Config
	v.Configs.Store(fileName+"."+fileType, v.Config)
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
		return err
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
func (v *ViperConfigStr) InitViperLocalWatch(filePath string, defaultConfig ...configx.DefaultConfig) error {
	v.Config.SetConfigFile(filePath)

	if len(defaultConfig) != 0 {
		for _, s := range defaultConfig {
			v.Config.SetDefault(s.Key, s.Val)
		}
	}

	// 开始监听配置文件变更
	v.Config.WatchConfig()
	// 配置文件变更时，执行回调函数【可自定义变更逻辑】
	v.Config.OnConfigChange(func(in fsnotify.Event) {
		//log.Println("本地配置文件发生变更: ", in.Name, in.Op)
		v.l.Warn("本地配置文件发生变更: ", logx.String("fileName", in.Name), logx.String("op", in.Op.String()))
	})

	err := v.Config.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

// InitViperLocalsWatchs 配置多个本地文件并监听文件变化
//   - filePath是文件路径 精确到文件名，如：config/dev.yaml
//   - defaultConfig是默认配置项【viper.SetDefault("mysql.dsn", "root:root@tcp(localhost:3306)/webook")】
func (v *ViperConfigStr) InitViperLocalsWatchs(fileName, fileType, filePath string, defaultConfig ...configx.DefaultConfig) error {
	v.Config = viper.New()
	v.Config.SetConfigName(fileName) // 配置文件名称(无扩展名)
	v.Config.SetConfigType(fileType) // 配置文件类型
	v.Config.AddConfigPath(filePath) // 添加配置文件路径，当前目录的config下【可以反复读多次，可以设置多个】

	if len(defaultConfig) != 0 {
		for _, s := range defaultConfig {
			v.Config.SetDefault(s.Key, s.Val)
		}
	}

	// 开始监听配置文件变更
	v.Config.WatchConfig()
	// 配置文件变更时，执行回调函数【可自定义变更逻辑】
	v.Config.OnConfigChange(func(in fsnotify.Event) {
		//log.Println("本地配置文件发生变更: ", in.Name, in.Op)
		v.l.Warn("本地配置文件发生变更: ", logx.String("fileName", in.Name), logx.String("op", in.Op.String()))
	})

	err := v.Config.ReadInConfig() // 读取配置文件
	if err != nil {
		return err
	}

	//v.Configs[fileName+"."+fileType] = v.Config
	v.Configs.Store(fileName+"."+fileType, v.Config)
	return nil
}

// InitViperRemoteWatch 配置远程文件并监听文件变化
//   - provider 是远程配置的提供者，这里使用的是etcd3
//   - endpoint 是远程配置的访问地址
//   - path 是远程配置的存储路径
//   - interval 是远程配置的监听间隔频率【几秒监听一次...】
func (v *ViperConfigStr) InitViperRemoteWatch(provider, endpoint, path string) error {
	// AddRemoteProvider参数 provider 是远程配置的提供者，这里使用的是etcd, endpoint 是远程配置的访问地址，path 是远程配置的存储路径
	err := v.Config.AddRemoteProvider(provider, endpoint, path)
	if err != nil {
		return err
	}
	err = v.Config.ReadRemoteConfig() // 读取远程配置文件
	if err != nil {
		return err
	}

	// 如果已有监听goroutine，先停止它
	if v.remoteWatchCancel != nil {
		v.remoteWatchCancel()
	}
	v.remoteWatchCtx, v.remoteWatchCancel = context.WithCancel(context.Background())

	// 启动监听远程配置变更
	go func() {
		ticker := time.NewTicker(v.interval)
		defer ticker.Stop()
		for {
			select {
			case <-v.remoteWatchCtx.Done():
				return
			case <-ticker.C:
				// 监听远程配置变更，如果有变更，则自动重新加载配置
				er := v.Config.WatchRemoteConfig()
				if er != nil {
					fmt.Printf("WatchRemoteConfig error: %v\n", er)
					continue
				}
				fmt.Println("远程配置中心监听变更，成功拉取并加载数据")
			}
		}
	}()

	return nil
}

// SetInterval 设置远程配置的监听间隔频率【几秒监听一次...】
//   - t 是远程配置的监听间隔频率【几秒监听一次...】
func (v *ViperConfigStr) SetInterval(t time.Duration) {
	v.interval = t
}

// Get 获取配置项【当整个项目读取/Init一个配置文件，fileName文件名留空，但整个项目读取/Init多个配置文件,需传入文件名eg: db.yaml】
//   - 新版本从configx.Get()单独读取配置文件
//   - 注意=============注意=============注意=============
func (v *ViperConfigStr) Get(key string, fileName ...string) any {
	//if len(fileName) == 0 || len(v.Configs) == 0 {
	if len(fileName) == 0 || v.Configs.IsEmpty() {
		return v.Config.Get(key)
	}
	//return v.Configs[fileName[0]].Get(key)
	val, ok := v.Configs.Load(fileName[0])
	if !ok {
		return nil
	}
	return val.Get(key)
}

// GetUnmarshalKey 获取配置项【当整个项目读取Init一个配置文件，fileName文件名留空，但整个项目读取Init多个配置文件,需传入文件名eg: db.yaml】
//   - 新版本从configx.GetUnmarshalStruct()单独读取配置文件
//   - 注意=============注意=============注意=============
func (v *ViperConfigStr) GetUnmarshalKey(key string, rawVal any, fileName ...string) error {
	//if len(fileName) == 0 || len(v.Configs) == 0 {
	if len(fileName) == 0 || v.Configs.IsEmpty() {
		return v.Config.UnmarshalKey(key, &rawVal)
	}
	//return v.Configs[fileName[0]].UnmarshalKey(key, &rawVal)
	val, ok := v.Configs.Load(fileName[0])
	if !ok {
		return fmt.Errorf("config file %s not found", fileName[0])
	}
	return val.UnmarshalKey(key, &rawVal)
}

// StopRemoteWatch 停止远程配置监控
func (v *ViperConfigStr) StopRemoteWatch() {
	if v.remoteWatchCancel != nil {
		v.remoteWatchCancel()
		v.remoteWatchCancel = nil
		v.remoteWatchCtx = nil
	}
}
