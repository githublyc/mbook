//go:build wireinject

package main

import (
	"github.com/google/wire"
	grpc2 "mbook/webook/comment/grpc"
	"mbook/webook/comment/ioc"
	"mbook/webook/comment/repository"
	"mbook/webook/comment/repository/dao"
	"mbook/webook/comment/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepo,
	service.NewCommentSvc,
	grpc2.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
