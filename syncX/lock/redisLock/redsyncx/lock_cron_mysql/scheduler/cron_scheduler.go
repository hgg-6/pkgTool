package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/executor"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
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
	schedule, err := cron.ParseStandard(job.CronExpr)
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
		defer func() {
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
		return fmt.Errorf("获取执行器失败: %w", err)
	}

	// 执行任务
	ctx, cancel := context.WithTimeout(s.ctx, time.Duration(job.Timeout)*time.Second)
	defer cancel()

	result, err := exec.Execute(ctx, job)
	if err != nil {
		s.l.Error("任务执行错误",
			logx.Int64("job_id", job.CronId),
			logx.Error(err),
		)
		return err
	}

	if result.Success {
		s.l.Info("任务执行成功",
			logx.Int64("job_id", job.CronId),
			logx.String("job_name", job.Name),
			logx.Int64("duration_ms", result.Duration),
			logx.String("message", result.Message),
		)
	} else {
		s.l.Warn("任务执行失败",
			logx.Int64("job_id", job.CronId),
			logx.String("job_name", job.Name),
			logx.String("message", result.Message),
		)
	}

	return nil
}

// GetJobCount 获取已注册任务数量
func (s *CronScheduler) GetJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobRegistry)
}
