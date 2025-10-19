package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "mbook/webook/interactive/grpc"
	"mbook/webook/pkg/grpcx"
	"mbook/webook/pkg/logger"
)

// intrSvc起名有争议，为了明确表示它是InteractiveServiceServer类型，我改成intrSvcServer
func InitGprcxServer(intrSvcServer *grpc2.InteractiveServiceServer,
	l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		EtcdAddr string `yaml:"etcdAddr"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
	}
	s := grpc.NewServer()
	//intrv1.RegisterInteractiveServiceServer(s, intrSvcServer)
	intrSvcServer.Register(s)
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	return &grpcx.Server{
		Server:   s,
		EtcdAddr: cfg.EtcdAddr,
		Port:     cfg.Port,
		Name:     cfg.Name,
		L:        l,
	}
}
