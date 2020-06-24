package srvrest

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/GotaX/go-server-skeleton/internal/example/pkg/cfg"
	"github.com/GotaX/go-server-skeleton/internal/example/pkg/rpc"
	"github.com/GotaX/go-server-skeleton/pkg/endpoint/server"
	"github.com/GotaX/go-server-skeleton/pkg/errors"
)

func Router() http.Handler {
	return server.Gin(func(r gin.IRouter) {
		// Connect grpc
		var cc *grpc.ClientConn
		cfg.Local(&cc)

		client := rpc.NewHelloServiceClient(cc)
		logrus.Debug("rpc connected")

		r.Any("",
			gzip.Gzip(gzip.DefaultCompression),
			HelloHandler(client))
	})
}

func HelloHandler(client rpc.HelloServiceClient) func(*gin.Context) {
	const op errors.Op = "server.HelloHandler"

	return func(c *gin.Context) {
		ctx, _ := context.WithTimeout(c.Request.Context(), time.Second)
		req := &rpc.HelloRequest{Greeting: "hero"}
		resp, err := client.SayHello(ctx, req)
		if err != nil {
			server.RenderError(c, op, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}
