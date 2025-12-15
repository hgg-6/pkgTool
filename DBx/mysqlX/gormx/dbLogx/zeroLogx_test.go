package dbLogx

import (
	"testing"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// User 模型
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255"`
	Email     string    `gorm:"uniqueIndex;size:255"`
	Age       int       `gorm:"default:18"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func InitLog() logx.Loggerx {
	//	// #########################控制台彩色打印#########################
	//	//output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	//	//output.FormatLevel = func(i interface{}) string {
	//	//	return fmt.Sprintf("\x1b[%dm%-5s\x1b[0m", levelColor(i.(string)), i)
	//	//}
	//	// #########################控制台彩色打印#########################
	//
	// 设置Zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// 设置Zerolog全局级别
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// 创建lumberjack日志适配器，输出到配置文件
	lumberjackLogger := &lumberjack.Logger{
		// 注意有没有权限
		//Filename:   "/var/log/user.log", // 指定日志文件路径
		Filename:   "D:/soft_az/docker/hggPkg/ELK/log/user.log", // 指定日志文件路径
		MaxSize:    500,                                         // 每个日志文件的最大大小，单位：MB
		MaxBackups: 3,                                           // 保留旧日志文件的最大个数
		MaxAge:     28,                                          // 保留旧日志文件的最大天数
		Compress:   true,                                        // 是否压缩旧的日志文件
	}

	//log.Logger = zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().Caller().Timestamp().Logger()
	logger := zerolog.New(lumberjackLogger).With().Timestamp().Logger()
	return zerologx.NewZeroLogger(&logger)
}
func TestLog(t *testing.T) {
	// #########################控制台彩色打印#########################
	//output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	//output.FormatLevel = func(i interface{}) string {
	//	return fmt.Sprintf("\x1b[%dm%-5s\x1b[0m", levelColor(i.(string)), i)
	//}
	// #########################控制台彩色打印#########################

	// 设置Zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// 设置Zerolog全局级别
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// 创建lumberjack日志适配器，输出到配置文件
	lumberjackLogger := &lumberjack.Logger{
		// 注意有没有权限
		//Filename:   "/var/log/user.log", // 指定日志文件路径
		Filename:   "D:/soft_az/docker/hggPkg/ELK/log/user.log", // 指定日志文件路径
		MaxSize:    50,                                          // 每个日志文件的最大大小，单位：MB
		MaxBackups: 3,                                           // 保留旧日志文件的最大个数
		MaxAge:     28,                                          // 保留旧日志文件的最大天数
		Compress:   true,                                        // 是否压缩旧的日志文件
	}

	//log.Logger = zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().Caller().Timestamp().Logger().Output(os.Stderr)
	log.Logger = zerolog.New(lumberjackLogger).Level(zerolog.DebugLevel).With().Caller().Timestamp().Logger()

	// 创建自定义日志适配器
	gormConf := NewGormLogStrx(time.Second, InitLog())

	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13306)/hgg"), &gorm.Config{Logger: gormConf})
	if err != nil {
		t.Skipf("无法连接数据库: %v", err)
		return
	}
	err = db.Where("id = ?", 1).First(&User{}).Error
}

// 为不同日志级别设置颜色
//func levelColor(level string) int {
//	switch level {
//	case "DEBUG":
//		return 36 // 青色
//	case "INFO":
//		return 32 // 绿色
//	case "WARN":
//		return 33 // 黄色
//	case "ERROR":
//		return 31 // 红色
//	case "FATAL":
//		return 35 // 紫色
//	case "PANIC":
//		return 35 // 紫色
//	default:
//		return 0 // 默认
//	}
//}
