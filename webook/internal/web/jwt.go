package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type jwtHandler struct {
	signingMethod jwt.SigningMethod
	refreshKey    []byte
}

func newJwtHandler() jwtHandler {
	return jwtHandler{
		signingMethod: jwt.SigningMethodHS512,
		refreshKey:    []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgA"),
	}
}

func (h *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) {
	uc := UserClaims{
		Uid:       uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			//30分钟过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTkey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登陆成功")
}
func (h *jwtHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	rc := RefreshClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(h.refreshKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
func ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	tokenStr := segs[1]
	return tokenStr
}

var JWTkey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid int64
}
type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
