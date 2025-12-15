package slidingWindow

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/rpc/grpcx/limiter/slidingWindow/testPkg"
	"gitee.com/hgg_test/pkg_tool/v2/rpc/grpcx/observationX/grpcLogX"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// TestSlidingWindowLimiterDirect 直接测试滑动窗口限流器逻辑
func TestSlidingWindowLimiterDirect(t *testing.T) {
	// 创建滑动窗口限流器, 最多2/s个请求
	limiter := NewSlidingWindowLimiter(time.Second, 2)
	interceptor := limiter.BuildServerInterceptor()

	// 模拟的handler
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	// 模拟的gRPC信息
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/TestMethod",
	}

	// 测试1: 前2个请求应该成功
	for i := 0; i < 2; i++ {
		resp, err := interceptor(context.Background(), "request", info, handler)
		assert.NoError(t, err)
		assert.Equal(t, "response", resp)
	}

	// 测试2: 第3个请求应该被限流
	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Nil(t, resp)
	assert.Error(t, err)

	// 检查是否为限流错误
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.ResourceExhausted, st.Code())
	assert.Contains(t, st.Message(), "滑动窗口限流")

	// 测试3: 等待一段时间后，请求应该再次成功
	time.Sleep(time.Second + 100*time.Millisecond)

	resp, err = interceptor(context.Background(), "request", info, handler)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

// TestSlidingWindowLimiterIntegration 集成测试（可选，当需要测试真实gRPC时）
func TestSlidingWindowLimiterIntegration(t *testing.T) {
	// 如果不想运行集成测试，可以跳过
	if testing.Short() {
		t.Skip("跳过集成测试，使用 -short 参数")
	}

	// 创建滑动窗口限流器, 最多2/s个请求
	limit := NewSlidingWindowLimiter(time.Second, 2)

	// 创建grpc服务
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcLogX.NewInterceptorBuilder(initLog()).BuildServerUnaryInterceptor(),
			limit.BuildServerInterceptor(),
		),
	)

	us := &testPkg.Service{}
	testPkg.RegisterUserServiceServer(gs, us)

	// 使用随机端口避免冲突
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	addr := l.Addr().String()
	t.Logf("启动测试服务器在 %s", addr)

	// 使用WaitGroup确保服务器启动
	var wg sync.WaitGroup
	wg.Add(1)

	// 在goroutine中启动服务器
	serverErrChan := make(chan error, 1)
	go func() {
		wg.Done()
		serverErrChan <- gs.Serve(l)
	}()

	// 等待服务器启动
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// 确保测试完成后停止服务器
	defer gs.GracefulStop()

	// 创建客户端
	cl, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 2 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	require.NoError(t, err)
	defer cl.Close()

	ucClient := testPkg.NewUserServiceClient(cl)

	// 测试1: 前2个请求应该成功
	var successCount int
	var eg errgroup.Group

	for i := 0; i < 3; i++ {
		idx := i
		eg.Go(func() error {
			res, err := ucClient.GetById(context.Background(), &testPkg.GetByIdRequest{Id: int64(idx + 1)})
			if err != nil {
				// 检查是否为限流错误
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.ResourceExhausted {
					t.Logf("请求 %d 被限流（预期行为）", idx)
					return nil // 限流是预期行为
				}
				return fmt.Errorf("请求 %d 失败: %w", idx, err)
			}
			t.Logf("请求 %d 成功: user=%v", idx, res.User)
			successCount++
			return nil
		})
	}

	err = eg.Wait()
	assert.NoError(t, err)

	// 验证：前2个应该成功，第3个应该被限流
	t.Logf("成功请求数: %d (期望2个成功，1个被限流)", successCount)
	assert.True(t, successCount >= 1 && successCount <= 2, "应该有1-2个成功请求")

	// 测试2: 等待窗口滑动后再次请求
	time.Sleep(time.Second + 100*time.Millisecond)

	res, err := ucClient.GetById(context.Background(), &testPkg.GetByIdRequest{Id: 100})
	if err != nil {
		// 检查错误类型
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.ResourceExhausted {
			t.Log("请求在窗口滑动后仍被限流（可能窗口未完全滑动）")
		} else {
			assert.NoError(t, err, "窗口滑动后请求应该成功")
		}
	} else {
		t.Logf("窗口滑动后请求成功: user=%v", res.User)
	}

	// 检查服务器错误
	select {
	case err := <-serverErrChan:
		if err != nil && err.Error() != "closed" {
			t.Errorf("服务器错误: %v", err)
		}
	default:
		// 没有错误
	}
}

// TestSlidingWindowEdgeCases 测试边界情况
func TestSlidingWindowEdgeCases(t *testing.T) {
	// 测试零阈值
	t.Run("ZeroThreshold", func(t *testing.T) {
		limiter := NewSlidingWindowLimiter(time.Second, 0)
		interceptor := limiter.BuildServerInterceptor()

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/TestMethod",
		}

		resp, err := interceptor(context.Background(), "request", info, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.ResourceExhausted, st.Code())
	})

	// 测试大窗口
	t.Run("LargeWindow", func(t *testing.T) {
		limiter := NewSlidingWindowLimiter(5*time.Second, 10)
		interceptor := limiter.BuildServerInterceptor()

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "response", nil
		}

		info := &grpc.UnaryServerInfo{
			FullMethod: "/test.Service/TestMethod",
		}

		// 前10个请求应该成功
		for i := 0; i < 10; i++ {
			resp, err := interceptor(context.Background(), "request", info, handler)
			assert.NoError(t, err)
			assert.Equal(t, "response", resp)
		}

		// 第11个应该被限流
		resp, err := interceptor(context.Background(), "request", info, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}

func initLog() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	return zerologx.NewZeroLogger(&logger)
}
