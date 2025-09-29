package redsyncx

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/go-redsync/redsync/v4"
	redRedis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// LockStatus 锁状态
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

// LockRedsync 分布式锁
type LockRedsync struct {
	rsMutex  *redsync.Mutex
	logger   logx.Loggerx
	lockName string

	// 状态控制
	// 锁状态
	statusChan chan LockResult
	// 锁状态变更
	stopRenewal chan struct{}
	// 锁状态变更完成
	renewalDone chan struct{}
	// 锁状态
	isLocked bool
	mutex    sync.RWMutex

	// 配置
	// 锁过期时间
	expiry time.Duration
	// 锁重试间隔时间
	retryDelay time.Duration
	// 获取锁最大重试次数
	maxRetries int
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
}

// NewLockRedsync 基于redis创建分布式锁
func NewLockRedsync(redisClient []*redis.Client, logger logx.Loggerx, config Config) *LockRedsync {
	if config.LockName == "" {
		config.LockName = "distributed-lock"
	}
	if config.Expiry == 0 {
		config.Expiry = 30 * time.Second
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 2 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	// 创建连接池
	pools := make([]redRedis.Pool, 0, len(redisClient))
	for _, client := range redisClient {
		pools = append(pools, goredis.NewPool(client))
	}

	// 创建 RedSync 实例
	rs := redsync.New(pools...)
	mutex := rs.NewMutex(
		config.LockName,
		redsync.WithExpiry(config.Expiry),
		redsync.WithTries(config.MaxRetries),
		redsync.WithRetryDelay(config.RetryDelay),
	)

	return &LockRedsync{
		rsMutex:     mutex,
		logger:      logger,
		lockName:    config.LockName,
		statusChan:  make(chan LockResult, 10),
		stopRenewal: make(chan struct{}),
		renewalDone: make(chan struct{}),
		expiry:      config.Expiry,
		retryDelay:  config.RetryDelay,
		maxRetries:  config.MaxRetries,
	}
}

// Start 启动锁获取和续约
func (dl *LockRedsync) Start() <-chan LockResult {
	//go dl.acquireAndRenew()
	go func() {
		for {
			dl.acquireAndRenew()
			// 等待下一次尝试
			time.Sleep(time.Second)
		}
	}()
	// 监听锁状态变化，避免statusChan阻塞
	go func() {
		for res := range dl.statusChan {
			dl.logger.Info("锁状态变更提示", logx.Any("锁状态", res.Status),
				logx.String("状态0", "锁未知状态"),
				logx.String("状态1", "锁已获取"),
				logx.String("状态2", "锁未持有"),
				logx.String("状态3", "锁释放"),
				logx.Error(res.Error))
		}
	}()
	return dl.statusChan
}

// Stop 停止锁并释放资源
func (dl *LockRedsync) Stop() {
	dl.logger.Info("停止锁...")

	// 先检查是否持有锁，避免无效操作
	dl.mutex.RLock()
	isLocked := dl.isLocked
	dl.mutex.RUnlock()
	if !isLocked {
		dl.logger.Info("锁未被持有，无需停止")
		return
	}

	// 发送停止信号，通知acquireAndRenew协程退出
	close(dl.stopRenewal)

	// 新增：用context控制等待renewalDone的超时，避免永久死锁
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-dl.renewalDone:
		dl.logger.Info("等待renewalDone完成")
	case <-ctx.Done():
		dl.logger.Warn("等待renewalDone超时，强制继续")
	}
}

// IsLocked 检查是否持有锁
func (dl *LockRedsync) IsLocked() bool {
	dl.mutex.RLock()
	defer dl.mutex.RUnlock()
	return dl.isLocked
}

// acquireAndRenew 获取锁并启动续约
func (dl *LockRedsync) acquireAndRenew() {
	dl.logger.Info("尝试获取分布式锁...")
	defer func() {
		dl.mutex.Lock()
		dl.isLocked = false // 无论如何，退出时标记为未持有
		dl.mutex.Unlock()
	}()

	// 尝试获取锁：非阻塞发送状态
	if err := dl.rsMutex.Lock(); err != nil {
		select {
		case dl.statusChan <- LockResult{Status: LockStatusLost, Error: err}:
		default:
			dl.logger.Warn("statusChan已满，无法发送LockStatusLost")
		}
		return
	}

	// 标记持有锁：非阻塞发送状态
	dl.mutex.Lock()
	dl.isLocked = true
	dl.mutex.Unlock()
	select {
	case dl.statusChan <- LockResult{Status: LockStatusAcquired, Error: nil}:
	default:
		dl.logger.Warn("statusChan已满，无法发送LockStatusAcquired")
	}
	dl.logger.Info("锁已获取")

	// 启动续约
	renewalErr := make(chan error, 1)
	go dl.startRenewal(renewalErr)

	// 监听停止或续约失败：非阻塞发送状态
	select {
	case err := <-renewalErr:
		select {
		case dl.statusChan <- LockResult{Status: LockStatusLost, Error: err}:
		default:
			dl.logger.Warn("statusChan已满，无法发送LockStatusLost")
		}
	case <-dl.stopRenewal:
		select {
		case dl.statusChan <- LockResult{Status: LockStatusReleased, Error: nil}:
		default:
			dl.logger.Warn("statusChan已满，无法发送LockStatusReleased")
		}
	}
}

// startRenewal 启动锁续约
func (dl *LockRedsync) startRenewal(renewalErr chan<- error) {
	defer close(dl.renewalDone)

	renewalInterval := dl.expiry / 3
	if renewalInterval < time.Second {
		renewalInterval = time.Second
	}

	ticker := time.NewTicker(renewalInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if ok, err := dl.rsMutex.Extend(); !ok || err != nil {
				renewalErr <- err
				return
			}
		case <-dl.stopRenewal:
			return
		}
	}
}
