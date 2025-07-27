package main

import (
	"errors"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	l := InitLogServer()
	l.Info().Err(errors.New("test error")).Msg("hello")
}

// InitLog 初始化zerolog日志模块
func InitLog() Zlogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().Timestamp().Caller().Logger()
	return NewZlog(&logger)
}
