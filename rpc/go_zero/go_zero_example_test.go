package go_zero

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"testing"
	"time"
)

type GoZeroTestSuite struct {
	suite.Suite
}

// TestGoZeroClient 启动 grpc 客户端
func (s *GoZeroTestSuite) TestGoZeroClient() {
	zClient := zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:12379"},
			Key:   "user",
		},
	},
		zrpc.WithDialOption(
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
		))
	client := NewUserServiceClient(zClient.Conn())
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

// TestGoZeroServer 启动 grpc 服务端
func (s *GoZeroTestSuite) TestGoZeroServer() {
	go func() {
		s.startServer(":8090")
	}()
	s.startServer(":8091")
}

func (s *GoZeroTestSuite) startServer(addr string) {
	// 正常来说，这个都是从配置文件中读取的
	//var c zrpc.RpcServerConf
	// 类似与 main 函数那样，从命令行接收配置文件的路径
	//conf.MustLoad(*configFile, &c)
	c := zrpc.RpcServerConf{
		ListenOn: addr,
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:12379"},
			Key:   "user",
		},
	}
	// 创建一个服务器，并且注册服务实例
	server := zrpc.MustNewServer(c, func(grpcServer *grpc.Server) {
		RegisterUserServiceServer(grpcServer, &Service{
			Name: addr,
		})
	})

	// 这个是往 gRPC 里面增加拦截器（也可以叫做插件）
	// server.AddUnaryInterceptors(interceptor)
	server.Start()
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}
