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
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
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

	// 使用gRPC反射动态调用服务方法
	reflectionClient := grpc_reflection_v1alpha.NewServerReflectionClient(conn)

	// 查询服务描述符
	fileDescriptors, err := getFileDescriptors(ctx, reflectionClient, config.Service)
	if err != nil {
		endTime := time.Now()
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("获取服务描述符失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  endTime.Sub(startTime).Milliseconds(),
		}, err
	}

	// 查找方法描述符
	_, err = findMethodDescriptor(fileDescriptors, config.Service, config.Method)
	if err != nil {
		endTime := time.Now()
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("查找方法描述符失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  endTime.Sub(startTime).Milliseconds(),
		}, err
	}

	// 解析请求数据
	request := make(map[string]interface{})
	if config.RequestData != "" {
		if err := json.Unmarshal([]byte(config.RequestData), &request); err != nil {
			endTime := time.Now()
			return &ExecutionResult{
				Success:   false,
				Message:   fmt.Sprintf("解析请求数据失败: %v", err),
				StartTime: startTime.Unix(),
				EndTime:   endTime.Unix(),
				Duration:  endTime.Sub(startTime).Milliseconds(),
			}, err
		}
	}

	// 调用gRPC方法
	// 注意：这里使用conn.Invoke进行动态调用，不需要具体的消息类型
	// 我们需要将JSON请求转换为protobuf消息
	// 由于我们没有具体的消息类型，这里暂时使用map作为请求和响应
	// 实际生产环境中应该使用反射创建具体的protobuf消息
	var response interface{}
	err = conn.Invoke(ctx, "/"+config.Service+"/"+config.Method, request, &response)
	if err != nil {
		endTime := time.Now()
		return &ExecutionResult{
			Success:   false,
			Message:   fmt.Sprintf("调用gRPC方法失败: %v", err),
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
			Duration:  endTime.Sub(startTime).Milliseconds(),
		}, err
	}

	g.l.Info("gRPC任务调用成功",
		logx.String("service", config.Service),
		logx.String("method", config.Method),
	)

	// 转换响应结果
	responseJSON, err := json.Marshal(response)
	if err != nil {
		g.l.Warn("转换响应结果失败", logx.Error(err))
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Milliseconds()

	g.l.Info("gRPC任务执行完成",
		logx.Int64("job_id", job.CronId),
		logx.Int64("duration_ms", duration),
	)

	return &ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("gRPC任务执行成功: %s/%s", config.Service, config.Method),
		Data: map[string]interface{}{
			"target":   config.Target,
			"service":  config.Service,
			"method":   config.Method,
			"response": string(responseJSON),
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

// getFileDescriptors 获取服务对应的文件描述符
func getFileDescriptors(ctx context.Context, client grpc_reflection_v1alpha.ServerReflectionClient, serviceName string) ([]*descriptorpb.FileDescriptorProto, error) {
	// 查询服务的文件描述符
	req := &grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: serviceName,
		},
	}

	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}

	if err := stream.Send(req); err != nil {
		return nil, err
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	// 处理响应
	if fileDescResp := resp.GetFileDescriptorResponse(); fileDescResp != nil {
		var fileDescriptors []*descriptorpb.FileDescriptorProto
		for _, fileDescBytes := range fileDescResp.FileDescriptorProto {
			fileDesc := &descriptorpb.FileDescriptorProto{}
			if err := proto.Unmarshal(fileDescBytes, fileDesc); err != nil {
				return nil, err
			}
			fileDescriptors = append(fileDescriptors, fileDesc)
		}
		return fileDescriptors, nil
	}

	return nil, fmt.Errorf("未找到服务 %s 的文件描述符", serviceName)
}

// findMethodDescriptor 查找方法描述符
func findMethodDescriptor(fileDescriptors []*descriptorpb.FileDescriptorProto, serviceName, methodName string) (*descriptorpb.MethodDescriptorProto, error) {
	// 遍历所有文件描述符
	for _, fileDesc := range fileDescriptors {
		// 查找服务
		for _, service := range fileDesc.Service {
			if service.GetName() == serviceName {
				// 查找方法
				for _, method := range service.Method {
					if method.GetName() == methodName {
						return method, nil
					}
				}
				return nil, fmt.Errorf("服务 %s 中未找到方法 %s", serviceName, methodName)
			}
		}
	}

	return nil, fmt.Errorf("未找到服务 %s", serviceName)
}
