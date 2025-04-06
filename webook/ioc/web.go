package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"mbook/webook/internal/web"
	"mbook/webook/internal/web/middleware"
	"mbook/webook/pkg/ginx/middleware/ratelimit"
	"mbook/webook/pkg/limiter"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	return server
}
func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	loginJWTMiddlewareBuilder := middleware.LoginJWTMiddlewareBuilder{}
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,

			AllowHeaders:  []string{"Content-Type", "Authorization"},
			ExposeHeaders: []string{"x-jwt-token"},
			AllowOriginFunc: func(origin string) bool {
				return strings.HasPrefix(origin, "http://localhost")
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(
			redisClient, time.Second, 1000)).Build(),
		loginJWTMiddlewareBuilder.CheckLogin(),
	}
}
