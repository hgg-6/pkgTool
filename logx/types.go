package logx

// Loggerx  接口（依赖抽象）
//
//go:generate mockgen -source=./types.go -package=logxmocks -destination=mocks/logx.mock.go Loggerx
type Loggerx interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}
type Field struct {
	Key   string
	Value any
}
