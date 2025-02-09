package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type MiddlewareLogger struct {
	allowReqBody  bool
	allowRespBody bool
	logFunc       func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareLogger {
	return &MiddlewareLogger{
		logFunc: fn,
	}
}

func (b *MiddlewareLogger) AllowReqBody() *MiddlewareLogger {
	b.allowReqBody = true
	return b
}

func (b *MiddlewareLogger) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		if ctx.Request.Body != nil && b.allowReqBody {
			body, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			al.ReqBody = string(body)
		}

		if b.allowRespBody {
			ctx.Writer = responseWriter{
				ResponseWriter: ctx.Writer,
				al:             al,
			}
		}
		defer func() {
			duration := time.Since(start).String()
			al.Duration = duration
			b.logFunc(ctx, al)
		}()
		ctx.Next()

	}
}

func (b *MiddlewareLogger) AllowRespBody() *MiddlewareLogger {
	b.allowRespBody = true
	return b
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w responseWriter) Write(data []byte) (int, error) {
	if len(data) > 1024 {
		data = data[:1024]
	}
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w responseWriter) WriteString(data string) (int, error) {
	if len(data) > 1024 {
		data = data[:1024]
	}
	w.al.RespBody = data
	return w.ResponseWriter.WriteString(data)
}

type AccessLog struct {
	Method   string
	Url      string
	Duration string
	ReqBody  string
	RespBody string
	Status   int
}
