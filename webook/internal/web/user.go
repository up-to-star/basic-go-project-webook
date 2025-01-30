package web

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/service"
	ijwt "basic-project/webook/internal/web/jwt"
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	emailRegexPattern    = "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$"
	passwordRegexPattern = "^(?=.*[a-zA-Z])(?=.*[0-9])(?=.*[!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>\\/?]).{8,}$"
	phoneRegexPattern    = "^1[3-9]\\d{9}$"
	biz                  = "login"
)

// UserHandle 定义和 user 用户有关的路由
type UserHandle struct {
	ijwt.Handler
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	phoneExp    *regexp.Regexp
	cmd         redis.Cmdable
}

func NewUserHandle(svc service.UserService, codeSvc service.CodeService, cmd redis.Cmdable, jwtHdl ijwt.Handler) *UserHandle {
	return &UserHandle{
		svc:         svc,
		emailExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		codeSvc:     codeSvc,
		phoneExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		cmd:         cmd,
		Handler:     jwtHdl,
	}
}

func (u *UserHandle) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", u.Signup)
	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
	ug.POST("/login_sms/code/send", u.SendLoginSmsCode)
	ug.POST("/login_sms", u.LoginSMS)
	ug.POST("/logout", u.Logout)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandle) RefreshToken(ctx *gin.Context) {
	refreshTokenStr := u.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshTokenStr, &rc, func(*jwt.Token) (interface{}, error) {
		return ijwt.RtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = u.CheckSession(ctx, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if err = u.SetJWTToken(ctx, rc.Uid, rc.Ssid); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "refresh token success",
	})
}

func (u *UserHandle) LoginSMS(ctx *gin.Context) {
	type LoginReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		zap.L().Error("绑定错误", zap.Error(err))
		return
	}

	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		zap.L().Error("手机号码匹配异常", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, &Result{
			Code: 4,
			Msg:  "手机号输入错误",
		})
		zap.L().Error("手机号输入错误", zap.Error(err))
		return
	}
	ok, err = u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		zap.L().Error("验证码校验异常", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, &Result{
			Code: 4,
			Msg:  "验证码有误",
		})
		zap.L().Error("验证码有误", zap.Error(err))
		return
	}

	user, err := u.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}

	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, &Result{
		Code: 0,
		Msg:  "登录成功",
	})

}

func (u *UserHandle) SendLoginSmsCode(ctx *gin.Context) {
	type SmsReq struct {
		Phone string `json:"phone"`
	}
	var req SmsReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	// 校验手机号
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, &Result{
			Code: 4,
			Msg:  "手机号输入错误",
		})
		return
	}
	err = u.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, &Result{
			Code: 0,
			Msg:  "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, &Result{
			Code: 4,
			Msg:  "发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
	}
}

func (u *UserHandle) Signup(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 校验email
	ok, err := u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不对")
		return
	}

	// 校验密码
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码至少8位, 包含字母、数字和特殊字符")
		return
	}
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}
	err = u.svc.Signup(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrUserDuplicateEmail) {
		span := trace.SpanFromContext(ctx.Request.Context())
		span.AddEvent("邮箱冲突")
		ctx.String(http.StatusOK, "邮箱冲突")
		zap.L().Error("邮箱冲突", zap.Error(err))
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

// Login session 版本的login
func (u *UserHandle) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "邮箱或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	// 登录成功
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		MaxAge: 30 * 60,
	})
	_ = sess.Save()
	ctx.String(http.StatusOK, "登录成功")
}

// LoginJWT session 版本的login
func (u *UserHandle) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.JSON(http.StatusOK, &Result{
			Code: 4,
			Msg:  "邮箱或密码错误",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	// 登录成功, jwt 设置登录状态

	if err = u.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, &Result{
		Code: 0,
		Msg:  "登录成功",
	})
}

func (u *UserHandle) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	tokenStr := u.ExtractToken(ctx)
	var claims ijwt.UserClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	t, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	user := domain.User{
		Id:       claims.Uid,
		Nickname: req.Nickname,
		Birthday: t,
		AboutMe:  req.AboutMe,
	}

	err = u.svc.Edit(ctx, user)

	ctx.JSON(http.StatusOK, &Result{
		Code: 0,
		Msg:  "修改成功",
	})

}

func (u *UserHandle) Profile(ctx *gin.Context) {
	var claims ijwt.UserClaims
	tokenStr := u.ExtractToken(ctx)
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	user, err := u.svc.Profile(ctx, claims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}

	type ProfileData struct {
		Nickname string `json:"Nickname"`
		Email    string `json:"Email"`
		Phone    string `json:"Phone"`
		Birthday string `json:"Birthday"`
		AboutMe  string `json:"AboutMe"`
	}
	ctx.JSON(http.StatusOK, &ProfileData{
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Birthday: user.Birthday.Format("2006-01-02"),
		AboutMe:  user.AboutMe,
	})
}

func (u *UserHandle) Logout(ctx *gin.Context) {
	// 清除token
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	tokenStr := u.ExtractToken(ctx)
	var claims ijwt.UserClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = u.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, &Result{
		Code: 0,
		Msg:  "退出登录成功",
	})
}
