package zerologx

import (
	"encoding"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/rs/zerolog"
	"time"
)

// Zlog 封装 zerolog.Logger
type ZeroLogger struct {
	logger *zerolog.Logger
}

//	func NewZeroLogger(l *zerolog.Logger) *ZeroLogger {
//		return &ZeroLogger{
//			logger: l,
//		}
//	}
func NewZeroLogger(l *zerolog.Logger) logx.Loggerx {
	return &ZeroLogger{
		logger: l,
	}
}

func (z *ZeroLogger) logEvent(level zerolog.Level, msg string, fields []logx.Field) {
	//event := z.logger.WithLevel(level)
	//for _, f := range fields {
	//	event = event.Any(f.Key, f.Value)
	//}
	//event.Msg(msg)
	event := z.logger.WithLevel(level)
	// 当日志级别为，警告war和错误err 级别时，调用堆栈【默认跳过2层】
	if level == 2 || level == 3 {
		event.Caller(2)
	}
	for _, f := range fields {
		if f.Key == "" {
			continue // 避免空 key
		}
		toIfType(f, event)
	}
	event.Msg(msg)
}

func (z *ZeroLogger) Debug(msg string, fields ...logx.Field) {
	z.logEvent(zerolog.DebugLevel, msg, fields)
}

func (z *ZeroLogger) Info(msg string, fields ...logx.Field) {
	z.logEvent(zerolog.InfoLevel, msg, fields)
}

func (z *ZeroLogger) Warn(msg string, fields ...logx.Field) {
	z.logEvent(zerolog.WarnLevel, msg, fields)
}

func (z *ZeroLogger) Error(msg string, fields ...logx.Field) {
	z.logEvent(zerolog.ErrorLevel, msg, fields)
}

func toIfType(f logx.Field, event *zerolog.Event) {
	switch v := f.Value.(type) {
	case string:
		event = event.Str(f.Key, v)
	case []string:
		event = event.Strs(f.Key, v)
	case int:
		event = event.Int(f.Key, v)
	case int8:
		event = event.Int8(f.Key, v)
	case int16:
		event = event.Int16(f.Key, v)
	case int32:
		event = event.Int32(f.Key, v)
	case int64:
		event = event.Int64(f.Key, v)
	case uint:
		event = event.Uint(f.Key, v)
	case uint8:
		event = event.Uint8(f.Key, v)
	case uint16:
		event = event.Uint16(f.Key, v)
	case uint32:
		event = event.Uint32(f.Key, v)
	case uint64:
		event = event.Uint64(f.Key, v)
	case float32:
		event = event.Float32(f.Key, v)
	case float64:
		event = event.Float64(f.Key, v)
	case bool:
		event = event.Bool(f.Key, v)
	case time.Time:
		event = event.Time(f.Key, v)
	case time.Duration:
		event = event.Dur(f.Key, v)
	case error:
		if v != nil {
			event = event.Str(f.Key, v.Error())
		} else {
			event = event.Interface(f.Key, nil)
		}
	case fmt.Stringer:
		if v != nil {
			event = event.Str(f.Key, v.String())
		} else {
			event = event.Interface(f.Key, nil)
		}
	case encoding.TextMarshaler:
		if v != nil {
			if data, err := v.MarshalText(); err == nil {
				event = event.Str(f.Key, string(data))
			} else {
				event = event.Interface(f.Key, v)
			}
		} else {
			event = event.Interface(f.Key, nil)
		}
	case nil:
		event = event.Interface(f.Key, nil)
	default:
		event = event.Interface(f.Key, v) // fallback to reflection
	}
}

// eg:

//import (
//"gitee.com/hgg_test/pkg_tool/logx/zerologx"
//"github.com/rs/zerolog"
//"os"
//)

//// InitLog 初始化zerolog日志模块【wire里可直接 InitLog】
//func InitLog() logx.Loggerx {
//	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
//	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
//	// 模块化: Str("module", "userService模块")
//	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
//	return zerologx.NewZeroLogger(&logger)
//}
