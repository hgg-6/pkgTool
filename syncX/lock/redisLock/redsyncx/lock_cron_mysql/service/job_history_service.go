package service

import (
	"context"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/repository"
)

// JobHistoryService 任务执行历史服务接口
type JobHistoryService interface {
	// RecordHistory 记录任务执行历史
	RecordHistory(ctx context.Context, history domain.JobHistory) error
	// GetHistory 获取单条执行历史
	GetHistory(ctx context.Context, id int64) (domain.JobHistory, error)
	// GetHistoryList 获取任务的执行历史列表
	GetHistoryList(ctx context.Context, cronId int64, page, pageSize int) ([]domain.JobHistory, int64, error)
	// GetHistoryListByStatus 根据状态获取执行历史列表
	GetHistoryListByStatus(ctx context.Context, status domain.ExecutionStatus, page, pageSize int) ([]domain.JobHistory, int64, error)
	// GetHistoryListByTimeRange 根据时间范围获取执行历史列表
	GetHistoryListByTimeRange(ctx context.Context, startTime, endTime int64, page, pageSize int) ([]domain.JobHistory, error)
	// GetLatestHistory 获取任务的最新执行历史
	GetLatestHistory(ctx context.Context, cronId int64) (domain.JobHistory, error)
	// GetStatistics 获取任务的执行统计信息
	GetStatistics(ctx context.Context, cronId int64, days int) (map[string]interface{}, error)
	// DeleteHistory 删除单条执行历史
	DeleteHistory(ctx context.Context, id int64) error
	// DeleteHistoryByCronId 删除任务的所有执行历史
	DeleteHistoryByCronId(ctx context.Context, cronId int64) error
	// CleanupOldHistory 清理指定天数之前的历史记录
	CleanupOldHistory(ctx context.Context, days int) error
}

type jobHistoryService struct {
	repo repository.JobHistoryRepository
}

// NewJobHistoryService 创建JobHistoryService实例
func NewJobHistoryService(repo repository.JobHistoryRepository) JobHistoryService {
	return &jobHistoryService{repo: repo}
}

func (s *jobHistoryService) RecordHistory(ctx context.Context, history domain.JobHistory) error {
	return s.repo.Create(ctx, history)
}

func (s *jobHistoryService) GetHistory(ctx context.Context, id int64) (domain.JobHistory, error) {
	return s.repo.FindById(ctx, id)
}

func (s *jobHistoryService) GetHistoryList(ctx context.Context, cronId int64, page, pageSize int) ([]domain.JobHistory, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	histories, err := s.repo.FindByCronId(ctx, cronId, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountByCronId(ctx, cronId)
	if err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (s *jobHistoryService) GetHistoryListByStatus(ctx context.Context, status domain.ExecutionStatus, page, pageSize int) ([]domain.JobHistory, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	histories, err := s.repo.FindByStatus(ctx, status, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountByStatus(ctx, status)
	if err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (s *jobHistoryService) GetHistoryListByTimeRange(ctx context.Context, startTime, endTime int64, page, pageSize int) ([]domain.JobHistory, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	return s.repo.FindByTimeRange(ctx, startTime, endTime, pageSize, offset)
}

func (s *jobHistoryService) GetLatestHistory(ctx context.Context, cronId int64) (domain.JobHistory, error) {
	return s.repo.GetLatestByCronId(ctx, cronId)
}

func (s *jobHistoryService) GetStatistics(ctx context.Context, cronId int64, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 7 // 默认最近7天
	}
	return s.repo.GetStatistics(ctx, cronId, days)
}

func (s *jobHistoryService) DeleteHistory(ctx context.Context, id int64) error {
	return s.repo.DeleteById(ctx, id)
}

func (s *jobHistoryService) DeleteHistoryByCronId(ctx context.Context, cronId int64) error {
	return s.repo.DeleteByCronId(ctx, cronId)
}

func (s *jobHistoryService) CleanupOldHistory(ctx context.Context, days int) error {
	if days <= 0 {
		days = 30 // 默认清理30天前的记录
	}

	// 计算时间戳（当前时间减去指定天数）
	beforeTime := time.Now().AddDate(0, 0, -days).Unix()

	return s.repo.DeleteBeforeTime(ctx, beforeTime)
}
