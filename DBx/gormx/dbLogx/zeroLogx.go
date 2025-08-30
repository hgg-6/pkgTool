package dbLogx

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	glogger "gorm.io/gorm/logger"
	"time"
)

// GormLogStrx 适配器，将GORM日志转换为Zerolog
type GormLogStrx struct {
	// SlowThreshold
	//	- 慢查询阈值，单位为秒
	//	- 0 表示不启用慢查询
	SlowThreshold time.Duration
	logx          zerologx.Zlogger
}

// NewGormLogStrx 初始化GORM日志适配器
//   - 需先初始化日志模块，参考测试用例InitLog()方法
//   - gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: NewGormLogStrx(time.Second, InitLog())})
func NewGormLogStrx(slowThreshold time.Duration, logx zerologx.Zlogger) GormLogIn {
	return &GormLogStrx{SlowThreshold: slowThreshold, logx: logx}
}

// LogMode 实现gorm.Logger接口
func (l *GormLogStrx) LogMode(level glogger.LogLevel) glogger.Interface {
	// 使用Zerolog的级别控制，所以这里不需要做任何事
	return l
}

// Info 实现gorm.Logger接口 - 信息日志
func (l *GormLogStrx) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logx.Info().Msgf(msg, data...)
}

// Warn 实现gorm.Logger接口 - 警告日志
func (l *GormLogStrx) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logx.Warn().Msgf(msg, data...)
}

// Error 实现gorm.Logger接口 - 错误日志
func (l *GormLogStrx) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logx.Error().Msgf(msg, data...)
}

// Trace 实现gorm.Logger接口 - 跟踪日志（拆分成不同级别）
func (l *GormLogStrx) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 计算耗时，单位为毫秒
	elapsed := time.Since(begin)
	sql, rows := fc()

	// 如果有错误，记录错误日志
	if err != nil {
		l.logx.Error().
			Err(err).
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed-ms", elapsed).
			Msg("SQL Error")
		return
	}

	// 如果是慢查询，记录警告日志
	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.logx.Warn().
			Str("sql", sql).
			Int64("rows", rows).
			Dur("elapsed-ms", elapsed).
			Msg("Slow SQL")
		return
	}

	// 普通查询记录调试日志
	l.logx.Debug().
		Str("sql", sql).
		Int64("rows", rows).
		Dur("elapsed-ms", elapsed).
		Msg("SQL Query")
}
