//go:build wireinject

package main

import (
	"github.com/google/wire"
	repository2 "mbook/webook/interactive/repository"
	cache2 "mbook/webook/interactive/repository/cache"
	dao2 "mbook/webook/interactive/repository/dao"
	service2 "mbook/webook/interactive/service"
	article2 "mbook/webook/internal/events/article"
	"mbook/webook/internal/repository"
	"mbook/webook/internal/repository/cache"
	"mbook/webook/internal/repository/dao"
	"mbook/webook/internal/service"
	"mbook/webook/internal/web"
	ijwt "mbook/webook/internal/web/jwt"
	"mbook/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)
var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedOnlyRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		//第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcd,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRlockClient,

		//interactiveSvcSet,
		//ioc.InitIntrClient,
		ioc.InitIntrClientV1,
		rankingSvcSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		article2.NewSaramaSyncProducer,
		//events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		cache.NewUserCache, cache.NewCodeCache,
		cache.NewArticleRedisCache,

		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService, service.NewCodeService,
		service.NewArticleService,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
