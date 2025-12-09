//go:build wireinject

package main

import (
	"github.com/google/wire"
	"mbook/webook/pkg/wego"
	"mbook/webook/reward/grpc"
	"mbook/webook/reward/ioc"
	"mbook/webook/reward/repository"
	"mbook/webook/reward/repository/cache"
	"mbook/webook/reward/repository/dao"
	"mbook/webook/reward/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis)

func Init() *wego.App {
	wire.Build(thirdPartySet,
		service.NewWechatNativeRewardService,
		ioc.InitAccountClient,
		ioc.InitGRPCxServer,
		ioc.InitPaymentClient,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
