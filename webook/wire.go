//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"mbook/webook/internal/repository"
	"mbook/webook/internal/repository/cache"
	"mbook/webook/internal/repository/dao"
	"mbook/webook/internal/service"
	"mbook/webook/internal/web"
	ijwt "mbook/webook/internal/web/jwt"
	"mbook/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,

		interactiveSvcSet,

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
	)
	return gin.Default()
}
