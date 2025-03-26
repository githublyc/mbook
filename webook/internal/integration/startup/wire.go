//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"mbook/webook/internal/repository"
	"mbook/webook/internal/repository/cache"
	"mbook/webook/internal/repository/dao"
	"mbook/webook/internal/service"
	"mbook/webook/internal/web"
	"mbook/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		InitRedis, ioc.InitDB,

		dao.NewUserDAO,
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
