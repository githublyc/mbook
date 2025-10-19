package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	// etcdv3.NewFromURLs()
	// etcdv3.New(etcdv3.Config{Endpoints: })
	require.NoError(s.T(), err)
	s.cli = cli
}
func (s *EtcdTestSuite) TestClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)
	cc, err := grpc.NewClient("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}

// 服务注册
func (s *EtcdTestSuite) TestServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)

	addr := "127.0.0.1:8090"
	key := "service/user" + addr
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)

	// 租期
	var ttl int64 = 5
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连你
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	// 续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err1)
		for kaResp := range ch {
			t.Log(kaResp.String())
		}
	}()

	go func() {
		// 模拟注册信息变动
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
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
					Opts: []etcdv3.OpOption{
						etcdv3.WithLease(leaseResp.ID),
					},
				},
			})

			// INSERT or update, save的语义
			//err1 := em.AddEndpoint(ctx1, key, endpoints.Endpoint{
			//	Addr:     addr,
			//	Metadata: now.String(),
			//}, etcdv3.WithLease(leaseResp.ID))
			cancel1()
			if err1 != nil {
				t.Log(err1)
			}
		}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	//先注册，再启动。因为Serve是阻塞的方法
	server.Serve(l)

	kaCancel()
	//删除注册信息
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	server.GracefulStop()
	s.cli.Close()
}
func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
