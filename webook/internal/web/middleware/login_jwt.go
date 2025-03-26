package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"mbook/webook/internal/web"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct{}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" {
			return
		}
		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			//没登陆，连authorization这个头部都没有
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			//authorization内容是乱传的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTkey, nil
		})
		if err != nil {
			//token 不对
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			//token解析出来了，但是token为非法的或者过期的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("user-agent") {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		//其实valid就可以判断过期
		expireTime := uc.ExpiresAt
		//if expireTime.Before(time.Now()) {
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		if expireTime.Sub(time.Now()) < time.Minute*29 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
			tokenStr, err = token.SignedString(web.JWTkey)
			ctx.Header("x-jwt-token", tokenStr)
			if err != nil {
				log.Println(err)
			}
		}
		ctx.Set("user", uc)
	}
}
