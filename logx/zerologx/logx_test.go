package zerologx

import (
	"errors"
	"github.com/rs/zerolog"
	"os"
	"testing"
)

func TestInitLog(t *testing.T) {
	// InitLog 初始化zerolog日志模块
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().Caller().Str("module", "userService模块").Timestamp().Logger()

	l := NewZeroLogx(&logger)
	l.Error().Err(errors.New("test error")).Int64("uid", 555).Msg("hello")

}
