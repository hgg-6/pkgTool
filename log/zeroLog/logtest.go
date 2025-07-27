package zeroLog

import (
	"github.com/rs/zerolog"
)

// Logger 接口（依赖抽象）
type Zlogger interface {
	Info() *zerolog.Event
	Error() *zerolog.Event
	Debug() *zerolog.Event
	Warn() *zerolog.Event
	With() zerolog.Context
}

// Zlog 封装 zerolog.Logger
type Zlog struct {
	logger *zerolog.Logger
}

func NewZlog(l *zerolog.Logger) Zlogger {
	return &Zlog{
		logger: l,
	}
}

func (z *Zlog) Info() *zerolog.Event  { return z.logger.Info() }
func (z *Zlog) Error() *zerolog.Event { return z.logger.Error() }
func (z *Zlog) Debug() *zerolog.Event { return z.logger.Debug() }
func (z *Zlog) Warn() *zerolog.Event  { return z.logger.Warn() }
func (z *Zlog) With() zerolog.Context { return z.logger.With() }

//// InitLog 初始化zerolog日志模块
//func InitLog() *zerolog.Logger {
//	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
//	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
//	// 模块化: Str("module", "userService模块")
//	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().Caller().Str("module", "userService模块").Timestamp().Logger()
//	return &logger
//}
