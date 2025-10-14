package etcdx

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

/*
	grpc中接入etcd
*/

type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (e *EtcdTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	// etcdv3.NewFromURLs()
	// etcdv3.New(etcdv3.Config{})
	//assert.NoError(e.T(), err)  // 仅标记测试失败，继续执行后续代码
	require.NoError(e.T(), err) // 立即终止当前测试函数（使用 return）
	e.cli = cli
}

// 注册中心、注册的实例，添加一个节点
func (e *EtcdTestSuite) TestServer() {
	t := e.T()
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	em, err := endpoints.NewManager(e.cli, "service/user") // 注册中心的信息
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr
	// 添加一个服务
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连接你
		Addr: addr,
	})
	require.NoError(t, err)

	go func() {
		// 模拟注册信息变更
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C { // ticker.C 是一个通道，每秒发送一个时间
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			// 更新注册中心
			err1 := em.Update(ctx1, []*endpoints.UpdateWithOpts{
				{
					Update: endpoints.Update{
						Op:  endpoints.Add,
						Key: key,
						Endpoint: endpoints.Endpoint{
							Addr:     addr,
							Metadata: now.String(),
						},
					},
				},
				// Update的话，两条【key】不能相同
				//{
				//	Update: endpoints.Update{
				//		Op:  endpoints.Add,
				//		Key: key1,
				//		Endpoint: endpoints.Endpoint{
				//			Addr:     addr,
				//			Metadata: now.String(),
				//		},
				//	},
				//},
			})
			// 更新注册中心，AddEndpoint简化写法。INSERT OR UPDATE, SAVE 的语义
			//err1 := em.AddEndpoint(ctx, key, endpoints.Endpoint{
			//	Addr:     addr,
			//	Metadata: now.String(),
			//})
			cancel1()
			if err1 != nil {
				t.Log(err1)
			}
		}
	}()

	// grpc服务端
	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Service{})
	// 启动服务
	err = server.Serve(l)
	require.NoError(t, err)

	// 删除一个服务
	err = em.DeleteEndpoint(ctx, key)
	require.NoError(t, err)

	server.GracefulStop() // gRPC优雅关闭退出
	err = e.cli.Close()   // 关闭etcd客户端
	require.NoError(t, err)
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
