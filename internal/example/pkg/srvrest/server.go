package srvrest

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2"
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

func Fiber() *fiber.App {
	return server.Fiber(func(r *fiber.App) {
		var cc *grpc.ClientConn
		cfg.Local(&cc)

		client := rpc.NewHelloServiceClient(cc)
		logrus.Debug("rpc connected")

		r.All("", FiberHelloHandler(client))
	})
}

func FiberHelloHandler(client rpc.HelloServiceClient) func(*fiber.Ctx) error {
	const op errors.Op = "server.HelloHandler"

	return func(c *fiber.Ctx) error {
		ctx, _ := context.WithTimeout(c.Context(), time.Second)
		req := &rpc.HelloRequest{Greeting: "hero"}
		resp, err := client.SayHello(ctx, req)
		if err != nil {
			return errors.E(op, err)
		}

		return c.JSON(resp)
	}
}
