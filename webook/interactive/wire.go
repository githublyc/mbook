//go:build wireinject

package main

import (
	"github.com/google/wire"
	"mbook/webook/interactive/events"
	"mbook/webook/interactive/grpc"
	"mbook/webook/interactive/ioc"
	repository2 "mbook/webook/interactive/repository"
	cache2 "mbook/webook/interactive/repository/cache"
	dao2 "mbook/webook/interactive/repository/dao"
	service2 "mbook/webook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	//第三方依赖
	ioc.InitRedis, ioc.InitDB,
	ioc.InitLogger,
	ioc.InitSaramaClient,
)
var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(thirdPartySet,
		interactiveSvcSet,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		grpc.NewInteractiveServiceServer,
		ioc.InitGprcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
