package main

import (
	"github.com/google/wire"
	"mbook/webook/pkg/wego"
	"mbook/webook/tag/grpc"
	"mbook/webook/tag/ioc"
	"mbook/webook/tag/repository/cache"
	"mbook/webook/tag/repository/dao"
	"mbook/webook/tag/service"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitDB,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		cache.NewRedisTagCache,
		dao.NewGORMTagDAO,
		ioc.InitRepository,
		service.NewTagService,
		grpc.NewTagServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
