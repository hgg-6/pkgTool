package logx

import (
	"time"
)

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Int8(key string, val int8) Field {
	return Field{Key: key, Value: val}
}

func Int16(key string, val int16) Field {
	return Field{Key: key, Value: val}
}
func Int32(key string, val int32) Field {
	return Field{Key: key, Value: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func Uint(key string, val uint) Field {
	return Field{Key: key, Value: val}
}

func Uint8(key string, val uint8) Field {
	return Field{Key: key, Value: val}
}

func Uint16(key string, val uint16) Field {
	return Field{Key: key, Value: val}
}
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Value: val}
}

func Uint64(key string, val uint64) Field {
	return Field{Key: key, Value: val}
}

func Float32(key string, val float32) Field {
	return Field{Key: key, Value: val}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}
func String(key string, val string) Field {
	return Field{Key: key, Value: val}
}

func TimeTime(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}

func TimeDuration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

func Any(key string, val any) Field {
	return Field{Key: key, Value: val}
}
