package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hgg-6/pkgTool/v2/logx"
	"github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"github.com/hgg-6/pkgTool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/service"
)

var (
	ErrExecutorNotFound   = errors.New("executor not found for task type")
	ErrExecutionTimeout   = errors.New("task execution timeout")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)

// DefaultExecutorFactory 默认执行器工厂
type DefaultExecutorFactory struct {
	executors      map[domain.TaskType]Executor
	historyService service.JobHistoryService
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
	if f.historyService != nil {
		executor = NewRetryableHistoryExecutor(executor, f.historyService, f.l)
	} else {
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

// ValidateTask 校验任务配置
func (f *DefaultExecutorFactory) ValidateTask(ctx context.Context, job domain.CronJob) error {
	exec, err := f.GetExecutor(job.TaskType)
	if err != nil {
		return fmt.Errorf("不支持的任务类型: %s", job.TaskType)
	}
	return exec.Validate(ctx, job)
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
			timeout = 30 * time.Second
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
		} else if result != nil && !result.Success {
			r.l.Warn("任务返回失败结果",
				logx.Int64("job_id", job.CronId),
				logx.String("job_name", job.Name),
				logx.Int("attempt", attempt+1),
				logx.String("message", result.Message),
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

// Validate 代理到内部执行器的校验
func (r *RetryableExecutor) Validate(ctx context.Context, job domain.CronJob) error {
	return r.executor.Validate(ctx, job)
}

// FunctionExecutor Function类型任务执行器
type FunctionExecutor struct {
	l         logx.Loggerx
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

// HasFunction 检查函数是否已注册
func (f *FunctionExecutor) HasFunction(name string) bool {
	_, exists := f.functions[name]
	return exists
}

// ListFunctions 列出所有已注册的函数名
func (f *FunctionExecutor) ListFunctions() []string {
	names := make([]string, 0, len(f.functions))
	for name := range f.functions {
		names = append(names, name)
	}
	return names
}

// Validate 校验Function任务配置
func (f *FunctionExecutor) Validate(ctx context.Context, job domain.CronJob) error {
	var config struct {
		FunctionName string                 `json:"function_name"`
		Parameters   map[string]interface{} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(job.Description), &config); err != nil {
		return fmt.Errorf("解析Function任务配置失败: %v", err)
	}
	if config.FunctionName == "" {
		return fmt.Errorf("function_name 不能为空")
	}
	if !f.HasFunction(config.FunctionName) {
		return fmt.Errorf("未找到注册的函数: %s，当前已注册: %v", config.FunctionName, f.ListFunctions())
	}
	return nil
}

// Execute 执行Function任务
func (f *FunctionExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	startTime := time.Now()

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

	if config.FunctionName == "" {
		return &ExecutionResult{
			Success:   false,
			Message:   "Function任务function_name不能为空",
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, fmt.Errorf("function_name is required")
	}

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
