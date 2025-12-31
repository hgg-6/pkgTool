package executor

import (
	"context"

	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
)

// ExecutionResult 任务执行结果
type ExecutionResult struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	StartTime int64       `json:"start_time"`
	EndTime   int64       `json:"end_time"`
	Duration  int64       `json:"duration"` // 毫秒
}

// Executor 任务执行器接口
type Executor interface {
	// Execute 执行任务
	Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error)

	// Type 返回执行器类型
	Type() domain.TaskType
}

// ExecutorFactory 执行器工厂
type ExecutorFactory interface {
	// GetExecutor 根据任务类型获取执行器
	GetExecutor(taskType domain.TaskType) (Executor, error)
}
