package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct{}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gob.Register(time.Now())
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			//中断，不往后执行
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		now := time.Now()
		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			//用户第一次进来
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
