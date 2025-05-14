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
	ijwt "mbook/webook/internal/web/jwt"
	"mbook/webook/ioc"
)

var thirdPartySet = wire.NewSet(
	//第三方依赖
	InitRedis, InitDB,
	InitLogger,
)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,
		cache.NewUserCache, cache.NewCodeCache,
		cache.NewArticleRedisCache,

		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		ioc.InitSMSService,
		service.NewUserService, service.NewCodeService,
		service.NewArticleService,
		InitWechatService,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		cache.NewArticleRedisCache,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
