package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "mbook/webook/follow/grpc"
	"mbook/webook/pkg/grpcx"
)

func InitGRPCxServer(followRelation *grpc2.FollowServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	followRelation.Register(server)
	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
