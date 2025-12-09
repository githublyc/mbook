//go:build wireinject

package main

import (
	"github.com/google/wire"
	"mbook/webook/feed/events"
	"mbook/webook/feed/grpc"
	"mbook/webook/feed/ioc"
	"mbook/webook/feed/repository"
	"mbook/webook/feed/repository/cache"
	"mbook/webook/feed/repository/dao"
	"mbook/webook/feed/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewFeedPushEventDAO,
	dao.NewFeedPullEventDAO,
	cache.NewFeedEventCache,
	repository.NewFeedEventRepo,
)

var thirdProvider = wire.NewSet(
	ioc.InitEtcdClient,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitDB,
	ioc.InitFollowClient,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.RegisterHandler,
		service.NewFeedService,
		grpc.NewFeedEventGrpcSvc,
		events.NewArticleEventConsumer,
		events.NewFeedEventConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
