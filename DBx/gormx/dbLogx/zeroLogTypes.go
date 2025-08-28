package dbLogx

import (
	"context"
	glogger "gorm.io/gorm/logger"
	"time"
)

type GormLogIn interface {
	LogMode(level glogger.LogLevel) glogger.Interface
	Info(ctx context.Context, msg string, data ...interface{})
	Warn(ctx context.Context, msg string, data ...interface{})
	Error(ctx context.Context, msg string, data ...interface{})
	Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error)
}
