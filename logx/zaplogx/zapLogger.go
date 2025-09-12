package zaplogx

import (
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"go.uber.org/zap"
)

type ZapLogger struct {
	l *zap.Logger
}

//	func NewZapLogger(l *zap.Logger) *ZapLogger {
//		return &ZapLogger{
//			l: l,
//		}
//	}
func NewZapLogger(l *zap.Logger) logx.Loggerx {
	return &ZapLogger{
		l: l,
	}
}

func (z *ZapLogger) Debug(msg string, fields ...logx.Field) {
	z.l.Debug(msg, z.toArgs(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...logx.Field) {
	z.l.Info(msg, z.toArgs(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...logx.Field) {
	z.l.Warn(msg, z.toArgs(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...logx.Field) {
	z.l.Error(msg, z.toArgs(fields)...)
}

// 转换参数
func (z *ZapLogger) toArgs(args []logx.Field) []zap.Field { // 适配
	res := make([]zap.Field, 0, len(args)) // 创建一个空切片，预分配内存，创建一个切片[]zap.Field，第一个参数是切片的容量，第二个参数是切片的长度
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}
