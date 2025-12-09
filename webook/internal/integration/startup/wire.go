//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "mbook/webook/interactive/repository"
	cache2 "mbook/webook/interactive/repository/cache"
	dao2 "mbook/webook/interactive/repository/dao"
	service2 "mbook/webook/interactive/service"
	"mbook/webook/internal/events/article"
	"mbook/webook/internal/repository"
	"mbook/webook/internal/repository/cache"
	"mbook/webook/internal/repository/dao"
	"mbook/webook/internal/service"
	"mbook/webook/internal/web"
	ijwt "mbook/webook/internal/web/jwt"
	"mbook/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var thirdPartySet = wire.NewSet(
	//第三方依赖
	InitRedis, InitDB,
	InitLogger,
	InitSaramaClient,
	InitSyncProducer,
)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService,
)
var articleSvcProvider = wire.NewSet(
	dao.NewArticleGORMDAO,
	cache.NewArticleRedisCache,
	repository.NewCachedArticleRepository,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcSet,

		cache.NewRedisCodeCache,

		repository.NewCodeRepository,

		article.NewSaramaSyncProducer,

		ioc.InitSMSService,
		service.NewCodeService,
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
		interactiveSvcSet,
		cache.NewArticleRedisCache,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
		article.NewSaramaSyncProducer,
	)
	return &web.ArticleHandler{}
}
