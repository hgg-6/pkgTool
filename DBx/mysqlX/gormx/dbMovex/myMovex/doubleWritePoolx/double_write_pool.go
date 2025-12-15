package doubleWritePoolx

/*
	=================================
	此文件主要用来处理双写逻辑
	=================================
*/

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/atomicx"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	errUnknownPattern      = errors.New("未知的双写模式")
	errPrepareNotSupported = errors.New("双写模式不支持 Prepare 方法")
	errNilSourcePool       = errors.New("源库连接池为nil")
	errNilTargetPool       = errors.New("目标库连接池为nil")
)

// DoubleWriteConfig 双写配置
type DoubleWriteConfig struct {
	StrictMode    bool // true严格模式：任一失败就返回错误，默认false
	RetryAttempts int  // 重试次数
	EnableMetrics bool // 是否启用指标收集
}

// DoubleWritePool 双写连接池
type DoubleWritePool struct {
	Src     gorm.ConnPool
	Dst     gorm.ConnPool
	Pattern *atomicx.Value[string]
	L       logx.Loggerx
	Config  DoubleWriteConfig
	Metrics *Metrics
	mu      sync.RWMutex

	// 添加控制goroutine退出的字段
	ctx        context.Context
	cancel     context.CancelFunc
	metricsWg  sync.WaitGroup
}

// Metrics 监控指标
type Metrics struct {
	DoubleWriteSuccess int64
	DoubleWriteFailure int64
	QueryDuration      []time.Duration
}

// NewDoubleWritePool 创建双写连接池
func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB, l logx.Loggerx, config ...DoubleWriteConfig) *DoubleWritePool {
	cfg := DoubleWriteConfig{
		StrictMode:    false,
		RetryAttempts: 1,
		EnableMetrics: false,
	}
	if len(config) > 0 {
		cfg = config[0]
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &DoubleWritePool{
		Src:     src.ConnPool,
		Dst:     dst.ConnPool,
		L:       l,
		Pattern: atomicx.NewValueOf(PatternSrcOnly),
		Config:  cfg,
		Metrics: &Metrics{},
		ctx:     ctx,
		cancel:  cancel,
	}

	if cfg.EnableMetrics {
		pool.startMetricsCollection()
	}

	return pool
}

// Close 关闭双写池，释放资源
func (d *DoubleWritePool) Close() error {
	d.cancel()
	if d.Config.EnableMetrics {
		d.metricsWg.Wait()
	}
	d.L.Info("双写池已关闭")
	return nil
}

// startMetricsCollection 启动指标收集
func (d *DoubleWritePool) startMetricsCollection() {
	d.metricsWg.Add(1)
	go func() {
		defer d.metricsWg.Done()
		d.collectMetrics()
	}()
}

// UpdatePattern 更新双写模式
func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst, PatternDstOnly, PatternDstFirst:
		d.Pattern.Store(pattern)
		d.L.Info("双写模式已更新", logx.String("pattern", pattern))
		return nil
	default:
		return fmt.Errorf("%w: %s", errUnknownPattern, pattern)
	}
}

// HealthCheck 健康检查
func (d *DoubleWritePool) HealthCheck() map[string]error {
	health := make(map[string]error)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查源库连接（先检查是否为nil）
	if d.Src == nil {
		health["src"] = errNilSourcePool
	} else if src, ok := d.Src.(interface{ PingContext(context.Context) error }); ok {
		if err := src.PingContext(ctx); err != nil {
			health["src"] = err
		}
	}

	// 检查目标库连接（先检查是否为nil）
	if d.Dst == nil {
		health["dst"] = errNilTargetPool
	} else if dst, ok := d.Dst.(interface{ PingContext(context.Context) error }); ok {
		if err := dst.PingContext(ctx); err != nil {
			health["dst"] = err
		}
	}

	return health
}

// GetMetrics 获取监控指标
func (d *DoubleWritePool) GetMetrics() *Metrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Metrics
}

// BeginTx 开始事务
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.Pattern.Load()

	switch pattern {
	case PatternSrcOnly:
		if d.Src == nil {
			return nil, errNilSourcePool
		}
		src, err := d.Src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &DoubleWriteTx{
			src:     src,
			pattern: pattern,
			l:       d.L,
			config:  &d.Config,
		}, nil

	case PatternSrcFirst:
		if d.Src == nil {
			return nil, errNilSourcePool
		}
		src, err := d.Src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}

		var dst *sql.Tx
		if d.Dst != nil {
			dst, err = d.Dst.(gorm.TxBeginner).BeginTx(ctx, opts)
			if err != nil {
				d.L.Error("双写目标表开启事务失败", logx.Error(err))
				if d.Config.StrictMode {
					_ = src.Rollback()
					return nil, fmt.Errorf("strict mode: dst begin tx failed: %w", err)
				}
			}
		}
		return &DoubleWriteTx{
			src:     src,
			dst:     dst,
			pattern: pattern,
			l:       d.L,
			config:  &d.Config,
		}, nil

	case PatternDstFirst:
		if d.Dst == nil {
			return nil, errNilTargetPool
		}
		dst, err := d.Dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}

		var src *sql.Tx
		if d.Src != nil {
			src, err = d.Src.(gorm.TxBeginner).BeginTx(ctx, opts)
			if err != nil {
				d.L.Error("双写源表开启事务失败", logx.Error(err))
				if d.Config.StrictMode {
					_ = dst.Rollback()
					return nil, fmt.Errorf("strict mode: src begin tx failed: %w", err)
				}
			}
		}
		return &DoubleWriteTx{
			src:     src,
			dst:     dst,
			pattern: pattern,
			l:       d.L,
			config:  &d.Config,
		}, nil

	case PatternDstOnly:
		if d.Dst == nil {
			return nil, errNilTargetPool
		}
		dst, err := d.Dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &DoubleWriteTx{
			dst:     dst,
			pattern: pattern,
			l:       d.L,
			config:  &d.Config,
		}, nil

	default:
		return nil, errUnknownPattern
	}
}

// PrepareContext 不支持 Prepare 方法
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errPrepareNotSupported
}

// ExecContext 执行写操作
func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	var execErr error
	defer func() {
		duration := time.Since(start)
		d.recordMetrics(duration, execErr)
	}()

	pattern := d.Pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		if d.Src == nil {
			execErr = errNilSourcePool
			return nil, execErr
		}
		return d.execWithRetry(ctx, d.Src, query, args...)

	case PatternSrcFirst:
		if d.Src == nil {
			execErr = errNilSourcePool
			return nil, execErr
		}

		res, err := d.execWithRetry(ctx, d.Src, query, args...)
		execErr = err

		if err != nil {
			// 源库执行失败，如果严格模式则直接返回
			if d.Config.StrictMode {
				return res, err
			}
			// 非严格模式下，源库失败时不尝试写目标库，直接返回错误
			return res, err
		}

		// 源库执行成功，尝试写目标库
		if d.Dst != nil {
			_, err1 := d.execWithRetry(ctx, d.Dst, query, args...)
			if err1 != nil {
				d.L.Error("双写写入目标库失败",
					logx.Error(err1),
					logx.String("sql", query),
					logx.Any("args", args))
				if d.Config.StrictMode {
					execErr = fmt.Errorf("strict mode: dst exec failed: %w", err1)
					return res, execErr
				}
			}
		}
		return res, nil

	case PatternDstOnly:
		if d.Dst == nil {
			execErr = errNilTargetPool
			return nil, execErr
		}
		return d.execWithRetry(ctx, d.Dst, query, args...)

	case PatternDstFirst:
		if d.Dst == nil {
			execErr = errNilTargetPool
			return nil, execErr
		}

		res, err := d.execWithRetry(ctx, d.Dst, query, args...)
		execErr = err

		if err != nil {
			if d.Config.StrictMode {
				return res, err
			}
			return res, err
		}

		if d.Src != nil {
			_, err1 := d.execWithRetry(ctx, d.Src, query, args...)
			if err1 != nil {
				d.L.Error("双写写入源库失败",
					logx.Error(err1),
					logx.String("sql", query),
					logx.Any("args", args))
				if d.Config.StrictMode {
					execErr = fmt.Errorf("strict mode: src exec failed: %w", err1)
					return res, execErr
				}
			}
		}
		return res, nil

	default:
		execErr = errUnknownPattern
		return nil, execErr
	}
}

// QueryContext 执行查询操作
func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	var queryErr error
	defer func() {
		duration := time.Since(start)
		d.recordMetrics(duration, queryErr)
	}()

	pattern := d.Pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		if d.Src == nil {
			queryErr = errNilSourcePool
			return nil, queryErr
		}
		return d.Src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		if d.Dst == nil {
			queryErr = errNilTargetPool
			return nil, queryErr
		}
		return d.Dst.QueryContext(ctx, query, args...)
	default:
		queryErr = errUnknownPattern
		return nil, queryErr
	}
}

// QueryRowContext 执行查询单行操作
func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		d.recordMetrics(duration, nil)
	}()

	pattern := d.Pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		if d.Src == nil {
			// 返回一个包含错误的 Row
			return &sql.Row{}
		}
		return d.Src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		if d.Dst == nil {
			// 返回一个包含错误的 Row
			return &sql.Row{}
		}
		return d.Dst.QueryRowContext(ctx, query, args...)
	default:
		// 返回一个包含错误的 Row
		return &sql.Row{}
	}
}

// execWithRetry 带重试的执行
func (d *DoubleWritePool) execWithRetry(ctx context.Context, pool gorm.ConnPool, query string, args ...interface{}) (sql.Result, error) {
	var lastErr error
	for i := 0; i < d.Config.RetryAttempts; i++ {
		result, err := pool.ExecContext(ctx, query, args...)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if i < d.Config.RetryAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i+1) * 100 * time.Millisecond):
				// 指数退避
			}
		}
	}
	return nil, lastErr
}

// recordMetrics 记录指标
func (d *DoubleWritePool) recordMetrics(duration time.Duration, err error) {
	if !d.Config.EnableMetrics {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if err == nil {
		d.Metrics.DoubleWriteSuccess++
	} else {
		d.Metrics.DoubleWriteFailure++
	}

	if len(d.Metrics.QueryDuration) < 1000 { // 限制记录数量
		d.Metrics.QueryDuration = append(d.Metrics.QueryDuration, duration)
	}
}

// collectMetrics 定期收集和清理指标
func (d *DoubleWritePool) collectMetrics() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			// 保留最近100个指标
			if len(d.Metrics.QueryDuration) > 100 {
				d.Metrics.QueryDuration = d.Metrics.QueryDuration[len(d.Metrics.QueryDuration)-100:]
			}
			d.mu.Unlock()
		case <-d.ctx.Done():
			d.L.Info("指标收集goroutine收到停止信号")
			return
		}
	}
}

type Tx interface {
	gorm.ConnPool
	Commit() error
	Rollback() error
}

// DoubleWriteTx 双写事务
type DoubleWriteTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
	l       logx.Loggerx
	config  *DoubleWriteConfig
}

// Commit 提交事务
func (d *DoubleWriteTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		if d.src == nil {
			return errNilSourcePool
		}
		return d.src.Commit()

	case PatternSrcFirst:
		if d.src == nil {
			return errNilSourcePool
		}

		err := d.src.Commit()
		if err != nil {
			if d.dst != nil {
				_ = d.dst.Rollback()
			}
			return fmt.Errorf("src commit failed: %w", err)
		}

		if d.dst != nil {
			if err := d.dst.Commit(); err != nil {
				d.l.Error("目标库提交事务失败", logx.Error(err))
				if d.config.StrictMode {
					return fmt.Errorf("strict mode: dst commit failed: %w", err)
				}
			}
		}
		return nil

	case PatternDstFirst:
		if d.dst == nil {
			return errNilTargetPool
		}

		err := d.dst.Commit()
		if err != nil {
			if d.src != nil {
				_ = d.src.Rollback()
			}
			return fmt.Errorf("dst commit failed: %w", err)
		}

		if d.src != nil {
			if err := d.src.Commit(); err != nil {
				d.l.Error("源库提交事务失败", logx.Error(err))
				if d.config.StrictMode {
					return fmt.Errorf("strict mode: src commit failed: %w", err)
				}
			}
		}
		return nil

	case PatternDstOnly:
		if d.dst == nil {
			return errNilTargetPool
		}
		return d.dst.Commit()

	default:
		return errUnknownPattern
	}
}

// Rollback 回滚事务
func (d *DoubleWriteTx) Rollback() error {
	var errs []error

	switch d.pattern {
	case PatternSrcOnly:
		if d.src == nil {
			return errNilSourcePool
		}
		return d.src.Rollback()

	case PatternSrcFirst:
		if d.src != nil {
			if err := d.src.Rollback(); err != nil {
				errs = append(errs, fmt.Errorf("src rollback: %w", err))
			}
		}
		if d.dst != nil {
			if err := d.dst.Rollback(); err != nil {
				errs = append(errs, fmt.Errorf("dst rollback: %w", err))
			}
		}

	case PatternDstFirst:
		if d.dst != nil {
			if err := d.dst.Rollback(); err != nil {
				errs = append(errs, fmt.Errorf("dst rollback: %w", err))
			}
		}
		if d.src != nil {
			if err := d.src.Rollback(); err != nil {
				errs = append(errs, fmt.Errorf("src rollback: %w", err))
			}
		}

	case PatternDstOnly:
		if d.dst == nil {
			return errNilTargetPool
		}
		return d.dst.Rollback()

	default:
		return errUnknownPattern
	}

	if len(errs) > 0 {
		return fmt.Errorf("rollback errors: %v", errs)
	}
	return nil
}

// PrepareContext 不支持 Prepare 方法
func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errPrepareNotSupported
}

// ExecContext 在事务中执行写操作
func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		if d.src == nil {
			return nil, errNilSourcePool
		}
		return d.src.ExecContext(ctx, query, args...)

	case PatternSrcFirst:
		if d.src == nil {
			return nil, errNilSourcePool
		}

		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}

		if d.dst != nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("事务中双写写入目标库失败",
					logx.Error(err1),
					logx.String("sql", query))
				if d.config.StrictMode {
					return res, fmt.Errorf("strict mode: dst exec in tx failed: %w", err1)
				}
			}
		}
		return res, err

	case PatternDstOnly:
		if d.dst == nil {
			return nil, errNilTargetPool
		}
		return d.dst.ExecContext(ctx, query, args...)

	case PatternDstFirst:
		if d.dst == nil {
			return nil, errNilTargetPool
		}

		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}

		if d.src != nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("事务中双写写入源库失败",
					logx.Error(err1),
					logx.String("sql", query))
				if d.config.StrictMode {
					return res, fmt.Errorf("strict mode: src exec in tx failed: %w", err1)
				}
			}
		}
		return res, err

	default:
		return nil, errUnknownPattern
	}
}

// QueryContext 在事务中执行查询
func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		if d.src == nil {
			return nil, errNilSourcePool
		}
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		if d.dst == nil {
			return nil, errNilTargetPool
		}
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryRowContext 在事务中执行单行查询
func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		if d.src == nil {
			return &sql.Row{}
		}
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		if d.dst == nil {
			return &sql.Row{}
		}
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		return &sql.Row{}
	}
}

// 双写模式常量
const (
	PatternSrcOnly  = "src_only"  // 只写源库
	PatternSrcFirst = "src_first" // 先写源库，再写目标库
	PatternDstFirst = "dst_first" // 先写目标库，再写源库
	PatternDstOnly  = "dst_only"  // 只写目标库
)
