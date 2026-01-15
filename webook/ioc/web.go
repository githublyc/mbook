package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"mbook/webook/internal/web"
	ijwt "mbook/webook/internal/web/jwt"
	"mbook/webook/internal/web/middleware"
	"mbook/webook/pkg/ginx"
	"mbook/webook/pkg/ginx/middleware/prometheus"
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

	logMiddlewareBuilder := middleware.NewLogMiddlewareBuilder(
		func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("", logger.Field{"req", al})
		}).AllowReqBody().AllowRespBody()

	pb := &prometheus.Builder{
		Namespace: "lyc",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 GIN 的HTTP接口数据",
	}
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})

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
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		otelgin.Middleware("webook"),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(
			redisClient, time.Second, 1000)).Build(),
		logMiddlewareBuilder.Build(),
		loginJWTMiddlewareBuilder.CheckLogin(),
	}
}
