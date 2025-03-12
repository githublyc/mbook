package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mbook/mbook/internal/repository"
	"mbook/mbook/internal/repository/dao"
	"mbook/mbook/internal/service"
	"mbook/mbook/internal/web"
	"mbook/mbook/internal/web/middleware"
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
	store := cookie.NewStore([]byte("secret"))

	loginMiddlewareBuilder := middleware.LoginMiddlewareBuilder{}

	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type"},
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		},
		MaxAge: 12 * time.Hour,
	}),
		//第一个是把session弄出来，包括所有接口，如login接口
		sessions.Sessions("ssid", store),
		//第二个是利用session来做登录校验
		loginMiddlewareBuilder.CheckLogin(),
	)
	return server
}
