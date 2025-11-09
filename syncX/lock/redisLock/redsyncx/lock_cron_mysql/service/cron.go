package service

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"
)

var (
	ErrDataRecordNotFound error = repository.ErrDataRecordNotFound
	ErrDuplicateData      error = repository.ErrDuplicateData
)

type CronService interface {
	GetCronJob(ctx context.Context, id int64) (domain.CronJob, error)
	GetCronJobs(ctx context.Context) ([]domain.CronJob, error)
	AddCronJob(ctx context.Context, job domain.CronJob) error
	AddCronJobs(ctx context.Context, jobs []domain.CronJob) error
	DelCronJob(ctx context.Context, id int64) error
	DelCronJobs(ctx context.Context, ids []int64) error
}

type cronService struct {
	cronRepo repository.CronRepository
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
