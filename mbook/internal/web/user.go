package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"mbook/mbook/internal/domain"
	"mbook/mbook/internal/service"
	"net/http"
	"time"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	emailRegexp    *regexp.Regexp
	passwordRegexp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRegexp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

// 分散注册路由，非集中式
func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	//分组路由
	ug := server.Group("/users")
	ug.POST("/signup", h.SignUp)
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(context *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignupReq
	if err := context.Bind(&req); err != nil {
		return
	}
	isEmail, err := h.emailRegexp.MatchString(req.Email)
	if err != nil {
		context.String(http.StatusOK, "系统错误")
	}
	if !isEmail {
		context.String(http.StatusOK, "非法邮箱格式")
	}
	if req.Password != req.ConfirmPassword {
		context.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordRegexp.MatchString(req.Password)
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		context.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}
	err = h.svc.Signup(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		context.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		context.String(http.StatusOK, "邮箱冲突")
	default:
		context.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Login(context *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := context.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(context, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(context)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			MaxAge: 900,
		})
		err := sess.Save()
		if err != nil {
			context.String(http.StatusOK, "系统错误")
			return
		}
		context.String(http.StatusOK, "登陆成功")
	case service.ErrInvalidUserOrPassword:
		context.String(http.StatusOK, "用户名或密码不对")
	default:
		context.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Edit(context *gin.Context) {

}

func (h *UserHandler) Profile(context *gin.Context) {

}

func (h *UserHandler) LoginJWT(context *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := context.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(context, req.Email, req.Password)
	switch err {
	case nil:
		uc := UserClaims{
			Uid:       u.Id,
			UserAgent: context.GetHeader("User-Agent"),
			RegisteredClaims: jwt.RegisteredClaims{
				//30分钟过期
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
		tokenStr, err := token.SignedString(JWTkey)
		if err != nil {
			context.String(http.StatusOK, "系统错误")
		}
		context.Header("x-jwt-token", tokenStr)
		context.String(http.StatusOK, "登陆成功")
	case service.ErrInvalidUserOrPassword:
		context.String(http.StatusOK, "用户名或密码不对")
	default:
		context.String(http.StatusOK, "系统错误")
	}
}

var JWTkey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
