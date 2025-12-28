package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/syncX/lock/redisLock/redsyncx/lock_cron_mysql/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCTaskConfig gRPC任务配置
type GRPCTaskConfig struct {
	Target      string            `json:"target"`       // gRPC服务地址，如 "localhost:50051"
	Service     string            `json:"service"`      // 服务名，如 "helloworld.Greeter"
	Method      string            `json:"method"`       // 方法名，如 "SayHello"
	RequestData string            `json:"request_data"` // 请求数据（JSON格式）
	Metadata    map[string]string `json:"metadata"`     // gRPC元数据
}

// GRPCExecutor gRPC任务执行器
type GRPCExecutor struct {
	l logx.Loggerx
}

// NewGRPCExecutor 创建gRPC执行器
func NewGRPCExecutor(l logx.Loggerx) *GRPCExecutor {
	return &GRPCExecutor{l: l}
}

// Execute 执行gRPC任务
func (g *GRPCExecutor) Execute(ctx context.Context, job domain.CronJob) (*ExecutionResult, error) {
	startTime := time.Now()

	// 解析gRPC任务配置
	var config GRPCTaskConfig
	if err := json.Unmarshal([]byte(job.Description), &config); err != nil {
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("解析gRPC任务配置失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, err
	}

	// 验证配置
	if config.Target == "" {
		return &ExecutionResult{
			Success:   false,
			Message:   "gRPC任务Target不能为空",
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, fmt.Errorf("target is required")
	}

	if config.Service == "" || config.Method == "" {
		return &ExecutionResult{
			Success:   false,
			Message:   "gRPC任务Service和Method不能为空",
			StartTime: startTime.Unix(),
			EndTime:   time.Now().Unix(),
		}, fmt.Errorf("service and method are required")
	}

	g.l.Info("执行gRPC任务",
		logx.Int64("job_id", job.CronId),
		logx.String("job_name", job.Name),
		logx.String("target", config.Target),
		logx.String("service", config.Service),
		logx.String("method", config.Method),
	)

	// 创建gRPC连接
	conn, err := grpc.DialContext(
		ctx,
		config.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		endTime := time.Now()
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("连接gRPC服务失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  endTime.Sub(startTime).Milliseconds(),
		}, err
	}
	defer conn.Close()

	// 注意：这里是简化实现
	// 实际生产环境需要使用反射或代码生成来动态调用gRPC方法
	// 这里仅作为占位符实现

	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	g.l.Info("gRPC任务执行完成（占位符实现）",
		logx.Int64("job_id", job.CronId),
		logx.Int64("duration_ms", duration),
	)

	return &ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("gRPC任务执行成功（占位符实现）: %s/%s", config.Service, config.Method),
		Data: map[string]interface{}{
			"target":  config.Target,
			"service": config.Service,
			"method":  config.Method,
			"note":    "这是占位符实现，实际生产需要使用gRPC反射或代码生成",
		},
		StartTime: startTime.Unix(),
		EndTime:   endTime.Unix(),
		Duration:  duration,
	}, nil
}

// Type 返回执行器类型
func (g *GRPCExecutor) Type() domain.TaskType {
	return domain.TaskTypeGRPC
}

// 注意：完整的gRPC动态调用实现需要：
// 1. 使用 grpc-go 的反射包
// 2. 或使用 protoreflect 包解析 proto 文件
// 3. 或使用代码生成预先生成所有可能的调用代码
//
// 示例完整实现（需要额外依赖）：
// import "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
// 然后使用 ServerReflectionClient 查询服务信息并动态调用
