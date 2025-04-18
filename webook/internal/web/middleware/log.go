package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, l AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, al AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{logFn: logFn}
}

// AllowReqBody 为了链式调用
func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

// AllowRespBody 为了链式调用
func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := ctx.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody {
			// Request.Body 是一个 Stream 对象，只能读一次
			body, _ := ctx.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			// 放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			//ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             &al,
			}
		}
		start := time.Now()
		defer func() {
			al.Duration = time.Since(start)
			l.logFn(ctx, al)
		}()
		// 直接执行下一个 middleware...直到业务逻辑
		ctx.Next()
		// 在这里，你就拿到了响应
	}
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	Status   int           `json:"status"`
	RespBody string        `json:"resp_body"`
	Duration time.Duration `json:"duration"`
}

// 装饰器
type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
