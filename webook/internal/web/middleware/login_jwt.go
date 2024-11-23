package middleware

import (
	ijwt "basic-project/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type LoginJWTMiddleWareBuilder struct {
	paths []string
	ijwt.Handler
}

func NewLoginJWTMiddleWareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddleWareBuilder {
	return &LoginJWTMiddleWareBuilder{
		Handler: jwtHdl,
	}
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
		tokenStr := l.ExtractToken(ctx)
		claims := ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
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
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
