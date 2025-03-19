package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"mbook/mbook/internal/web"
	"mbook/mbook/internal/web/middleware"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	return server
}
func InitGinMiddlewares() []gin.HandlerFunc {
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

		loginJWTMiddlewareBuilder.CheckLogin(),
	}
}
