package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hgg-6/pkgTool/v2/logx"
	"github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx"
	"github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/executor"
	"github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
	"github.com/robfig/cron/v3"
)

// CronScheduler Cron任务调度器
type CronScheduler struct {
	cron            *cron.Cron
	cronService     service.CronService
	executorFactory executor.ExecutorFactory
	redSync         redsyncx.RedSyncIn
	l               logx.Loggerx

	// 任务注册表
	jobRegistry map[int64]cron.EntryID // jobId -> cronEntryId
	mu          sync.RWMutex

	// 控制
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCronScheduler 创建调度器
func NewCronScheduler(
	cronService service.CronService,
	executorFactory executor.ExecutorFactory,
	redSync redsyncx.RedSyncIn,
	l logx.Loggerx,
) *CronScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &CronScheduler{
		cron:            cron.New(cron.WithSeconds()), // 支持秒级调度
		cronService:     cronService,
		executorFactory: executorFactory,
		redSync:         redSync,
		l:               l,
		jobRegistry:     make(map[int64]cron.EntryID),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start 启动调度器
func (s *CronScheduler) Start() error {
	s.l.Info("启动Cron调度器...")

	// 加载所有active状态的任务
	jobs, err := s.cronService.GetCronJobs(s.ctx)
	if err != nil {
		s.l.Error("加载任务列表失败", logx.Error(err))
		return err
	}

	// 注册所有任务
	for _, job := range jobs {
		if job.Status == domain.JobStatusActive {
			if err := s.AddJob(job); err != nil {
				s.l.Error("注册任务失败",
					logx.Int64("job_id", job.CronId),
					logx.String("job_name", job.Name),
					logx.Error(err),
				)
			}
		}
	}

	// 启动cron调度器
	s.cron.Start()
	s.l.Info("Cron调度器启动完成", logx.Int("job_count", len(s.jobRegistry)))

	return nil
}

// Stop 停止调度器
func (s *CronScheduler) Stop() {
	s.l.Info("停止Cron调度器...")
	s.cancel()

	ctx := s.cron.Stop()
	<-ctx.Done()

	s.l.Info("Cron调度器已停止")
}

// AddJob 添加任务到调度器
func (s *CronScheduler) AddJob(job domain.CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查任务是否已存在
	if _, exists := s.jobRegistry[job.CronId]; exists {
		return fmt.Errorf("job already exists: %d", job.CronId)
	}

	// 解析Cron表达式
	// 注意：调度器用 cron.WithSeconds() 创建（6 字段），必须用 cron.Parse（6 字段）解析，
	// 不能用 cron.ParseStandard（5 字段），否则用户写的 6 字段表达式会被拒绝、
	// 5 字段表达式在 6 字段调度器里行为错乱（P0-9）。
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(job.CronExpr)
	if err != nil {
		s.l.Error("解析Cron表达式失败",
			logx.Int64("job_id", job.CronId),
			logx.String("cron_expr", job.CronExpr),
			logx.Error(err),
		)
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// 创建任务执行函数（带分布式锁）
	jobFunc := s.createJobFunc(job)

	// 添加到cron调度器
	entryID := s.cron.Schedule(schedule, cron.FuncJob(jobFunc))
	s.jobRegistry[job.CronId] = entryID

	s.l.Info("任务已添加到调度器",
		logx.Int64("job_id", job.CronId),
		logx.String("job_name", job.Name),
		logx.String("cron_expr", job.CronExpr),
		logx.Int("entry_id", int(entryID)),
	)

	return nil
}

// RemoveJob 从调度器移除任务
func (s *CronScheduler) RemoveJob(jobId int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.jobRegistry[jobId]
	if !exists {
		return fmt.Errorf("job not found: %d", jobId)
	}

	s.cron.Remove(entryID)
	delete(s.jobRegistry, jobId)

	s.l.Info("任务已从调度器移除", logx.Int64("job_id", jobId))
	return nil
}

// UpdateJob 更新任务
func (s *CronScheduler) UpdateJob(job domain.CronJob) error {
	// 先移除旧任务
	if err := s.RemoveJob(job.CronId); err != nil {
		// 任务不存在也继续
		s.l.Warn("移除旧任务失败", logx.Int64("job_id", job.CronId), logx.Error(err))
	}

	// 如果是active状态，重新添加
	if job.Status == domain.JobStatusActive {
		return s.AddJob(job)
	}

	return nil
}

// createJobFunc 创建任务执行函数（带分布式锁）
func (s *CronScheduler) createJobFunc(job domain.CronJob) func() {
	return func() {
		// 使用分布式锁确保同一任务在集群中只执行一次
		lockKey := fmt.Sprintf("cron:lock:%d", job.CronId)
		mutex := s.redSync.CreateMutex(lockKey)

		// 尝试获取锁（非阻塞）
		if err := mutex.Lock(); err != nil {
			s.l.Debug("获取分布式锁失败，跳过本次执行",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Error(err),
			)
			return
		}
		// P0-8 修复：任务执行期间定期续约，避免长任务超过锁 TTL（默认 30s）
		// 被 Redis 自动释放，导致另一节点抢到锁造成双机并发执行。
		// 续约间隔取锁有效期的 1/3（与 redsync 推荐续约节奏一致），最低 1s。
		renewInterval := mutex.Until().Sub(time.Now()) / 3
		if renewInterval < time.Second {
			renewInterval = time.Second
		}
		renewCtx, renewCancel := context.WithCancel(context.Background())
		var renewWg sync.WaitGroup
		renewWg.Add(1)
		go func() {
			defer renewWg.Done()
			ticker := time.NewTicker(renewInterval)
			defer ticker.Stop()
			for {
				select {
				case <-renewCtx.Done():
					return
				case <-ticker.C:
					// 续约失败说明锁已被释放（如 Redis 故障或被强制删除），
					// 此时继续执行会有双机风险，记录错误并退出续约循环
					// （任务本身仍会跑完，但日志可追踪异常）。
					if ok, err := mutex.Extend(); err != nil || !ok {
						s.l.Error("分布式锁续约失败",
							logx.Int64("job_id", job.CronId),
							logx.Bool("ok", ok),
							logx.Error(err),
						)
						return
					}
				}
			}
		}()
		defer func() {
			// 先停止续约 goroutine，再释放锁，避免释放后续约又把锁加回去。
			renewCancel()
			renewWg.Wait()
			if ok, err := mutex.Unlock(); !ok || err != nil {
				s.l.Error("释放分布式锁失败",
					logx.Int64("job_id", job.CronId),
					logx.Error(err),
				)
			}
		}()

		// 更新任务状态为运行中
		if err := s.cronService.UpdateJobStatus(context.Background(), job.CronId, domain.JobStatusRunning); err != nil {
			s.l.Error("更新任务状态为运行中失败",
				logx.Int64("job_id", job.CronId),
				logx.Error(err),
			)
		}

		s.l.Info("开始执行任务",
			logx.Int64("job_id", job.CronId),
			logx.String("job_name", job.Name),
			logx.String("task_type", string(job.TaskType)),
		)

		// 执行任务
		if err := s.executeJob(job); err != nil {
			s.l.Error("任务执行失败",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Error(err),
			)
			// 执行失败后恢复状态为active
			if err := s.cronService.UpdateJobStatus(context.Background(), job.CronId, domain.JobStatusActive); err != nil {
				s.l.Error("恢复任务状态失败",
					logx.Int64("job_id", job.CronId),
					logx.Error(err),
				)
			}
		} else {
			// 执行成功后恢复状态为active
			if err := s.cronService.UpdateJobStatus(context.Background(), job.CronId, domain.JobStatusActive); err != nil {
				s.l.Error("恢复任务状态失败",
					logx.Int64("job_id", job.CronId),
					logx.Error(err),
				)
			}
		}
	}
}

// executeJob 执行任务
func (s *CronScheduler) executeJob(job domain.CronJob) error {
	// 获取执行器
	exec, err := s.executorFactory.GetExecutor(job.TaskType)
	if err != nil {
		s.sendExecutionAlert(job, fmt.Errorf("获取执行器失败: %w", err))
		return fmt.Errorf("获取执行器失败: %w", err)
	}

	// 执行任务（不设置额外超时，由执行器内部管理重试和每次尝试的超时）
	result, err := exec.Execute(s.ctx, job)
	if err != nil {
		s.l.Error("任务执行错误",
			logx.Int64("job_id", job.CronId),
			logx.Error(err),
		)
		s.sendExecutionAlert(job, err)
		return err
	}

	if result.Success {
		s.l.Info("任务执行成功",
			logx.Int64("job_id", job.CronId),
			logx.String("job_name", job.Name),
			logx.Int64("duration_ms", result.Duration),
			logx.String("message", result.Message),
		)
		return nil
	}

	// P0-12 修复：旧实现在 result.Success==false 时只 Warn + 告警，最后 return nil，
	// 导致上层 createJobFunc 的"执行失败"分支永远进不去，MaxRetry/失败计数等
	// 对外可见的失败语义全部失效。改为返回 error，让上层正确感知失败。
	s.l.Warn("任务执行失败",
		logx.Int64("job_id", job.CronId),
		logx.String("job_name", job.Name),
		logx.String("message", result.Message),
	)
	failErr := fmt.Errorf("任务执行失败: %s", result.Message)
	s.sendExecutionAlert(job, failErr)
	return failErr
}

// sendExecutionAlert 发送执行告警
func (s *CronScheduler) sendExecutionAlert(job domain.CronJob, err error) {
	// 这里可以实现告警逻辑，比如发送邮件、短信、企业微信通知等
	// 目前先记录日志，后续可以扩展为实际的告警渠道
	s.l.Error("发送任务执行告警",
		logx.Int64("job_id", job.CronId),
		logx.String("job_name", job.Name),
		logx.String("task_type", string(job.TaskType)),
		logx.Error(err),
	)
}

// GetJobCount 获取已注册任务数量
func (s *CronScheduler) GetJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobRegistry)
}
