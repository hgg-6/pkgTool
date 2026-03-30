package service

import (
	"context"
	"errors"
	"fmt"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"
)

// TaskValidator 任务校验接口（由 executor.ExecutorFactory 实现）
type TaskValidator interface {
	ValidateTask(ctx context.Context, job domain.CronJob) error
}

var (
	ErrDataRecordNotFound  error = repository.ErrDataRecordNotFound
	ErrDuplicateData       error = repository.ErrDuplicateData
	ErrInvalidStatusChange error = errors.New("无效的状态变更")
	ErrTaskValidateFailed  error = errors.New("任务配置校验失败")
)

type CronService interface {
	GetCronJob(ctx context.Context, id int64) (domain.CronJob, error)
	GetCronJobs(ctx context.Context) ([]domain.CronJob, error)
	AddCronJob(ctx context.Context, job domain.CronJob) error
	AddCronJobs(ctx context.Context, jobs []domain.CronJob) error
	DelCronJob(ctx context.Context, id int64) error
	DelCronJobs(ctx context.Context, ids []int64) error
	// 状态管理方法
	StartJob(ctx context.Context, id int64) error
	PauseJob(ctx context.Context, id int64) error
	ResumeJob(ctx context.Context, id int64) error
	UpdateJobStatus(ctx context.Context, id int64, status domain.JobStatus) error
	// 设置调度器
	SetScheduler(scheduler Scheduler)
	// 设置任务校验器
	SetTaskValidator(validator TaskValidator)
}

// Scheduler 调度器接口
type Scheduler interface {
	UpdateJob(job domain.CronJob) error
}

type cronService struct {
	cronRepo      repository.CronRepository
	scheduler     Scheduler
	taskValidator TaskValidator
}

// NewCronService 创建CronService实例
func NewCronService(cronRepo repository.CronRepository, scheduler Scheduler) CronService {
	return &cronService{cronRepo: cronRepo, scheduler: scheduler}
}

// SetTaskValidator 设置任务校验器（用于创建任务时校验）
func (c *cronService) SetTaskValidator(validator TaskValidator) {
	c.taskValidator = validator
}

// SetScheduler 设置调度器
func (c *cronService) SetScheduler(scheduler Scheduler) {
	c.scheduler = scheduler
}

func (c *cronService) GetCronJob(ctx context.Context, id int64) (domain.CronJob, error) {
	return c.cronRepo.FindById(ctx, id)
}

func (c *cronService) GetCronJobs(ctx context.Context) ([]domain.CronJob, error) {
	return c.cronRepo.FindAll(ctx)
}

func (c *cronService) AddCronJob(ctx context.Context, job domain.CronJob) error {
	if c.taskValidator != nil {
		if err := c.taskValidator.ValidateTask(ctx, job); err != nil {
			return fmt.Errorf("%w: %v", ErrTaskValidateFailed, err)
		}
	}
	return c.cronRepo.CreateCron(ctx, job)
}
func (c *cronService) AddCronJobs(ctx context.Context, jobs []domain.CronJob) error {
	if c.taskValidator != nil {
		for _, job := range jobs {
			if err := c.taskValidator.ValidateTask(ctx, job); err != nil {
				return fmt.Errorf("%w: 任务[%s] %v", ErrTaskValidateFailed, job.Name, err)
			}
		}
	}
	return c.cronRepo.CreateCrons(ctx, jobs)
}

func (c *cronService) DelCronJob(ctx context.Context, id int64) error {
	return c.cronRepo.DelCron(ctx, id)
}
func (c *cronService) DelCronJobs(ctx context.Context, ids []int64) error {
	return c.cronRepo.DelCrons(ctx, ids)
}

// StartJob 启动任务
func (c *cronService) StartJob(ctx context.Context, id int64) error {
	// 获取当前任务状态
	job, err := c.cronRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 只有暂停状态的任务才能启动
	if job.Status != domain.JobStatusPaused {
		return ErrInvalidStatusChange
	}

	// 更新任务状态
	if err := c.cronRepo.UpdateStatus(ctx, id, domain.JobStatusActive); err != nil {
		return err
	}

	// 重新获取更新后的任务
	updatedJob, err := c.cronRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 通知调度器更新任务
	if c.scheduler != nil {
		return c.scheduler.UpdateJob(updatedJob)
	}

	return nil
}

// PauseJob 暂停任务
func (c *cronService) PauseJob(ctx context.Context, id int64) error {
	// 获取当前任务状态
	job, err := c.cronRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 只有活跃或运行中的任务才能暂停
	if job.Status != domain.JobStatusActive && job.Status != domain.JobStatusRunning {
		return ErrInvalidStatusChange
	}

	// 更新任务状态
	if err := c.cronRepo.UpdateStatus(ctx, id, domain.JobStatusPaused); err != nil {
		return err
	}

	// 重新获取更新后的任务
	updatedJob, err := c.cronRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 通知调度器更新任务
	if c.scheduler != nil {
		return c.scheduler.UpdateJob(updatedJob)
	}

	return nil
}

// ResumeJob 恢复任务
func (c *cronService) ResumeJob(ctx context.Context, id int64) error {
	// 获取当前任务状态
	job, err := c.cronRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 只有暂停状态的任务才能恢复
	if job.Status != domain.JobStatusPaused {
		return ErrInvalidStatusChange
	}

	// 更新任务状态
	if err := c.cronRepo.UpdateStatus(ctx, id, domain.JobStatusActive); err != nil {
		return err
	}

	// 重新获取更新后的任务
	updatedJob, err := c.cronRepo.FindById(ctx, id)
	if err != nil {
		return err
	}

	// 通知调度器更新任务
	if c.scheduler != nil {
		return c.scheduler.UpdateJob(updatedJob)
	}

	return nil
}

// UpdateJobStatus 更新任务状态（直接更新，不检查状态转换）
func (c *cronService) UpdateJobStatus(ctx context.Context, id int64, status domain.JobStatus) error {
	return c.cronRepo.UpdateStatus(ctx, id, status)
}
