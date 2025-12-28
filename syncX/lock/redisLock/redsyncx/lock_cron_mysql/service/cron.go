package service

import (
	"context"
	"errors"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"
)

var (
	ErrDataRecordNotFound  error = repository.ErrDataRecordNotFound
	ErrDuplicateData       error = repository.ErrDuplicateData
	ErrInvalidStatusChange error = errors.New("无效的状态变更")
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
}

type cronService struct {
	cronRepo repository.CronRepository
}

// NewCronService 创建CronService实例
func NewCronService(cronRepo repository.CronRepository) CronService {
	return &cronService{cronRepo: cronRepo}
}

func (c *cronService) GetCronJob(ctx context.Context, id int64) (domain.CronJob, error) {
	return c.cronRepo.FindById(ctx, id)
}

func (c *cronService) GetCronJobs(ctx context.Context) ([]domain.CronJob, error) {
	return c.cronRepo.FindAll(ctx)
}

func (c *cronService) AddCronJob(ctx context.Context, job domain.CronJob) error {
	return c.cronRepo.CreateCron(ctx, job)
}
func (c *cronService) AddCronJobs(ctx context.Context, jobs []domain.CronJob) error {
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

	return c.cronRepo.UpdateStatus(ctx, id, domain.JobStatusActive)
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

	return c.cronRepo.UpdateStatus(ctx, id, domain.JobStatusPaused)
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

	return c.cronRepo.UpdateStatus(ctx, id, domain.JobStatusActive)
}

// UpdateJobStatus 更新任务状态（直接更新，不检查状态转换）
func (c *cronService) UpdateJobStatus(ctx context.Context, id int64, status domain.JobStatus) error {
	return c.cronRepo.UpdateStatus(ctx, id, status)
}
