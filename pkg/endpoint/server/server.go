package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/pkg/errors"
	"github.com/GotaX/go-server-skeleton/pkg/ext/tracing"
)

const (
	keyReqId = "server.request_id"

	lf       = "log.fields"
	lfReqId  = "request_id"
	lfCode   = "code"
	lfStatus = "status"
	lfError  = "error"
)

func Gin(router func(gin.IRouter)) http.Handler {
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: logrus.StandardLogger().WriterLevel(logrus.DebugLevel),
	}))
	r.Use(gin.Recovery())
	r.Use(genRequestId())
	r.Use(accessLog())

	router(r)
	return r
}

func genRequestId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		info := tracing.GetRequestInfo(ctx.Request.Context())
		ctx.Set(keyReqId, info.String())
		ctx.Next()
	}
}

func accessLog() gin.HandlerFunc {
	pool := &sync.Pool{
		New: func() interface{} { return &bytes.Buffer{} },
	}

	return func(ctx *gin.Context) {
		buf := pool.Get().(*bytes.Buffer)
		defer pool.Put(buf)
		buf.Reset()

		reader := ctx.Request.Body
		defer func() { _ = reader.Close() }()

		ctx.Request.Body = ioutil.NopCloser(io.TeeReader(reader, buf))
		st := time.Now()

		ctx.Set(lf, map[string]interface{}{
			lfReqId:  ctx.GetString(keyReqId),
			lfStatus: errors.CodeToStr(errors.OK),
			lfCode:   http.StatusOK,
			lfError:  "",
		})

		ctx.Next()

		fields := ctx.GetStringMap(lf)
		message := fmt.Sprintf("[%v] %v(%v) - %v %v - %v",
			time.Since(st).Truncate(time.Millisecond),
			fields[lfStatus], fields[lfCode],
			ctx.Request.Method, ctx.Request.URL,
			buf.String())

		logger := logrus.WithFields(fields)
		if fields[lfCode] == http.StatusOK {
			logger.Info(message)
		} else {
			logger.Warn(message)
		}
	}
}

func RenderError(ctx *gin.Context, op errors.Op, err error) {
	requestId := ctx.GetString(keyReqId)
	resp := errors.Http(requestId, errors.E(op, err))

	fields := ctx.GetStringMap(lf)
	fields[lfStatus] = resp.Error.Status
	fields[lfCode] = resp.Error.Code
	fields[lfError] = resp.Error.Message

	ctx.AbortWithStatusJSON(resp.Error.Code, resp)
}

func NotFound(name, id string) error {
	return errors.E(errors.NotFound,
		&errors.NotFoundError{Type: name, ID: id})
}
