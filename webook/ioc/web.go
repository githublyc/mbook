package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"mbook/webook/internal/web"
	ijwt "mbook/webook/internal/web/jwt"
	"mbook/webook/internal/web/middleware"
	"mbook/webook/pkg/ginx/middleware/ratelimit"
	"mbook/webook/pkg/limiter"
	"mbook/webook/pkg/logger"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	artHdl *web.ArticleHandler,
	wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	return server
}
func InitGinMiddlewares(redisClient redis.Cmdable,
	hdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	loginJWTMiddlewareBuilder := middleware.NewLoginJWTMiddlewareBuilder(hdl)

	logMiddlewareBuilder := middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
		l.Debug("", logger.Field{"req", al})
	}).AllowReqBody().AllowRespBody()

	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,

			AllowHeaders: []string{"Content-Type", "Authorization"},
			//这个是允许前端访问你的后端响应中带的头部
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
			AllowOriginFunc: func(origin string) bool {
				return strings.HasPrefix(origin, "http://localhost")
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(
			redisClient, time.Second, 1000)).Build(),
		logMiddlewareBuilder.Build(),
		loginJWTMiddlewareBuilder.CheckLogin(),
	}
}
