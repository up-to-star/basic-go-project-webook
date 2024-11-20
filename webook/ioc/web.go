package ioc

import (
	"basic-project/webook/internal/pkg/ginx/middlewares/ratelimit"
	ratelimit2 "basic-project/webook/internal/pkg/ratelimit"
	"basic-project/webook/internal/web"
	"basic-project/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
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
		ratelimit.NewBuilder(ratelimit2.NewRedisSlideWindowLimiter(redisClient, time.Second, 100)).Build(),
		middleware.NewLoginJWTMiddleWareBuilder().
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").Build(),
	}
}

func InitWebserver(mdls []gin.HandlerFunc, userHdl *web.UserHandle) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	return server
}
