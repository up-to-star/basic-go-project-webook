package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginMiddleWareBuilder struct {
	paths []string
}

func NewLoginMiddleWareBuilder() *LoginMiddleWareBuilder {
	return &LoginMiddleWareBuilder{}
}

func (l *LoginMiddleWareBuilder) IgnorePath(path string) *LoginMiddleWareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddleWareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
