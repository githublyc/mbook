package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	grpc2 "mbook/webook/payment/grpc"
	"mbook/webook/pkg/grpcx"
	ilogger "mbook/webook/pkg/grpcx/interceptor/logger"
	"mbook/webook/pkg/logger"
)

func InitGRPCServer(wesvc *grpc2.WechatServiceServer,
	ecli *clientv3.Client,
	l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port    int   `yaml:"port"`
		EtcdTTL int64 `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		ilogger.NewInterceptorBuilder(l).BuildServerUnaryInterceptor(),
	))
	wesvc.Register(server)
	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       "payment",
		L:          l,
		EtcdTTL:    cfg.EtcdTTL,
		EtcdClient: ecli,
	}
}
