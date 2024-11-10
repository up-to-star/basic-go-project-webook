package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type LoginJWTMiddleWareBuilder struct {
	paths []string
}

func NewLoginJWTMiddleWareBuilder() *LoginJWTMiddleWareBuilder {
	return &LoginJWTMiddleWareBuilder{}
}

func (l *LoginJWTMiddleWareBuilder) IgnorePath(path string) *LoginJWTMiddleWareBuilder {
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
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("BTv_D7]5q+f)9MTLwAA'5N!PJ6d6PNQQ"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
