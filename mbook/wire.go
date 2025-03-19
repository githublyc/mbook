//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"mbook/mbook/internal/repository"
	"mbook/mbook/internal/repository/cache"
	"mbook/mbook/internal/repository/dao"
	"mbook/mbook/internal/service"
	"mbook/mbook/internal/web"
	"mbook/mbook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitRedis, ioc.InitDB,

		dao.NewUserDao,
		cache.NewUserCache, cache.NewCodeCache,
		repository.NewCachedUserRepository, repository.NewCodeRepository,
		ioc.InitSMSService,
		service.NewUserService, service.NewCodeService,
		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
