package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
		updateTime := sess.Get("update_time")
		now := time.Now().UnixMilli()
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 30 * 60,
		})
		// 还没有刷新过
		if updateTime == nil {
			sess.Set("update_time", now)
			_ = sess.Save()
			return
		}

		// updateTime 存在
		updateTimeVal, ok := updateTime.(int64)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// 超过1分钟刷新一次
		if now-updateTimeVal > 60*1000 {
			sess.Set("update_time", now)
			_ = sess.Save()
			return
		}
	}
}
