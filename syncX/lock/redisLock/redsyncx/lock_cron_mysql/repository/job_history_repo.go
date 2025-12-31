package repository

import (
	"context"
	"database/sql"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository/dao"
)

// JobHistoryRepository 任务执行历史仓储接口
type JobHistoryRepository interface {
	// Create 创建执行历史记录
	Create(ctx context.Context, history domain.JobHistory) error
	// FindById 根据ID查询执行历史
	FindById(ctx context.Context, id int64) (domain.JobHistory, error)
	// FindByCronId 根据任务ID查询执行历史列表
	FindByCronId(ctx context.Context, cronId int64, limit, offset int) ([]domain.JobHistory, error)
	// FindByStatus 根据执行状态查询执行历史列表
	FindByStatus(ctx context.Context, status domain.ExecutionStatus, limit, offset int) ([]domain.JobHistory, error)
	// FindByTimeRange 根据时间范围查询执行历史列表
	FindByTimeRange(ctx context.Context, startTime, endTime int64, limit, offset int) ([]domain.JobHistory, error)
	// CountByCronId 统计指定任务的执行历史数量
	CountByCronId(ctx context.Context, cronId int64) (int64, error)
	// CountByStatus 统计指定状态的执行历史数量
	CountByStatus(ctx context.Context, status domain.ExecutionStatus) (int64, error)
	// DeleteById 根据ID删除执行历史
	DeleteById(ctx context.Context, id int64) error
	// DeleteByCronId 删除指定任务的所有执行历史
	DeleteByCronId(ctx context.Context, cronId int64) error
	// DeleteBeforeTime 删除指定时间之前的执行历史
	DeleteBeforeTime(ctx context.Context, beforeTime int64) error
	// GetLatestByCronId 获取指定任务的最新执行历史
	GetLatestByCronId(ctx context.Context, cronId int64) (domain.JobHistory, error)
	// GetStatistics 获取指定任务的执行统计信息
	GetStatistics(ctx context.Context, cronId int64, days int) (map[string]interface{}, error)
}

type jobHistoryRepository struct {
	dao dao.JobHistoryDAO
}

// NewJobHistoryRepository 创建JobHistoryRepository实例
func NewJobHistoryRepository(dao dao.JobHistoryDAO) JobHistoryRepository {
	return &jobHistoryRepository{dao: dao}
}

func (j *jobHistoryRepository) Create(ctx context.Context, history domain.JobHistory) error {
	return j.dao.Insert(ctx, toHistoryEntity(history))
}

func (j *jobHistoryRepository) FindById(ctx context.Context, id int64) (domain.JobHistory, error) {
	entity, err := j.dao.FindById(ctx, id)
	if err != nil {
		return domain.JobHistory{}, err
	}
	return toHistoryDomain(entity), nil
}

func (j *jobHistoryRepository) FindByCronId(ctx context.Context, cronId int64, limit, offset int) ([]domain.JobHistory, error) {
	entities, err := j.dao.FindByCronId(ctx, cronId, limit, offset)
	if err != nil {
		return nil, err
	}

	histories := make([]domain.JobHistory, len(entities))
	for i, entity := range entities {
		histories[i] = toHistoryDomain(entity)
	}
	return histories, nil
}

func (j *jobHistoryRepository) FindByStatus(ctx context.Context, status domain.ExecutionStatus, limit, offset int) ([]domain.JobHistory, error) {
	entities, err := j.dao.FindByStatus(ctx, dao.ExecutionStatus(status), limit, offset)
	if err != nil {
		return nil, err
	}

	histories := make([]domain.JobHistory, len(entities))
	for i, entity := range entities {
		histories[i] = toHistoryDomain(entity)
	}
	return histories, nil
}

func (j *jobHistoryRepository) FindByTimeRange(ctx context.Context, startTime, endTime int64, limit, offset int) ([]domain.JobHistory, error) {
	entities, err := j.dao.FindByTimeRange(ctx, startTime, endTime, limit, offset)
	if err != nil {
		return nil, err
	}

	histories := make([]domain.JobHistory, len(entities))
	for i, entity := range entities {
		histories[i] = toHistoryDomain(entity)
	}
	return histories, nil
}

func (j *jobHistoryRepository) CountByCronId(ctx context.Context, cronId int64) (int64, error) {
	return j.dao.CountByCronId(ctx, cronId)
}

func (j *jobHistoryRepository) CountByStatus(ctx context.Context, status domain.ExecutionStatus) (int64, error) {
	return j.dao.CountByStatus(ctx, dao.ExecutionStatus(status))
}

func (j *jobHistoryRepository) DeleteById(ctx context.Context, id int64) error {
	return j.dao.DeleteById(ctx, id)
}

func (j *jobHistoryRepository) DeleteByCronId(ctx context.Context, cronId int64) error {
	return j.dao.DeleteByCronId(ctx, cronId)
}

func (j *jobHistoryRepository) DeleteBeforeTime(ctx context.Context, beforeTime int64) error {
	return j.dao.DeleteBeforeTime(ctx, beforeTime)
}

func (j *jobHistoryRepository) GetLatestByCronId(ctx context.Context, cronId int64) (domain.JobHistory, error) {
	entity, err := j.dao.GetLatestByCronId(ctx, cronId)
	if err != nil {
		return domain.JobHistory{}, err
	}
	return toHistoryDomain(entity), nil
}

func (j *jobHistoryRepository) GetStatistics(ctx context.Context, cronId int64, days int) (map[string]interface{}, error) {
	return j.dao.GetStatistics(ctx, cronId, days)
}

// toHistoryDomain 将DAO实体转换为Domain实体
func toHistoryDomain(entity dao.JobHistory) domain.JobHistory {
	return domain.JobHistory{
		ID:           entity.ID,
		CronId:       entity.CronId,
		JobName:      entity.JobName,
		Status:       domain.ExecutionStatus(entity.Status),
		StartTime:    entity.StartTime,
		EndTime:      entity.EndTime,
		Duration:     entity.Duration,
		RetryCount:   entity.RetryCount,
		ErrorMessage: entity.ErrorMessage.String,
		Result:       entity.Result.String,
		Ctime:        entity.Ctime,
	}
}

// toHistoryEntity 将Domain实体转换为DAO实体
func toHistoryEntity(history domain.JobHistory) dao.JobHistory {
	return dao.JobHistory{
		ID:         history.ID,
		CronId:     history.CronId,
		JobName:    history.JobName,
		Status:     dao.ExecutionStatus(history.Status),
		StartTime:  history.StartTime,
		EndTime:    history.EndTime,
		Duration:   history.Duration,
		RetryCount: history.RetryCount,
		ErrorMessage: sql.NullString{
			String: history.ErrorMessage,
			Valid:  history.ErrorMessage != "",
		},
		Result: sql.NullString{
			String: history.Result,
			Valid:  history.Result != "",
		},
		Ctime: history.Ctime,
	}
}
