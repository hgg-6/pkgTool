package slidingWindow

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/rpc/grpcx/limiter/slidingWindow/testPkg"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
	"testing"
	"time"
)

func TestNewSlidingWindowLimiter(t *testing.T) {
	// 创建滑动窗口限流器, 最多2/s个请求，多了就触发限流
	limit := NewSlidingWindowLimiter(time.Second, 2)
	// 创建grpc服务，注册限流拦截器
	gs := grpc.NewServer(
		grpc.UnaryInterceptor(limit.BuildServerInterceptor()),
	)

	us := &testPkg.Service{}                  // 实例化一个服务
	testPkg.RegisterUserServiceServer(gs, us) // 注册服务

	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)

	err = gs.Serve(l)
	require.NoError(t, err)

	t.Log("服务启动成功")
}

// GRPC客户端发起调用
func TestClient(t *testing.T) {
	// insecure.NewCredentials是创建一个不安全的凭证，不启用https
	cl, err := grpc.NewClient( // 创建 grpc 客户端
		"127.0.0.1:8090", // 本机测试域名localhost有点慢。
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second, // 最小连接超时（默认20秒）
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Second, // 发送 keepalive 探测的时间间隔
			Timeout:             5 * time.Second, // 等待确认的超时时间
			PermitWithoutStream: true,            // 即使没有活跃流也发送 keepalive
		}),
	)
	require.NoError(t, err)
	defer cl.Close()

	ucClient := testPkg.NewUserServiceClient(cl)

	log.Println("开始调用服务")

	// 并发调用服务
	var eg errgroup.Group
	// 模拟3个客户端并发调用服务
	for i := 0; i < 3; i++ {
		eg.Go(func() error {
			res, er := ucClient.GetById(context.Background(), &testPkg.GetByIdRequest{Id: 123})
			if er != nil {
				return er
			}
			log.Println("resp.user: ", res.User)
			return nil
		})
	}
	err = eg.Wait()
	t.Log(err)

}
