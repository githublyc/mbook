package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"mbook/mbook/internal/domain"
	"mbook/mbook/internal/service"
	"net/http"
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
	ug.POST("/login", h.Login)
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
	h.svc.Signup(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
}

func (h *UserHandler) Login(context *gin.Context) {

}

func (h *UserHandler) Edit(context *gin.Context) {

}

func (h *UserHandler) Profile(context *gin.Context) {

}
