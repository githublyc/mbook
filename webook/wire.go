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

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,

		dao.NewUserDAO,
		cache.NewUserCache, cache.NewCodeCache,
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,

		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService, service.NewCodeService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
