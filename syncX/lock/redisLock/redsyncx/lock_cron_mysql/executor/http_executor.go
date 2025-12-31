package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
)

// HTTPTaskConfig HTTP任务配置（存储在CronJob的额外字段中）
type HTTPTaskConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // GET, POST, PUT, DELETE
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// HTTPExecutor HTTP任务执行器
type HTTPExecutor struct {
	client *http.Client
	l      logx.Loggerx
}

// NewHTTPExecutor 创建HTTP执行器
func NewHTTPExecutor(l logx.Loggerx) *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: 30 * time.Second, // 默认超时30秒
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		l: l,
	}
}

// Execute 执行HTTP任务
func (h *HTTPExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	startTime := time.Now()

	// 解析HTTP任务配置
	var config HTTPTaskConfig
	if err := json.Unmarshal([]byte(job.Description), &config); err != nil {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("解析HTTP任务配置失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, err
	}

	// 验证配置
	if config.URL == "" {
		return &ExecutionResult{
			Success:   false,
			Message:   "HTTP任务URL不能为空",
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, fmt.Errorf("URL is required")
	}

	if config.Method == "" {
		config.Method = "GET"
	}

	h.l.Info("执行HTTP任务",
		logx.Int64("job_id", job.CronId),
		logx.String("job_name", job.Name),
		logx.String("url", config.URL),
		logx.String("method", config.Method),
	)

	// 创建HTTP请求
	var reqBody io.Reader
	if config.Body != "" {
		reqBody = bytes.NewBufferString(config.Body)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, reqBody)
	if err != nil {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("创建HTTP请求失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, err
	}

	// 设置请求头
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 如果没有设置Content-Type且有Body，默认设置为application/json
	if config.Body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 执行HTTP请求
	resp, err := h.client.Do(req)
	if err != nil {
		endTime := time.Now()
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("HTTP请求失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  endTime.Sub(startTime).Milliseconds(),
		}, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		endTime := time.Now()
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("读取响应失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  endTime.Sub(startTime).Milliseconds(),
		}, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	// 判断是否成功（2xx状态码）
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	result := &ExecutionResult{
		Success: success,
		Message: fmt.Sprintf("HTTP %s %s - 状态码: %d", config.Method, config.URL, resp.StatusCode),
		Data: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
			"body":        string(respBody),
		},
		StartTime: startTime.Unix(),
		EndTime:   endTime.Unix(),
		Duration:  duration,
	}

	if !success {
		result.Message = fmt.Sprintf("%s - 响应: %s", result.Message, string(respBody))
		h.l.Warn("HTTP任务返回非成功状态码",
			logx.Int64("job_id", job.CronId),
			logx.Int("status_code", resp.StatusCode),
			logx.String("response", string(respBody)),
		)
	} else {
		h.l.Info("HTTP任务执行成功",
			logx.Int64("job_id", job.CronId),
			logx.Int("status_code", resp.StatusCode),
			logx.Int64("duration_ms", duration),
		)
	}

	return result, nil
}

// Type 返回执行器类型
func (h *HTTPExecutor) Type() domain.TaskType {
	return domain.TaskTypeHTTP
}
