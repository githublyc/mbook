package grpcx

import (
	"context"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"mbook/webook/pkg/logger"
	"mbook/webook/pkg/netx"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	Addr     string
	EtcdAddr string
	Port     int
	Name     string
	L        logger.LoggerV1

	client   *etcdv3.Client
	kaCancel func()
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	s.register()
	return s.Server.Serve(listen)
}
func (s *Server) register() error {
	client, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.client = client
	em, err := endpoints.NewManager(client, "service/"+s.Name)

	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + addr

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期
	var ttl int64 = 5
	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连你
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	// 续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := s.client.KeepAlive(kaCtx, leaseResp.ID)
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return err
}
func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.client != nil {
		// 依赖注入，就不要关，可能被别人用
		// 我是自己初始化的，当然由我来关
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}
