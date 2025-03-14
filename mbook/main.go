package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mbook/mbook/internal/repository"
	"mbook/mbook/internal/repository/dao"
	"mbook/mbook/internal/service"
	"mbook/mbook/internal/web"
	"mbook/mbook/internal/web/middleware"
	"mbook/mbook/pkg/ginx/middleware/ratelimit"
	"mbook/mbook/pkg/limiter"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHdl(db, server)
	server.Run(":8080")
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDao(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}
func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {

	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,

		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token"},
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		},
		MaxAge: 12 * time.Hour,
	}),
	)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	server.Use(ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build())
	useJWT(server)
	return server
}
func useJWT(server *gin.Engine) {
	loginJWTMiddlewareBuilder := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(loginJWTMiddlewareBuilder.CheckLogin())

}
func useSession(server *gin.Engine) {
	//存储数据的，也就是userId存哪里
	//直接存cookie
	//store := cookie.NewStore([]byte("secret"))
	//基于redis的实现
	//store, _ := redis.NewStore(16, "tcp", "localhost:6379", "", []byte(""))
	//loginMiddlewareBuilder := middleware.LoginMiddlewareBuilder{}
	//第一个是把session弄出来，包括所有接口，如login接口
	//第二个是利用session来做登录校验
	//server.Use(sessions.Sessions("ssid", store),loginMiddlewareBuilder.CheckLogin())

}
