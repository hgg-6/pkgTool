package redsyncx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/go-redsync/redsync/v4"
	redRedis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockStopped 锁已停止错误
	ErrLockStopped = errors.New("锁已停止")
	// ErrLockNotAcquired 锁未获取错误
	ErrLockNotAcquired = errors.New("锁未获取")
)

// LockStatus 锁状态枚举
type LockStatus int

const (
	// LockStatusUnknown 锁未知状态
	LockStatusUnknown LockStatus = iota
	// LockStatusAcquired 锁已获取
	LockStatusAcquired
	// LockStatusLost 锁丢失
	LockStatusLost
	// LockStatusReleased 锁释放
	LockStatusReleased
)

// LockResult 锁操作结果
type LockResult struct {
	Status LockStatus
	Error  error
}

// LockRedsync 分布式锁重构版
type LockRedsync struct {
	rsMutex  *redsync.Mutex
	logger   logx.Loggerx
	lockName string
	config   Config

	// 状态管理
	mu           sync.RWMutex
	status       LockStatus
	isRunning    bool
	acquiredTime time.Time

	// 控制goroutine
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	statusChan chan LockResult

	// 续约控制
	renewalTicker *time.Ticker
}

// Config 锁配置
type Config struct {
	// 锁名称
	LockName string
	// 锁过期时间
	Expiry time.Duration
	// 锁重试间隔时间
	RetryDelay time.Duration
	// 获取锁最大重试次数
	MaxRetries int
	// 续约间隔时间（默认过期时间的1/3）
	RenewalInterval time.Duration
	// 状态通道缓冲区大小
	StatusChanBuffer int
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		LockName:         "distributed-lock",
		Expiry:           30 * time.Second,
		RetryDelay:       2 * time.Second,
		MaxRetries:       3,
		RenewalInterval:  0, // 0表示自动计算
		StatusChanBuffer: 10,
	}
}

// NewLockRedsync 创建分布式锁
func NewLockRedsync(redisClient []*redis.Client, logger logx.Loggerx, config Config) *LockRedsync {
	// 合并默认配置
	if config.LockName == "" {
		config.LockName = DefaultConfig().LockName
	}
	if config.Expiry == 0 {
		config.Expiry = DefaultConfig().Expiry
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = DefaultConfig().RetryDelay
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultConfig().MaxRetries
	}
	if config.RenewalInterval == 0 {
		config.RenewalInterval = config.Expiry / 3
		if config.RenewalInterval < time.Second {
			config.RenewalInterval = time.Second
		}
	}
	if config.StatusChanBuffer <= 0 {
		config.StatusChanBuffer = DefaultConfig().StatusChanBuffer
	}

	// 创建连接池
	pools := make([]redRedis.Pool, 0, len(redisClient))
	for _, client := range redisClient {
		pools = append(pools, goredis.NewPool(client))
	}

	// 创建RedSync实例
	rs := redsync.New(pools...)
	mutex := rs.NewMutex(
		config.LockName,
		redsync.WithExpiry(config.Expiry),
		redsync.WithTries(config.MaxRetries),
		redsync.WithRetryDelay(config.RetryDelay),
	)

	ctx, cancel := context.WithCancel(context.Background())

	return &LockRedsync{
		rsMutex:    mutex,
		logger:     logger,
		lockName:   config.LockName,
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
		status:     LockStatusUnknown,
		statusChan: make(chan LockResult, config.StatusChanBuffer),
	}
}

// Start 启动锁服务
// 返回状态通道，用于接收锁状态变更通知
func (dl *LockRedsync) Start() <-chan LockResult {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.isRunning {
		dl.logger.Warn("锁服务已在运行中")
		return dl.statusChan
	}

	dl.isRunning = true
	dl.logger.Info("启动锁服务", logx.String("lockName", dl.lockName))

	// 启动锁获取协程
	dl.wg.Add(1)
	go dl.acquisitionLoop()

	// 启动状态发送协程（避免阻塞主逻辑）
	dl.wg.Add(1)
	go dl.statusBroadcastLoop()

	return dl.statusChan
}

// Stop 停止锁服务，释放所有资源
func (dl *LockRedsync) Stop() {
	dl.mu.Lock()
	if !dl.isRunning {
		dl.mu.Unlock()
		dl.logger.Info("锁服务未运行，无需停止")
		return
	}

	dl.logger.Info("停止锁服务...")
	dl.isRunning = false
	dl.mu.Unlock()

	// 取消上下文，通知所有goroutine退出
	dl.cancel()

	// 停止续约ticker
	if dl.renewalTicker != nil {
		dl.renewalTicker.Stop()
	}

	// 如果持有锁，尝试释放
	if dl.Status() == LockStatusAcquired {
		dl.mu.Lock()
		dl.releaseLock()
		dl.mu.Unlock()
	}

	// 等待所有goroutine退出
	dl.wg.Wait()

	// 关闭状态通道
	close(dl.statusChan)

	dl.logger.Info("锁服务已停止")
}

// Status 获取当前锁状态
func (dl *LockRedsync) Status() LockStatus {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return dl.status
}

// IsLocked 检查是否持有锁
func (dl *LockRedsync) IsLocked() bool {
	return dl.Status() == LockStatusAcquired
}

// acquisitionLoop 锁获取循环
func (dl *LockRedsync) acquisitionLoop() {
	defer dl.wg.Done()

	dl.logger.Debug("开始锁获取循环")

	// 初始延迟，避免立即重试
	initialDelay := time.NewTimer(time.Second)
	defer initialDelay.Stop()

	select {
	case <-initialDelay.C:
		// 继续执行
	case <-dl.ctx.Done():
		return
	}

	for {
		select {
		case <-dl.ctx.Done():
			dl.logger.Debug("锁获取循环收到停止信号")
			return

		default:
			// 尝试获取锁
			err := dl.tryAcquireLock()
			if err != nil {
				if errors.Is(err, ErrLockStopped) {
					return
				}
				dl.logger.Warn("获取锁失败，等待重试",
					logx.Error(err),
					logx.TimeDuration("retryDelay", dl.config.RetryDelay))

				// 等待重试
				select {
				case <-time.After(dl.config.RetryDelay):
					continue
				case <-dl.ctx.Done():
					return
				}
			}

			// 锁获取成功，启动续约
			dl.startRenewal()

			// 等待锁丢失或停止
			dl.waitForLockLossOrStop()

			// 停止续约
			dl.stopRenewal()
		}
	}
}

// tryAcquireLock 尝试获取锁
func (dl *LockRedsync) tryAcquireLock() error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if !dl.isRunning {
		return ErrLockStopped
	}

	dl.logger.Info("尝试获取分布式锁", logx.String("lockName", dl.lockName))

	// 尝试获取锁
	if err := dl.rsMutex.Lock(); err != nil {
		dl.updateStatus(LockStatusLost, err)
		return fmt.Errorf("获取锁失败: %w", err)
	}

	// 更新状态
	dl.status = LockStatusAcquired
	dl.acquiredTime = time.Now()
	dl.updateStatus(LockStatusAcquired, nil)
	dl.logger.Info("锁已获取",
		logx.String("lockName", dl.lockName),
		logx.TimeTime("acquiredTime", dl.acquiredTime))

	return nil
}

// releaseLock 释放锁
func (dl *LockRedsync) releaseLock() error {
	if dl.status != LockStatusAcquired {
		return ErrLockNotAcquired
	}

	dl.logger.Info("释放锁", logx.String("lockName", dl.lockName))

	// 尝试解锁
	if _, err := dl.rsMutex.Unlock(); err != nil {
		dl.logger.Error("释放锁失败", logx.Error(err))
		dl.updateStatus(LockStatusLost, err)
		return fmt.Errorf("释放锁失败: %w", err)
	}

	dl.updateStatus(LockStatusReleased, nil)
	return nil
}

// startRenewal 启动锁续约
func (dl *LockRedsync) startRenewal() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.renewalTicker != nil {
		return // 续约已在运行
	}

	dl.renewalTicker = time.NewTicker(dl.config.RenewalInterval)
	dl.logger.Debug("启动锁续约",
		logx.TimeDuration("interval", dl.config.RenewalInterval),
		logx.TimeDuration("expiry", dl.config.Expiry))

	dl.wg.Add(1)
	go dl.renewalLoop(dl.renewalTicker)
}

// stopRenewal 停止锁续约
func (dl *LockRedsync) stopRenewal() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.renewalTicker != nil {
		dl.renewalTicker.Stop()
		dl.renewalTicker = nil
		dl.logger.Debug("停止锁续约")
	}
}

// renewalLoop 锁续约循环
func (dl *LockRedsync) renewalLoop(ticker *time.Ticker) {
	defer dl.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			dl.logger.Error("续约循环发生panic", logx.Any("panic", r))
		}
	}()

	dl.logger.Debug("开始锁续约循环")

	for {
		select {
		case <-dl.ctx.Done():
			dl.logger.Debug("续约循环收到停止信号")
			return

		case <-ticker.C:
			if err := dl.doRenewal(); err != nil {
				if errors.Is(err, ErrLockStopped) {
					return
				}
				// 续约失败，锁可能已丢失
				dl.mu.Lock()
				if dl.status == LockStatusAcquired {
					dl.updateStatus(LockStatusLost, err)
					dl.logger.Error("锁续约失败，锁已丢失", logx.Error(err))
				}
				dl.mu.Unlock()
				return
			}
		}
	}
}

// doRenewal 执行锁续约
func (dl *LockRedsync) doRenewal() error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if !dl.isRunning {
		return ErrLockStopped
	}

	if dl.status != LockStatusAcquired {
		return ErrLockNotAcquired
	}

	// 计算锁持有时间
	holdDuration := time.Since(dl.acquiredTime)
	dl.logger.Debug("执行锁续约",
		logx.TimeDuration("holdDuration", holdDuration),
		logx.TimeDuration("expiry", dl.config.Expiry))

	// 执行续约
	if ok, err := dl.rsMutex.Extend(); err != nil || !ok {
		if err != nil {
			return fmt.Errorf("锁续约失败: %w", err)
		}
		return errors.New("锁续约返回失败")
	}

	dl.logger.Debug("锁续约成功",
		logx.TimeDuration("holdDuration", holdDuration))
	return nil
}

// waitForLockLossOrStop 等待锁丢失或停止信号
func (dl *LockRedsync) waitForLockLossOrStop() {
	// 创建一个检查锁状态的ticker
	checkTicker := time.NewTicker(dl.config.RenewalInterval)
	defer checkTicker.Stop()

	for {
		select {
		case <-dl.ctx.Done():
			return

		case <-checkTicker.C:
			dl.mu.RLock()
			status := dl.status
			dl.mu.RUnlock()

			if status != LockStatusAcquired {
				dl.logger.Debug("锁状态变更，退出等待", logx.Any("status", status))
				return
			}
		}
	}
}

// updateStatus 更新锁状态并发送通知
func (dl *LockRedsync) updateStatus(status LockStatus, err error) {
	oldStatus := dl.status
	dl.status = status

	// 只有在状态真正变更时才发送通知
	if oldStatus != status {
		// 非阻塞发送状态变更
		select {
		case dl.statusChan <- LockResult{Status: status, Error: err}:
			// 发送成功
		default:
			dl.logger.Warn("状态通道已满，无法发送状态变更",
				logx.Any("status", status),
				logx.Any("oldStatus", oldStatus))
		}
	}
}

// statusBroadcastLoop 状态广播循环，确保状态变更不会阻塞主逻辑
func (dl *LockRedsync) statusBroadcastLoop() {
	defer dl.wg.Done()

	dl.logger.Debug("开始状态广播循环")

	for result := range dl.statusChan {
		dl.logStatusChange(result)
	}
}

// logStatusChange 记录锁状态变更
func (dl *LockRedsync) logStatusChange(result LockResult) {
	statusText := "未知状态"
	switch result.Status {
	case LockStatusUnknown:
		statusText = "未知状态"
	case LockStatusAcquired:
		statusText = "锁已获取"
	case LockStatusLost:
		statusText = "锁丢失"
	case LockStatusReleased:
		statusText = "锁释放"
	}

	if result.Error != nil {
		dl.logger.Info("锁状态变更",
			logx.String("status", statusText),
			logx.Error(result.Error))
	} else {
		dl.logger.Info("锁状态变更",
			logx.String("status", statusText))
	}
}

// GetLockInfo 获取锁信息
func (dl *LockRedsync) GetLockInfo() map[string]interface{} {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	info := map[string]interface{}{
		"lockName":  dl.lockName,
		"status":    dl.status,
		"isRunning": dl.isRunning,
		"isLocked":  dl.status == LockStatusAcquired,
	}

	if dl.status == LockStatusAcquired && !dl.acquiredTime.IsZero() {
		info["acquiredTime"] = dl.acquiredTime
		info["holdDuration"] = time.Since(dl.acquiredTime).String()
	}

	return info
}
