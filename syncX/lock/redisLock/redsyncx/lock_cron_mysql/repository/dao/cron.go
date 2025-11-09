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
}

type cronCronDb struct {
	db *gorm.DB
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
	err := c.db.Model(&cronJobs).WithContext(ctx).Where("cron_id >= ", 0).Find(&cronJobs).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return []CronJob{}, ErrDataRecordNotFound
	default:
		return cronJobs, err
	}
}

func (c *cronCronDb) Insert(ctx context.Context, job CronJob) error {
	var cron CronJob
	err := c.db.Model(&cron).WithContext(ctx).Where("cron_id = ?", job.ID).First(&cron).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return c.db.Model(&job).WithContext(ctx).Create(&job).Error
}

func (c *cronCronDb) Inserts(ctx context.Context, jobs []domain.CronJob) error {
	err := c.db.Model(&CronJob{}).WithContext(ctx).Create(&jobs).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if e.Number == duplicateError {
			return ErrDuplicateData
		}
	}
	return c.db.Model(&jobs).WithContext(ctx).Create(&jobs).Error
}

func (c *cronCronDb) Delete(ctx context.Context, id int64) error {
	return c.db.Model(&CronJob{}).WithContext(ctx).Where("cron_id = ?", id).Delete(&CronJob{}).Error
}
func (c *cronCronDb) Deletes(ctx context.Context, ids []int64) error {
	return c.db.Model(&CronJob{}).WithContext(ctx).Where("cron_id in ?", ids).Delete(&CronJob{}).Error
}
