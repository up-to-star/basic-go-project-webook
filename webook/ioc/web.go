package ioc

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/web"
	ijwt "github.com/basic-go-project-webook/webook/internal/web/jwt"
	"github.com/basic-go-project-webook/webook/internal/web/middleware"
	"github.com/basic-go-project-webook/webook/pkg/ginx/middlewares/logger"
	"github.com/basic-go-project-webook/webook/pkg/ginx/middlewares/ratelimit"
	ratelimit2 "github.com/basic-go-project-webook/webook/pkg/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"strings"
	"time"
)

func InitGinMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			//AllowOrigins: []string{"http://localhost:3000"},
			//AllowMethods: []string{"PUT", "PATCH", "POST"},
			AllowHeaders:  []string{"Content-Type", "Authorization"},
			ExposeHeaders: []string{"x-jwt-token"},
			// 是否允许带 cookie 之类的东西
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				if strings.Contains(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, "www.xxx.com")
			},
			MaxAge: 12 * time.Hour,
		}),
		otelgin.Middleware("webook"),
		ratelimit.NewBuilder(ratelimit2.NewRedisSlideWindowLimiter(redisClient, time.Second, 100)).Build(),
		middleware.NewLoginJWTMiddleWareBuilder(jwtHdl).
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/articles/edit").
			IgnorePaths("/users/login_sms").Build(),
		logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
			zap.L().Debug("HTTP请求", zap.Any("AccessLog", al))
		}).Build(),
	}
}

func InitWebserver(mdls []gin.HandlerFunc, userHdl *web.UserHandle,
	oauth2WechatHandler *web.OAuth2WechatHandler, artHdl *web.ArticleHandle) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2WechatHandler.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}
