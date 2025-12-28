package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
)

// HistoryRecordingExecutor 带历史记录功能的执行器包装器
type HistoryRecordingExecutor struct {
	executor       Executor
	historyService service.JobHistoryService
	l              logx.Loggerx
}

// NewHistoryRecordingExecutor 创建带历史记录的执行器
func NewHistoryRecordingExecutor(executor Executor, historyService service.JobHistoryService, l logx.Loggerx) *HistoryRecordingExecutor {
	return &HistoryRecordingExecutor{
		executor:       executor,
		historyService: historyService,
		l:              l,
	}
}

// Execute 执行任务并记录历史
func (h *HistoryRecordingExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	startTime := time.Now()

	// 执行任务
	result, err := h.executor.Execute(ctx, job)

	// 记录执行历史
	h.recordHistory(ctx, job, result, err, startTime)

	return result, err
}

// Type 返回执行器类型
func (h *HistoryRecordingExecutor) Type() domain.TaskType {
	return h.executor.Type()
}

// recordHistory 记录执行历史
func (h *HistoryRecordingExecutor) recordHistory(ctx context.Context, job domain.CronJob, result *ExecutionResult, execErr error, startTime time.Time) {
	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	// 确定执行状态
	var status domain.ExecutionStatus
	var errorMessage string
	var resultStr string

	if result != nil {
		// 转换result为JSON字符串
		if resultData, err := json.Marshal(result.Data); err == nil {
			resultStr = string(resultData)
		}

		if result.Success {
			status = domain.ExecutionStatusSuccess
		} else {
			// 检查是否超时
			if execErr == context.DeadlineExceeded {
				status = domain.ExecutionStatusTimeout
			} else {
				status = domain.ExecutionStatusFailure
			}
			errorMessage = result.Message
		}
	} else {
		// result为nil，说明执行出错
		if execErr == context.DeadlineExceeded {
			status = domain.ExecutionStatusTimeout
			errorMessage = "任务执行超时"
		} else {
			status = domain.ExecutionStatusFailure
			if execErr != nil {
				errorMessage = execErr.Error()
			} else {
				errorMessage = "任务执行失败，未知错误"
			}
		}
	}

	// 创建历史记录
	history := domain.JobHistory{
		CronId:       job.CronId,
		JobName:      job.Name,
		Status:       status,
		StartTime:    startTime.Unix(),
		EndTime:      endTime.Unix(),
		Duration:     duration,
		RetryCount:   0, // 在RetryableExecutor中会更新
		ErrorMessage: errorMessage,
		Result:       resultStr,
		Ctime:        float64(time.Now().Unix()),
	}

	// 异步保存历史记录，避免影响任务执行
	go func() {
		// 创建新的context，避免使用已取消的context
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.historyService.RecordHistory(saveCtx, history); err != nil {
			h.l.Error("保存任务执行历史失败",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Error(err),
			)
		} else {
			h.l.Debug("任务执行历史已记录",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.String("status", string(status)),
				logx.Int64("duration_ms", duration),
			)
		}
	}()
}

// RetryableHistoryExecutor 带重试和历史记录的执行器包装器
type RetryableHistoryExecutor struct {
	executor       Executor
	historyService service.JobHistoryService
	l              logx.Loggerx
}

// NewRetryableHistoryExecutor 创建带重试和历史记录的执行器
func NewRetryableHistoryExecutor(executor Executor, historyService service.JobHistoryService, l logx.Loggerx) *RetryableHistoryExecutor {
	return &RetryableHistoryExecutor{
		executor:       executor,
		historyService: historyService,
		l:              l,
	}
}

// Execute 执行任务（带重试和历史记录）
func (r *RetryableHistoryExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	maxRetry := job.MaxRetry
	if maxRetry <= 0 {
		maxRetry = 1
	}

	var lastErr error
	var lastResult *ExecutionResult
	overallStartTime := time.Now()

	for attempt := 0; attempt < maxRetry; attempt++ {
		attemptStartTime := time.Now()

		if attempt > 0 {
			r.l.Info("重试执行任务",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Int("attempt", attempt+1),
				logx.Int("max_retry", maxRetry),
			)

			// 记录重试状态的历史
			r.recordRetryHistory(ctx, job, attempt, lastResult, lastErr, attemptStartTime)

			// 重试前等待（指数退避）
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		// 创建带超时的context
		timeout := time.Duration(job.Timeout) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second // 默认30秒
		}

		executeCtx, cancel := context.WithTimeout(ctx, timeout)
		result, err := r.executor.Execute(executeCtx, job)
		cancel()

		if err == nil && result != nil && result.Success {
			// 执行成功，记录历史
			r.recordSuccessHistory(ctx, job, result, attempt, overallStartTime)
			return result, nil
		}

		lastErr = err
		lastResult = result

		if err != nil {
			r.l.Error("任务执行失败",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Int("attempt", attempt+1),
				logx.Error(err),
			)
		}
	}

	// 所有重试都失败，记录最终失败的历史
	r.recordFailureHistory(ctx, job, lastResult, lastErr, maxRetry, overallStartTime)

	// 所有重试都失败
	if lastResult != nil {
		lastResult.Success = false
		lastResult.Message = fmt.Sprintf("执行失败（重试%d次后）: %s", maxRetry, lastResult.Message)
		return lastResult, ErrMaxRetriesExceeded
	}

	return &ExecutionResult{
		Success: false,
		Message: fmt.Sprintf("执行失败（重试%d次后）: %v", maxRetry, lastErr),
	}, ErrMaxRetriesExceeded
}

// Type 返回执行器类型
func (r *RetryableHistoryExecutor) Type() domain.TaskType {
	return r.executor.Type()
}

// recordSuccessHistory 记录成功的历史
func (r *RetryableHistoryExecutor) recordSuccessHistory(ctx context.Context, job domain.CronJob, result *ExecutionResult, retryCount int, startTime time.Time) {
	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	var resultStr string
	if result != nil && result.Data != nil {
		if data, err := json.Marshal(result.Data); err == nil {
			resultStr = string(data)
		}
	}

	history := domain.JobHistory{
		CronId:       job.CronId,
		JobName:      job.Name,
		Status:       domain.ExecutionStatusSuccess,
		StartTime:    startTime.Unix(),
		EndTime:      endTime.Unix(),
		Duration:     duration,
		RetryCount:   retryCount,
		ErrorMessage: "",
		Result:       resultStr,
		Ctime:        float64(time.Now().Unix()),
	}

	r.saveHistory(ctx, history, job)
}

// recordFailureHistory 记录失败的历史
func (r *RetryableHistoryExecutor) recordFailureHistory(ctx context.Context, job domain.CronJob, result *ExecutionResult, execErr error, retryCount int, startTime time.Time) {
	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	var status domain.ExecutionStatus
	var errorMessage string

	if execErr == context.DeadlineExceeded {
		status = domain.ExecutionStatusTimeout
		errorMessage = "任务执行超时"
	} else {
		status = domain.ExecutionStatusFailure
		if result != nil {
			errorMessage = result.Message
		} else if execErr != nil {
			errorMessage = execErr.Error()
		} else {
			errorMessage = "未知错误"
		}
	}

	history := domain.JobHistory{
		CronId:       job.CronId,
		JobName:      job.Name,
		Status:       status,
		StartTime:    startTime.Unix(),
		EndTime:      endTime.Unix(),
		Duration:     duration,
		RetryCount:   retryCount,
		ErrorMessage: errorMessage,
		Result:       "",
		Ctime:        float64(time.Now().Unix()),
	}

	r.saveHistory(ctx, history, job)
}

// recordRetryHistory 记录重试中的历史
func (r *RetryableHistoryExecutor) recordRetryHistory(ctx context.Context, job domain.CronJob, retryCount int, result *ExecutionResult, execErr error, startTime time.Time) {
	var errorMessage string
	if result != nil {
		errorMessage = result.Message
	} else if execErr != nil {
		errorMessage = execErr.Error()
	}

	history := domain.JobHistory{
		CronId:       job.CronId,
		JobName:      job.Name,
		Status:       domain.ExecutionStatusRetrying,
		StartTime:    startTime.Unix(),
		EndTime:      time.Now().Unix(),
		Duration:     0,
		RetryCount:   retryCount,
		ErrorMessage: errorMessage,
		Result:       "",
		Ctime:        float64(time.Now().Unix()),
	}

	r.saveHistory(ctx, history, job)
}

// saveHistory 保存历史记录
func (r *RetryableHistoryExecutor) saveHistory(ctx context.Context, history domain.JobHistory, job domain.CronJob) {
	// 异步保存历史记录
	go func() {
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.historyService.RecordHistory(saveCtx, history); err != nil {
			r.l.Error("保存任务执行历史失败",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Error(err),
			)
		} else {
			r.l.Debug("任务执行历史已记录",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.String("status", string(history.Status)),
				logx.Int64("duration_ms", history.Duration),
				logx.Int("retry_count", history.RetryCount),
			)
		}
	}()
}
