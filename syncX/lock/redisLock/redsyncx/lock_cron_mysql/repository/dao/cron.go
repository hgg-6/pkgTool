package dao

import (
	"context"
	"errors"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDataRecordNotFound error = errors.New("数据不存在, 查询为空")
	ErrDuplicateData      error = errors.New("数据已存在, 重复添加")
)

type CronDb interface {
	FindById(ctx context.Context, id int64) (CronJob, error)
	FindAll(ctx context.Context) ([]CronJob, error)
	Insert(ctx context.Context, job CronJob) error
	Inserts(ctx context.Context, jobs []domain.CronJob) error
	Delete(ctx context.Context, id int64) error
	Deletes(ctx context.Context, ids []int64) error
	// 状态管理方法
	UpdateStatus(ctx context.Context, id int64, status JobStatus) error
	UpdateJob(ctx context.Context, job CronJob) error
}

type cronCronDb struct {
	db *gorm.DB
}

// NewCronDb 创建CronDb实例
func NewCronDb(db *gorm.DB) CronDb {
	return &cronCronDb{db: db}
}

func (c *cronCronDb) FindById(ctx context.Context, id int64) (CronJob, error) {
	var cronJob CronJob
	err := c.db.Model(&cronJob).WithContext(ctx).Where("cron_id = ?", id).First(&cronJob).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return CronJob{}, ErrDataRecordNotFound
	default:
		return cronJob, err
	}
}
func (c *cronCronDb) FindAll(ctx context.Context) ([]CronJob, error) {
	var cronJobs []CronJob
	err := c.db.Model(&CronJob{}).WithContext(ctx).Find(&cronJobs).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return []CronJob{}, ErrDataRecordNotFound
	default:
		return cronJobs, err
	}
}

func (c *cronCronDb) Insert(ctx context.Context, job CronJob) error {
	err := c.db.Model(&CronJob{}).WithContext(ctx).Create(&job).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (c *cronCronDb) Inserts(ctx context.Context, jobs []domain.CronJob) error {
	err := c.db.Model(&CronJob{}).WithContext(ctx).Create(&jobs).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return err
}

func (c *cronCronDb) Delete(ctx context.Context, id int64) error {
	return c.db.Model(&CronJob{}).WithContext(ctx).Where("cron_id = ?", id).Delete(&CronJob{}).Error
}
func (c *cronCronDb) Deletes(ctx context.Context, ids []int64) error {
	return c.db.Model(&CronJob{}).WithContext(ctx).Where("cron_id in ?", ids).Delete(&CronJob{}).Error
}

// UpdateStatus 更新任务状态
func (c *cronCronDb) UpdateStatus(ctx context.Context, id int64, status JobStatus) error {
	return c.db.Model(&CronJob{}).WithContext(ctx).Where("cron_id = ?", id).Update("status", status).Error
}

// UpdateJob 更新任务信息
func (c *cronCronDb) UpdateJob(ctx context.Context, job CronJob) error {
	return c.db.Model(&CronJob{}).WithContext(ctx).Where("cron_id = ?", job.CronId).Updates(&job).Error
}
