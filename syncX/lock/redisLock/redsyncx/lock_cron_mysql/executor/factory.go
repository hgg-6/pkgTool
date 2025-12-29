package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
)

var (
	ErrExecutorNotFound   = errors.New("executor not found for task type")
	ErrExecutionTimeout   = errors.New("task execution timeout")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)

// DefaultExecutorFactory 默认执行器工厂
type DefaultExecutorFactory struct {
	executors      map[domain.TaskType]Executor
	historyService service.JobHistoryService // 添加历史服务
	l              logx.Loggerx
}

// NewExecutorFactory 创建执行器工厂
func NewExecutorFactory(l logx.Loggerx) ExecutorFactory {
	factory := &DefaultExecutorFactory{
		executors: make(map[domain.TaskType]Executor),
		l:         l,
	}
	return factory
}

// NewExecutorFactoryWithHistory 创建带历史记录的执行器工厂
func NewExecutorFactoryWithHistory(historyService service.JobHistoryService, l logx.Loggerx) ExecutorFactory {
	factory := &DefaultExecutorFactory{
		executors:      make(map[domain.TaskType]Executor),
		historyService: historyService,
		l:              l,
	}
	return factory
}

// RegisterExecutor 注册执行器
func (f *DefaultExecutorFactory) RegisterExecutor(executor Executor) {
	// 如果有历史服务，包装执行器以同时支持重试和历史记录
	if f.historyService != nil {
		executor = NewRetryableHistoryExecutor(executor, f.historyService, f.l)
	} else {
		// 没有历史服务时，至少包装重试逻辑
		executor = NewRetryableExecutor(executor, f.l)
	}
	f.executors[executor.Type()] = executor
	f.l.Info("注册任务执行器", logx.String("type", string(executor.Type())))
}

// GetExecutor 获取执行器
func (f *DefaultExecutorFactory) GetExecutor(taskType domain.TaskType) (Executor, error) {
	executor, ok := f.executors[taskType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrExecutorNotFound, taskType)
	}
	return executor, nil
}

// RetryableExecutor 带重试机制的执行器包装器
type RetryableExecutor struct {
	executor Executor
	l        logx.Loggerx
}

// NewRetryableExecutor 创建带重试的执行器
func NewRetryableExecutor(executor Executor, l logx.Loggerx) *RetryableExecutor {
	return &RetryableExecutor{
		executor: executor,
		l:        l,
	}
}

// Execute 执行任务（带重试）
func (r *RetryableExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	maxRetry := job.MaxRetry
	if maxRetry <= 0 {
		maxRetry = 1
	}

	var lastErr error
	var lastResult *ExecutionResult

	for attempt := 0; attempt < maxRetry; attempt++ {
		if attempt > 0 {
			r.l.Info("重试执行任务",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Int("attempt", attempt+1),
				logx.Int("max_retry", maxRetry),
			)

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
func (r *RetryableExecutor) Type() domain.TaskType {
	return r.executor.Type()
}

// FunctionExecutor Function类型任务执行器
type FunctionExecutor struct {
	l logx.Loggerx
	// 函数注册表：函数名 -> 函数实现
	functions map[string]func(context.Context, map[string]interface{}) (interface{}, error)
}

// NewFunctionExecutor 创建Function执行器
func NewFunctionExecutor(l logx.Loggerx) *FunctionExecutor {
	return &FunctionExecutor{
		l:         l,
		functions: make(map[string]func(context.Context, map[string]interface{}) (interface{}, error)),
	}
}

// RegisterFunction 注册函数
func (f *FunctionExecutor) RegisterFunction(name string, fn func(context.Context, map[string]interface{}) (interface{}, error)) {
	f.functions[name] = fn
	f.l.Info("注册Function任务函数", logx.String("name", name))
}

// Execute 执行Function任务
func (f *FunctionExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	startTime := time.Now()

	// 解析任务配置
	var config struct {
		FunctionName string                 `json:"function_name"`
		Parameters   map[string]interface{} `json:"parameters"`
	}

	if err := json.Unmarshal([]byte(job.Description), &config); err != nil {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("解析Function任务配置失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, err
	}

	// 验证配置
	if config.FunctionName == "" {
		return &ExecutionResult{
			Success:   false,
			Message:   "Function任务function_name不能为空",
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, fmt.Errorf("function_name is required")
	}

	// 查找函数
	fn, exists := f.functions[config.FunctionName]
	if !exists {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("未找到注册的函数: %s", config.FunctionName),
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, fmt.Errorf("function %s not found", config.FunctionName)
	}

	f.l.Info("执行Function任务",
		logx.Int64("job_id", job.CronId),
		logx.String("job_name", job.Name),
		logx.String("function_name", config.FunctionName),
	)

	// 执行函数
	result, err := fn(ctx, config.Parameters)
	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	if err != nil {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("Function执行失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  duration,
		}, err
	}

	// 构造执行结果
	resultMsg := fmt.Sprintf("Function执行成功: %v", result)
	return &ExecutionResult{
		Success: true,
		Message: resultMsg,
		Data: map[string]interface{}{
			"result": result,
		},
		StartTime: startTime.Unix(),
		EndTime:   endTime.Unix(),
		Duration:  duration,
	}, nil
}

// Type 返回执行器类型
func (f *FunctionExecutor) Type() domain.TaskType {
	return domain.TaskTypeFunction
}
