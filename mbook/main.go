package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})
	server.Run(":8080")
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
