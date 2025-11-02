package scheduler

/*
	=================================
	此文件主要封装迁移调度器、校验逻辑
	=================================
*/

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/doubleWritePoolx"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/mysqlX/gormx/dbMovex/myMovex/validator"
	"gitee.com/hgg_test/pkg_tool/v2/channelx/mqX"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/webx/ginx"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
	"time"
)

// MigrationState 迁移状态
type MigrationState string

const (
	StateInitial   MigrationState = "initial"   // 初始状态
	StateSrcOnly   MigrationState = "src_only"  // 只写源库
	StateSrcFirst  MigrationState = "src_first" // 源库优先
	StateDstFirst  MigrationState = "dst_first" // 目标库优先
	StateDstOnly   MigrationState = "dst_only"  // 只写目标库
	StateCompleted MigrationState = "completed" // 迁移完成
)

// MigrationStats 迁移统计信息
type MigrationStats struct {
	StartTime          time.Time      `json:"start_time"`           // 迁移开始时间
	CurrentState       MigrationState `json:"current_state"`        // 当前状态
	FullValidationRuns int            `json:"full_validation_runs"` // 全量校验次数
	IncrValidationRuns int            `json:"incr_validation_runs"` // 增量校验次数
	DataDiscrepancies  int            `json:"data_discrepancies"`   // 数据不一致数量
	LastError          string         `json:"last_error"`           // 最后一次错误信息
}

// Scheduler 用来统一管理整个迁移过程
// 它不是必须的，你可以理解为这是为了方便用户操作而引入。
type Scheduler[T myMovex.Entity, Pdr any] struct {
	lock    sync.Mutex
	src     *gorm.DB
	dst     *gorm.DB
	pool    *doubleWritePoolx.DoubleWritePool // 双写池
	l       logx.Loggerx
	Pattern string         // 模式
	State   MigrationState // 迁移状态
	Stats   MigrationStats // 迁移统计信息

	cancelFull func() // 全量校验的取消函数
	cancelIncr func() // 增量校验的取消函数
	//producer          messageQueuex.ProducerIn[Pdr] // 消息队列生产者
	producer          mqX.Producer      // 消息队列生产者
	MessageQueueTopic string            // 消息队列主题Topic【默认为dbMove】
	fulls             map[string]func() // 全量校验的函数

	// 新增字段
	vdt    map[string]*validator.Validator[T, Pdr] // 活跃的校验器
	config SchedulerConfig                         // 调度器配置
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	DefaultSleepInterval time.Duration `json:"default_sleep_interval"` // 默认的睡眠间隔
	MaxValidationErrors  int           `json:"max_validation_errors"`  // 最大允许的校验错误数
	EnableAutoPromotion  bool          `json:"enable_auto_promotion"`  // 自动升级
	ValidationTimeout    time.Duration `json:"validation_timeout"`     // 校验超时
}

// NewScheduler
//   - pdr 消息队列生产者sync/async
func NewScheduler[T myMovex.Entity, Pdr any](l logx.Loggerx, src *gorm.DB, dst *gorm.DB,
	// 这个是业务用的 DoubleWritePool
	pool *doubleWritePoolx.DoubleWritePool,
	//producer messageQueuex.ProducerIn[Pdr]) *Scheduler[T, Pdr] {
	producer mqX.Producer) *Scheduler[T, Pdr] {
	return &Scheduler[T, Pdr]{
		l:       l,
		src:     src,
		dst:     dst,
		State:   StateInitial,
		Pattern: doubleWritePoolx.PatternSrcOnly,
		cancelFull: func() {
			// 初始的时候，啥也不用做
		},
		cancelIncr: func() {
			// 初始的时候，啥也不用做
		},
		pool:              pool,
		producer:          producer,
		MessageQueueTopic: "dbMove",
	}
}

// RegisterRoutes 这一个也不是必须的，就是你可以考虑利用配置中心，监听配置中心的变化
// 把全量校验，增量校验做成分布式任务，利用分布式任务调度平台来调度
func (s *Scheduler[T, Pdr]) RegisterRoutes(server *gin.RouterGroup) {
	// 模式切换
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))   // 源库只写
	server.POST("/src_first", ginx.Wrap(s.SrcFirst)) // 双写，源库优先
	server.POST("/dst_first", ginx.Wrap(s.DstFirst)) // 双写，目标库优先
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))   // 目标库只写

	// 校验控制
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))                            // 开启全量校验
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))                              // 停止全量校验
	server.POST("/incr/start", ginx.WrapBody[StartIncrRequest](s.StartIncrementValidation)) // 开启增量校验
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrementValidation))                         // 停止增量校验

	// 新增API
	server.GET("/status", ginx.Wrap(s.GetStatus))          // 获取状态
	server.POST("/auto-migrate", ginx.Wrap(s.AutoMigrate)) // 自动迁移
	server.GET("/health", ginx.Wrap(s.HealthCheck))        // 健康检查
	server.GET("/stats", ginx.Wrap(s.GetStats))            // 统计信息
}

// ---- 下面是四个阶段 ---- //

// SrcOnly 只读写源表
func (s *Scheduler[T, Pdr]) SrcOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Pattern = doubleWritePoolx.PatternSrcOnly
	if err := s.pool.UpdatePattern(doubleWritePoolx.PatternSrcOnly); err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "切换模式失败",
		}, err
	}
	s.Pattern = doubleWritePoolx.PatternSrcOnly
	s.State = StateSrcOnly
	s.Stats.CurrentState = StateSrcOnly
	s.l.Info("切换到源库只写模式")
	return ginx.Result{
		Msg: "已切换到源库只写模式",
	}, nil
}

// SrcFirst 双写，源库优先
func (s *Scheduler[T, Pdr]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Pattern = doubleWritePoolx.PatternSrcFirst
	if err := s.pool.UpdatePattern(doubleWritePoolx.PatternSrcFirst); err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "切换模式失败",
		}, err
	}
	s.Pattern = doubleWritePoolx.PatternSrcFirst
	s.State = StateSrcFirst
	s.Stats.CurrentState = StateSrcFirst
	s.l.Info("切换到双写，源库优先模式")
	return ginx.Result{
		Msg: "已切换到双写，源库优先模式",
	}, nil
}

// DstFirst 双写，目标库优先
func (s *Scheduler[T, Pdr]) DstFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Pattern = doubleWritePoolx.PatternDstFirst
	if err := s.pool.UpdatePattern(doubleWritePoolx.PatternDstFirst); err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "切换模式失败",
		}, err
	}
	s.Pattern = doubleWritePoolx.PatternDstFirst
	s.State = StateDstFirst
	s.Stats.CurrentState = StateDstFirst
	s.l.Info("切换到双写，目标库优先模式")
	return ginx.Result{
		Msg: "已切换到双写，目标库优先模式",
	}, nil
}

// DstOnly 只读写目标表
func (s *Scheduler[T, Pdr]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Pattern = doubleWritePoolx.PatternDstOnly
	if err := s.pool.UpdatePattern(doubleWritePoolx.PatternDstOnly); err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "切换模式失败",
		}, err
	}
	s.Pattern = doubleWritePoolx.PatternDstOnly
	s.State = StateDstOnly
	s.Stats.CurrentState = StateDstOnly
	s.l.Info("切换到目标库只写模式")
	return ginx.Result{
		Msg: "已切换到目标库只写模式",
	}, nil
}

// StopIncrementValidation 停止增量校验
func (s *Scheduler[T, Pdr]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// StartIncrementValidation 开启增量校验
func (s *Scheduler[T, Pdr]) StartIncrementValidation(c *gin.Context, req StartIncrRequest) (ginx.Result, error) {
	// 开启增量校验
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, nil
	}
	v.Incr().Utime(req.Utime).
		SleepInterval(time.Duration(req.Interval) * time.Millisecond)

	// 增加统计信息
	s.Stats.IncrValidationRuns++

	go func() {
		var ctx context.Context
		ctx, s.cancelIncr = context.WithCancel(context.Background())
		cancel()
		er := v.Validate(ctx)
		s.l.Warn("退出增量校验", logx.Error(er))
	}()
	return ginx.Result{
		Msg: "OK, 启动增量校验成功",
	}, nil
}

// StopFullValidation 停止全量校验
func (s *Scheduler[T, Pdr]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

// StartFullValidation 开始全量校验
func (s *Scheduler[T, Pdr]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	// 可以考虑去重的问题
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	// 增加统计信息
	s.Stats.FullValidationRuns++

	go func() {
		// 先取消上一次的
		cancel()
		er := v.Validate(ctx)
		if er != nil {
			s.l.Warn("退出全量校验", logx.Error(er))
		}
	}()
	return ginx.Result{
		Msg: "OK, 启动全量校验成功",
	}, nil
}

// newValidator 创建校验器
func (s *Scheduler[T, Pdr]) newValidator() (*validator.Validator[T, Pdr], error) {
	switch s.Pattern {
	case doubleWritePoolx.PatternSrcOnly, doubleWritePoolx.PatternSrcFirst:
		return validator.NewValidator[T, Pdr](s.src, s.dst, "SRC", s.l, &validator.MessageQueueStr[Pdr]{Producer: s.producer, MessageQueueTopic: s.MessageQueueTopic}), nil
	case doubleWritePoolx.PatternDstFirst, doubleWritePoolx.PatternDstOnly:
		//return validator.NewValidator[T, Pdr](s.dst, s.src, "DST", s.l, s.producer), nil
		return validator.NewValidator[T, Pdr](s.dst, s.src, "DST", s.l, &validator.MessageQueueStr[Pdr]{Producer: s.producer, MessageQueueTopic: s.MessageQueueTopic}), nil
	default:
		return nil, fmt.Errorf("未知的 Pattern %s", s.Pattern)
	}
}

func (s *Scheduler[T, Pdr]) GetStatus(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	health := s.pool.HealthCheck()
	metrics := s.pool.GetMetrics()

	status := map[string]interface{}{
		"current_state":   s.State,
		"current_pattern": s.Pattern,
		"migration_stats": s.Stats,
		"pool_health":     health,
		"pool_metrics":    metrics,
		"uptime":          time.Since(s.Stats.StartTime).String(),
	}

	return ginx.Result{
		Data: status,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T, Pdr]) HealthCheck(c *gin.Context) (ginx.Result, error) {
	health := s.pool.HealthCheck()

	if len(health) > 0 {
		return ginx.Result{
			Code: 5,
			Msg:  "健康检查失败",
			Data: health,
		}, nil
	}

	return ginx.Result{
		Msg: "服务健康",
	}, nil
}

func (s *Scheduler[T, Pdr]) GetStats(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return ginx.Result{
		Data: s.Stats,
		Msg:  "OK",
	}, nil
}

// AutoMigrate 自动迁移
func (s *Scheduler[T, Pdr]) AutoMigrate(c *gin.Context) (ginx.Result, error) {
	go s.executeMigrationPlan()

	return ginx.Result{
		Msg: "自动迁移流程已启动",
	}, nil
}
func (s *Scheduler[T, Pdr]) executeMigrationPlan() {
	s.l.Info("开始自动迁移流程")

	// 阶段1: 只写源库
	err := s.pool.UpdatePattern(doubleWritePoolx.PatternSrcOnly)
	if err != nil {
		s.l.Error("自动迁移失败, 更新双写模式失败", logx.Error(err))
	}
	s.State = StateSrcOnly

	// 阶段2: 双写，源库优先 + 全量校验
	time.Sleep(5 * time.Second) // 等待稳定
	err = s.pool.UpdatePattern(doubleWritePoolx.PatternSrcFirst)
	if err != nil {
		s.l.Error("自动迁移失败, 更新双写模式失败", logx.Error(err))
	}
	s.State = StateSrcFirst

	// 启动全量校验
	v, _ := s.newValidator()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err = v.Validate(ctx); err == nil {
		s.l.Info("全量校验通过，切换到目标库优先")

		// 阶段3: 双写，目标库优先
		err = s.pool.UpdatePattern(doubleWritePoolx.PatternDstFirst)
		if err != nil {
			s.l.Error("自动迁移失败, 更新双写模式失败", logx.Error(err))
		}
		s.State = StateDstFirst

		// 阶段4: 只写目标库
		time.Sleep(5 * time.Second)
		err = s.pool.UpdatePattern(doubleWritePoolx.PatternDstOnly)
		if err != nil {
			s.l.Error("自动迁移失败, 更新双写模式失败", logx.Error(err))
		}
		s.State = StateDstOnly

		s.l.Info("数据迁移完成")
	} else {
		s.l.Error("迁移失败，全量校验未通过", logx.Error(err))
	}
}

// 自动升级
func (s *Scheduler[T, Pdr]) autoPromoteIfReady() {
	if !s.config.EnableAutoPromotion {
		return
	}

	switch s.State {
	case StateSrcFirst:
		if s.Stats.DataDiscrepancies == 0 {
			s.l.Info("数据一致，自动切换到双写目标库优先模式")
			err := s.pool.UpdatePattern(doubleWritePoolx.PatternDstFirst)
			if err != nil {
				s.l.Error("自动切换到双写目标库优先模式, 更新双写模式失败", logx.Error(err))
			}
			s.State = StateDstFirst
			s.Stats.CurrentState = StateDstFirst
			s.Pattern = doubleWritePoolx.PatternDstFirst
		}
	case StateDstFirst:
		if s.Stats.DataDiscrepancies == 0 {
			s.l.Info("数据一致，自动切换到只写目标库模式")
			err := s.pool.UpdatePattern(doubleWritePoolx.PatternDstOnly)
			if err != nil {
				s.l.Error("自动切换到只写目标库模式, 更新双写模式失败", logx.Error(err))
			}
			s.State = StateDstOnly
			s.Stats.CurrentState = StateDstOnly
			s.Pattern = doubleWritePoolx.PatternDstOnly
		}
	}
}

// StartIncrRequest 增量校验请求
type StartIncrRequest struct {
	Utime    int64 `json:"utime"`    // 校验时间
	Interval int64 `json:"interval"` // 睡眠间隔
}

// SetMessageQueueTopic  设置消息队列主题Topic
func (s *Scheduler[T, Pdr]) SetMessageQueueTopic(Topic string) {
	s.MessageQueueTopic = Topic
}
