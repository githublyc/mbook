//go:build wireinject

package main

import (
	"github.com/google/wire"
	grpc2 "mbook/webook/follow/grpc"
	"mbook/webook/follow/ioc"
	"mbook/webook/follow/repository"
	"mbook/webook/follow/repository/dao"
	"mbook/webook/follow/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMFollowRelationDAO,
	repository.NewFollowRelationRepository,
	service.NewFollowRelationService,
	grpc2.NewFollowRelationServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
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
