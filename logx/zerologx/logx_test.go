package zerologx

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/rs/zerolog"
	"os"
	"testing"
	"time"
)

func TestNewZeroLogger(t *testing.T) {
	// InitLog 初始化zerolog日志模块
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()

	l := NewZeroLogger(&logger)
	t.Log(time.Now().UnixMilli())
	// 当日志级别为，警告war和错误err 级别时，调用堆栈
	l.Info("初始化zerolog日志模块", logx.Int64("id", 1), logx.String("name", "hgg"))
	l.Error("初始化zerolog日志模块", logx.Int64("id", 1), logx.String("name", "hgg"))
	t.Log(time.Now().UnixMilli())
}
