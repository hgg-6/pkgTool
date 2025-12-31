package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// JobHistoryDAO 任务执行历史数据访问接口
type JobHistoryDAO interface {
	// Insert 插入执行历史记录
	Insert(ctx context.Context, history JobHistory) error
	// FindById 根据ID查询执行历史
	FindById(ctx context.Context, id int64) (JobHistory, error)
	// FindByCronId 根据任务ID查询执行历史列表
	FindByCronId(ctx context.Context, cronId int64, limit, offset int) ([]JobHistory, error)
	// FindByStatus 根据执行状态查询执行历史列表
	FindByStatus(ctx context.Context, status ExecutionStatus, limit, offset int) ([]JobHistory, error)
	// FindByTimeRange 根据时间范围查询执行历史列表
	FindByTimeRange(ctx context.Context, startTime, endTime int64, limit, offset int) ([]JobHistory, error)
	// CountByCronId 统计指定任务的执行历史数量
	CountByCronId(ctx context.Context, cronId int64) (int64, error)
	// CountByStatus 统计指定状态的执行历史数量
	CountByStatus(ctx context.Context, status ExecutionStatus) (int64, error)
	// DeleteById 根据ID删除执行历史
	DeleteById(ctx context.Context, id int64) error
	// DeleteByCronId 删除指定任务的所有执行历史
	DeleteByCronId(ctx context.Context, cronId int64) error
	// DeleteBeforeTime 删除指定时间之前的执行历史
	DeleteBeforeTime(ctx context.Context, beforeTime int64) error
	// GetLatestByCronId 获取指定任务的最新执行历史
	GetLatestByCronId(ctx context.Context, cronId int64) (JobHistory, error)
	// GetStatistics 获取指定任务的执行统计信息
	GetStatistics(ctx context.Context, cronId int64, days int) (map[string]interface{}, error)
}

type jobHistoryDAO struct {
	db *gorm.DB
}

// NewJobHistoryDAO 创建JobHistoryDAO实例
func NewJobHistoryDAO(db *gorm.DB) JobHistoryDAO {
	return &jobHistoryDAO{db: db}
}

func (j *jobHistoryDAO) Insert(ctx context.Context, history JobHistory) error {
	return j.db.WithContext(ctx).Create(&history).Error
}

func (j *jobHistoryDAO) FindById(ctx context.Context, id int64) (JobHistory, error) {
	var history JobHistory
	err := j.db.WithContext(ctx).Where("id = ?", id).First(&history).Error
	if err == gorm.ErrRecordNotFound {
		return JobHistory{}, ErrDataRecordNotFound
	}
	return history, err
}

func (j *jobHistoryDAO) FindByCronId(ctx context.Context, cronId int64, limit, offset int) ([]JobHistory, error) {
	var histories []JobHistory
	err := j.db.WithContext(ctx).
		Where("cron_id = ?", cronId).
		Order("start_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&histories).Error
	return histories, err
}

func (j *jobHistoryDAO) FindByStatus(ctx context.Context, status ExecutionStatus, limit, offset int) ([]JobHistory, error) {
	var histories []JobHistory
	err := j.db.WithContext(ctx).
		Where("status = ?", status).
		Order("start_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&histories).Error
	return histories, err
}

func (j *jobHistoryDAO) FindByTimeRange(ctx context.Context, startTime, endTime int64, limit, offset int) ([]JobHistory, error) {
	var histories []JobHistory
	err := j.db.WithContext(ctx).
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Order("start_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&histories).Error
	return histories, err
}

func (j *jobHistoryDAO) CountByCronId(ctx context.Context, cronId int64) (int64, error) {
	var count int64
	err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ?", cronId).
		Count(&count).Error
	return count, err
}

func (j *jobHistoryDAO) CountByStatus(ctx context.Context, status ExecutionStatus) (int64, error) {
	var count int64
	err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

func (j *jobHistoryDAO) DeleteById(ctx context.Context, id int64) error {
	return j.db.WithContext(ctx).Where("id = ?", id).Delete(&JobHistory{}).Error
}

func (j *jobHistoryDAO) DeleteByCronId(ctx context.Context, cronId int64) error {
	return j.db.WithContext(ctx).Where("cron_id = ?", cronId).Delete(&JobHistory{}).Error
}

func (j *jobHistoryDAO) DeleteBeforeTime(ctx context.Context, beforeTime int64) error {
	return j.db.WithContext(ctx).Where("start_time < ?", beforeTime).Delete(&JobHistory{}).Error
}

func (j *jobHistoryDAO) GetLatestByCronId(ctx context.Context, cronId int64) (JobHistory, error) {
	var history JobHistory
	err := j.db.WithContext(ctx).
		Where("cron_id = ?", cronId).
		Order("start_time DESC").
		First(&history).Error
	if err == gorm.ErrRecordNotFound {
		return JobHistory{}, ErrDataRecordNotFound
	}
	return history, err
}

func (j *jobHistoryDAO) GetStatistics(ctx context.Context, cronId int64, days int) (map[string]interface{}, error) {
	// 计算时间范围（最近N天）
	endTime := time.Now().Unix()
	startTime := time.Now().AddDate(0, 0, -days).Unix()

	stats := make(map[string]interface{})

	// 总执行次数
	var totalCount int64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND start_time >= ? AND start_time <= ?", cronId, startTime, endTime).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// 成功次数
	var successCount int64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND status = ? AND start_time >= ? AND start_time <= ?", cronId, ExecutionStatusSuccess, startTime, endTime).
		Count(&successCount).Error; err != nil {
		return nil, err
	}
	stats["success_count"] = successCount

	// 失败次数
	var failureCount int64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND status = ? AND start_time >= ? AND start_time <= ?", cronId, ExecutionStatusFailure, startTime, endTime).
		Count(&failureCount).Error; err != nil {
		return nil, err
	}
	stats["failure_count"] = failureCount

	// 超时次数
	var timeoutCount int64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND status = ? AND start_time >= ? AND start_time <= ?", cronId, ExecutionStatusTimeout, startTime, endTime).
		Count(&timeoutCount).Error; err != nil {
		return nil, err
	}
	stats["timeout_count"] = timeoutCount

	// 平均执行时长
	var avgDuration float64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND start_time >= ? AND start_time <= ?", cronId, startTime, endTime).
		Select("AVG(duration) as avg_duration").
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	stats["avg_duration"] = avgDuration

	// 最大执行时长
	var maxDuration int64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND start_time >= ? AND start_time <= ?", cronId, startTime, endTime).
		Select("MAX(duration) as max_duration").
		Scan(&maxDuration).Error; err != nil {
		return nil, err
	}
	stats["max_duration"] = maxDuration

	// 最小执行时长
	var minDuration int64
	if err := j.db.WithContext(ctx).
		Model(&JobHistory{}).
		Where("cron_id = ? AND start_time >= ? AND start_time <= ?", cronId, startTime, endTime).
		Select("MIN(duration) as min_duration").
		Scan(&minDuration).Error; err != nil {
		return nil, err
	}
	stats["min_duration"] = minDuration

	// 成功率
	if totalCount > 0 {
		stats["success_rate"] = float64(successCount) / float64(totalCount) * 100
	} else {
		stats["success_rate"] = 0.0
	}

	return stats, nil
}
