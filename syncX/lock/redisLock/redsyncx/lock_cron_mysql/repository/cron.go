package repository

import (
	"context"
	"database/sql"

	"gitee.com/hgg_test/pkg_tool/v2/sliceX"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
)

var (
	ErrDataRecordNotFound error = dao.ErrDataRecordNotFound
	ErrDuplicateData      error = dao.ErrDuplicateData
)

type CronRepository interface {
	FindById(ctx context.Context, id int64) (domain.CronJob, error)
	FindAll(ctx context.Context) ([]domain.CronJob, error)
	CreateCron(ctx context.Context, job domain.CronJob) error
	CreateCrons(ctx context.Context, jobs []domain.CronJob) error
	DelCron(ctx context.Context, id int64) error
	DelCrons(ctx context.Context, ids []int64) error
	// 状态管理方法
	UpdateStatus(ctx context.Context, id int64, status domain.JobStatus) error
	UpdateJob(ctx context.Context, job domain.CronJob) error
}

type cronRepository struct {
	db dao.CronDb
}

// NewCronRepository 创建CronRepository实例
func NewCronRepository(db dao.CronDb) CronRepository {
	return &cronRepository{db: db}
}

func (c *cronRepository) FindById(ctx context.Context, id int64) (domain.CronJob, error) {
	cron, err := c.db.FindById(ctx, id)
	if err != nil {
		return domain.CronJob{}, err
	}
	return toDomain(cron), nil
}
func (c *cronRepository) FindAll(ctx context.Context) ([]domain.CronJob, error) {
	crons, err := c.db.FindAll(ctx)
	if err != nil {
		return []domain.CronJob{}, err
	}
	return sliceX.Map[dao.CronJob, domain.CronJob](crons, func(idx int, src dao.CronJob) domain.CronJob {
		return toDomain(src)
	}), nil
}

func (c *cronRepository) CreateCron(ctx context.Context, job domain.CronJob) error {
	return c.db.Insert(ctx, toEntity(job))
}

func (c *cronRepository) CreateCrons(ctx context.Context, jobs []domain.CronJob) error {
	return c.db.Inserts(ctx, jobs)
}

func (c *cronRepository) DelCron(ctx context.Context, id int64) error {
	return c.db.Delete(ctx, id)
}

func (c *cronRepository) DelCrons(ctx context.Context, ids []int64) error {
	return c.db.Deletes(ctx, ids)
}

// UpdateStatus 更新任务状态
func (c *cronRepository) UpdateStatus(ctx context.Context, id int64, status domain.JobStatus) error {
	return c.db.UpdateStatus(ctx, id, dao.JobStatus(status))
}

// UpdateJob 更新任务信息
func (c *cronRepository) UpdateJob(ctx context.Context, job domain.CronJob) error {
	return c.db.UpdateJob(ctx, toEntity(job))
}

func toDomain(cron dao.CronJob) domain.CronJob {
	return domain.CronJob{
		ID:          cron.ID,
		CronId:      cron.CronId,
		Name:        cron.Name,
		Description: cron.Description.String,
		CronExpr:    cron.CronExpr,
		TaskType:    domain.TaskType(cron.TaskType),
		Status:      domain.JobStatus(cron.Status),
		MaxRetry:    cron.MaxRetry,
		Timeout:     cron.Timeout,
		Ctime:       cron.Ctime,
		Utime:       cron.Utime,
	}
}

func toEntity(cron domain.CronJob) dao.CronJob {
	return dao.CronJob{
		ID:     cron.ID,
		CronId: cron.CronId,
		Name:   cron.Name,
		Description: sql.NullString{
			String: cron.Description,
			Valid:  cron.Description != "",
		},
		CronExpr: cron.CronExpr,
		TaskType: dao.TaskType(cron.TaskType),
		Status:   dao.JobStatus(cron.Status),
		MaxRetry: cron.MaxRetry,
		Timeout:  cron.Timeout,
		Ctime:    cron.Ctime,
		Utime:    cron.Utime,
	}
}
