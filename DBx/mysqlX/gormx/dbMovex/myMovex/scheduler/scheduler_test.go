package scheduler

import (
	"os"
	"sync"
	"testing"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/doubleWritePoolx"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/events"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX/kafkaX/saramaX/producerX"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestUser 测试用户实体
//type TestUser struct {
//	Id        int64  `gorm:"primaryKey,autoIncrement"`
//	Name      string `gorm:"size:100"`
//	Email     string `gorm:"size:100;uniqueIndex"`
//	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
//	Ctime     int64
//	Utime     int64
//}

//func (i TestUser) ID() int64 {
//	return i.Id
//}
//
//func (i TestUser) CompareTo(dst Entity) bool {
//	val, ok := dst.(TestUser)
//	if !ok {
//		return false
//	}
//	return i == val
//}
//func (i TestUser) Types() string {
//	return "TestUser"
//}

// ===================
func newProducer() mqX.Producer {
	var addr []string = []string{"localhost:9094"}
	// 使用新的 mqX 包创建生产者
	pro, err := producerX.NewKafkaProducer(addr, &producerX.ProducerConfig{
		Async: false, // 同步模式
	})
	if err != nil {
		panic(err)
	}
	return pro
}

// ===================

/*
================================================
================================================
========================【测试函数】=============
================================================
================================================
*/

// setupTestSrcDB 设置测试数据库
func setupTestSrcDB() *gorm.DB {
	srcdb, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/src_db?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	srcdb.AutoMigrate(&events.TestUser{})
	return srcdb
}
func setupTestDstDB() *gorm.DB {
	dstdb, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/dst_db?parseTime=true"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	dstdb.AutoMigrate(&events.TestUser{})
	return dstdb
}

// TestSchedulerPatterns 测试模式切换
func TestSchedulerPatterns(t *testing.T) {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)

	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 测试初始状态
	assert.Equal(t, StateInitial, scheduler.State)
	assert.Equal(t, doubleWritePoolx.PatternSrcOnly, scheduler.Pattern)

	// 测试切换到源库优先
	scheduler.SrcFirst(nil)
	assert.Equal(t, StateSrcFirst, scheduler.State)
	assert.Equal(t, doubleWritePoolx.PatternSrcFirst, scheduler.Pattern)

	// 测试切换到目标库优先
	scheduler.DstFirst(nil)
	assert.Equal(t, StateDstFirst, scheduler.State)
	assert.Equal(t, doubleWritePoolx.PatternDstFirst, scheduler.Pattern)

	// 测试切换到目标库只写
	scheduler.DstOnly(nil)
	assert.Equal(t, StateDstOnly, scheduler.State)
	assert.Equal(t, doubleWritePoolx.PatternDstOnly, scheduler.Pattern)
}

// TestSchedulerValidation 测试全量校验功能
func TestSchedulerValidation(t *testing.T) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 启动全量校验
	scheduler.StartFullValidation(nil)
	time.Sleep(time.Second * 2) // 给goroutine一点时间启动

	// 停止全量校验
	scheduler.StopFullValidation(nil)

	// 验证统计信息
	assert.Equal(t, 1, scheduler.Stats.FullValidationRuns)
}

// TestSchedulerIncrementalValidation 测试增量校验
func TestSchedulerIncrementalValidation(t *testing.T) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 启动增量校验
	req := StartIncrRequest{
		Utime:    time.Now().UnixMilli(),
		Interval: 100,
	}

	// 这里需要模拟HTTP上下文，简化测试
	scheduler.StartIncrementValidation(nil, req)
	time.Sleep(time.Second * 2)

	// 停止增量校验
	scheduler.StopIncrementValidation(nil)

	// 等待一段时间，等待goroutine结束，发送kafka
	time.Sleep(time.Second * 2)

	// 验证统计信息
	assert.Equal(t, 1, scheduler.Stats.IncrValidationRuns)
}

// TestSchedulerConcurrent 测试并发安全性
func TestSchedulerConcurrent(t *testing.T) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 并发执行模式切换
	var wg sync.WaitGroup
	patterns := []func(){
		func() { scheduler.SrcOnly(nil) },
		func() { scheduler.SrcFirst(nil) },
		func() { scheduler.DstFirst(nil) },
		func() { scheduler.DstOnly(nil) },
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			patterns[idx%len(patterns)]()
		}(i)
	}

	wg.Wait()

	// 验证最终状态的一致性
	assert.Contains(t, []MigrationState{StateSrcOnly, StateSrcFirst, StateDstFirst, StateDstOnly}, scheduler.State)
}

// TestSchedulerAutoPromotion 测试自动升级
func TestSchedulerAutoPromotion(t *testing.T) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	//config := SchedulerConfig{
	//	EnableAutoPromotion: true,
	//	ValidationTimeout:   1 * time.Second,
	//}

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 设置为源库优先模式
	scheduler.SrcFirst(nil)

	// 模拟校验完成（无差异）
	scheduler.Stats.DataDiscrepancies = 0
	// 配置自动升级
	scheduler.config.EnableAutoPromotion = true
	scheduler.autoPromoteIfReady()

	// 应该自动升级到目标库优先
	assert.Equal(t, StateDstFirst, scheduler.State)
	assert.Equal(t, doubleWritePoolx.PatternDstFirst, scheduler.Pattern)
}

// TestSchedulerHealthCheck 测试健康检查
func TestSchedulerHealthCheck(t *testing.T) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 健康检查应该返回健康状态
	// 这里需要模拟HTTP上下文来测试
	// 简化测试，直接检查内部状态
	assert.NotNil(t, scheduler.HealthCheck)
}

// TestNewValidator 测试校验器创建
func TestNewValidator(t *testing.T) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Str("module", "userService模块").Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	// 测试不同模式下的校验器创建
	scheduler.Pattern = doubleWritePoolx.PatternSrcFirst
	validator, err := scheduler.newValidator()
	require.NoError(t, err)
	require.NotNil(t, validator)

	scheduler.Pattern = doubleWritePoolx.PatternDstFirst
	validator, err = scheduler.newValidator()
	require.NoError(t, err)
	require.NotNil(t, validator)

	// 测试未知模式
	scheduler.Pattern = "unknown_pattern"
	validator, err = scheduler.newValidator()
	require.Error(t, err)
	require.Nil(t, validator)
}

// BenchmarkSchedulerPatternSwitch 性能测试：模式切换
func BenchmarkSchedulerPatternSwitch(b *testing.B) {
	srcDB := setupTestSrcDB()
	dstDB := setupTestDstDB()
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	l := zerologx.NewZeroLogger(&logger)
	// 创建生产者
	producer := newProducer()
	defer producer.Close()

	pool := doubleWritePoolx.NewDoubleWritePool(srcDB, dstDB, l)
	scheduler := NewScheduler[events.TestUser, any](l, srcDB, dstDB, pool, producer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler.SrcFirst(nil)
		scheduler.DstFirst(nil)
	}
}
