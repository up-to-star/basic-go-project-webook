package middleware

import (
	"basic-project/webook/internal/web"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type LoginJWTMiddleWareBuilder struct {
	paths []string
	cmd   redis.Cmdable
}

func NewLoginJWTMiddleWareBuilder() *LoginJWTMiddleWareBuilder {
	return &LoginJWTMiddleWareBuilder{}
}

func (l *LoginJWTMiddleWareBuilder) IgnorePaths(path string) *LoginJWTMiddleWareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddleWareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// JWT 校验
		tokenStr := web.ExtractToken(ctx)
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("BTv_D7]5q+f)9MTLwAA'5N!PJ6d6PNQQ"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if ctx.Request.UserAgent() != claims.UserAgent {
			// 安全问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		cnt, err := l.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid)).Result()
		if err != nil || cnt > 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
