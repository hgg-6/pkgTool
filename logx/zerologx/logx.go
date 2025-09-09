package zerologx

import (
	"github.com/rs/zerolog"
)

// Zlog 封装 zerolog.Logger
type ZeroLogStrx struct {
	logger *zerolog.Logger
}

func NewZeroLogx(l *zerolog.Logger) Zlogger {
	return &ZeroLogStrx{
		logger: l,
	}
}

func (z *ZeroLogStrx) Info() *zerolog.Event  { return z.logger.Info() }
func (z *ZeroLogStrx) Error() *zerolog.Event { return z.logger.Error() }
func (z *ZeroLogStrx) Debug() *zerolog.Event { return z.logger.Debug() }
func (z *ZeroLogStrx) Warn() *zerolog.Event  { return z.logger.Warn() }
func (z *ZeroLogStrx) With() zerolog.Context { return z.logger.With() }

func (z *ZeroLogStrx) GetZerolog() *zerolog.Logger {
	return z.logger
}

// eg:

//import (
//"gitee.com/hgg_test/pkg_tool/logx/zerologx"
//"github.com/rs/zerolog"
//"os"
//)

//// InitLog 初始化zerolog日志模块【wire里可直接 InitLog】
//func InitLog() zerologx.Zlogger {
//	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
//	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
//	// 模块化: Str("module", "userService模块")
//	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().Caller().Timestamp().Logger()
//	return zerologx.NewZlog(&logger)
//}
