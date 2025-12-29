package config

import (
	"fmt"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/configx/viperX"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/fsnotify/fsnotify"
)

// Config 配置结构体
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Mysql  MysqlConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Log    LogConfig    `mapstructure:"log"`
	Cron   CronConfig   `mapstructure:"cron"`
	Alert  AlertConfig  `mapstructure:"alert"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// MysqlConfig MySQL配置
type MysqlConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// CronConfig Cron配置
type CronConfig struct {
	DefaultTimeout  int           `mapstructure:"default_timeout"`
	DefaultMaxRetry int           `mapstructure:"default_max_retry"`
	RetryBackoff    time.Duration `mapstructure:"retry_backoff"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled    bool       `mapstructure:"enabled"`
	WebhookURL string     `mapstructure:"webhook_url"`
	SMTP       SMTPConfig `mapstructure:"smtp"`
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string   `mapstructure:"host"`
	Port     int      `mapstructure:"port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	From     string   `mapstructure:"from"`
	To       []string `mapstructure:"to"`
}

// LoadConfig 加载配置文件
// 如果filePath为空，则自动搜索默认路径
func LoadConfig(filePath string, l logx.Loggerx) (*Config, error) {
	// 创建viper配置实例
	viperConfig := viperX.NewViperConfigStr(l).(*viperX.ViperConfigStr)

	// 初始化配置
	var err error
	if filePath != "" {
		// 使用指定路径
		err = viperConfig.InitViperLocal(filePath)
	} else {
		// 自动搜索默认路径
		err = viperConfig.InitViperLocal("config.yaml")
		if err != nil {
			// 尝试当前目录
			err = viperConfig.InitViperLocal("./config.yaml")
		}
		if err != nil {
			// 尝试config目录
			err = viperConfig.InitViperLocal("config/config.yaml")
		}
		if err != nil {
			// 尝试../config目录
			err = viperConfig.InitViperLocal("../config/config.yaml")
		}
	}

	if err != nil {
		return nil, err
	}

	// 解析配置到结构体（使用configx接口）
	var cfg Config
	//err = viperConfig.GetUnmarshalKey("", &cfg, "")
	err = viperX.GetUnmarshalStruct(viperConfig, "", &cfg)
	if err != nil {
		return nil, err
	}

	// 设置默认值
	SetDefaults(&cfg)

	// 验证配置
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// WatchConfig 监听配置变更
func WatchConfig(filePath string, l logx.Loggerx, onChange func(*Config)) error {
	// 创建viper配置实例
	viperConfig := viperX.NewViperConfigStr(l).(*viperX.ViperConfigStr)

	// 初始化配置（带监听）
	err := viperConfig.InitViperLocalWatch(filePath)
	if err != nil {
		return err
	}

	// 获取viper实例（用于监听变更）
	viper := viperConfig.GetViper()

	// 监听配置变更
	viper.OnConfigChange(func(e fsnotify.Event) {
		l.Info("配置文件变更", logx.String("file", e.Name))

		// 重新加载配置（使用configx接口）
		var cfg Config
		//if err := viperConfig.GetUnmarshalKey("", &cfg, ""); err != nil {
		if err := viperX.GetUnmarshalStruct(viperConfig, "", &cfg); err != nil {
			l.Error("重新加载配置失败", logx.Error(err))
			return
		}

		// 设置默认值
		SetDefaults(&cfg)

		// 验证配置
		if err := ValidateConfig(cfg); err != nil {
			l.Error("配置验证失败", logx.Error(err))
			return
		}

		// 调用变更回调
		onChange(&cfg)
	})

	return nil
}

// SetDefaults 设置配置默认值
func SetDefaults(cfg *Config) {
	// Server默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}

	// MySQL默认值
	if cfg.Mysql.MaxIdleConns == 0 {
		cfg.Mysql.MaxIdleConns = 10
	}
	if cfg.Mysql.MaxOpenConns == 0 {
		cfg.Mysql.MaxOpenConns = 100
	}
	if cfg.Mysql.ConnMaxLifetime == 0 {
		cfg.Mysql.ConnMaxLifetime = 3600 * time.Second
	}

	// Redis默认值
	if cfg.Redis.DB == 0 {
		cfg.Redis.DB = 0
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 100
	}
	if cfg.Redis.MinIdleConns == 0 {
		cfg.Redis.MinIdleConns = 10
	}

	// Log默认值
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "stdout"
	}

	// Cron默认值
	if cfg.Cron.DefaultTimeout == 0 {
		cfg.Cron.DefaultTimeout = 30
	}
	if cfg.Cron.DefaultMaxRetry == 0 {
		cfg.Cron.DefaultMaxRetry = 3
	}
	if cfg.Cron.RetryBackoff == 0 {
		cfg.Cron.RetryBackoff = 1 * time.Second
	}
}

// ValidateConfig 验证配置有效性
func ValidateConfig(cfg Config) error {
	// 验证MySQL配置
	if cfg.Mysql.DSN == "" {
		return fmt.Errorf("mysql dsn is required")
	}

	// 验证Redis配置
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis addr is required")
	}

	// 验证告警配置
	if cfg.Alert.Enabled {
		if cfg.Alert.WebhookURL == "" && (cfg.Alert.SMTP.Host == "" || cfg.Alert.SMTP.Port == 0) {
			return fmt.Errorf("alert webhook_url or smtp config is required when alert is enabled")
		}
	}

	return nil
}
