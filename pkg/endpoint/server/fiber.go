package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	mLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/pkg/errors"
)

func Fiber(config func(r *fiber.App)) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: handleError})

	app.Use(requestid.New(requestid.Config{ContextKey: keyReqId}))
	app.Use(mLogger.New())
	app.Use(handleAccessLog())
	app.Use(recover.New())

	config(app)
	return app
}

func handleError(ctx *fiber.Ctx, err error) error {
	requestId := xid.New().String()
	response := errors.Http(requestId, err)
	return ctx.Status(response.Error.Code).JSON(response)
}

func handleAccessLog() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		st := time.Now()

		err := ctx.Next()

		fields := map[string]interface{}{
			lfReqId:  ctx.Locals(keyReqId),
			lfStatus: errors.CodeToStr(errors.OK),
			lfCode:   fiber.StatusOK,
			lfError:  "",
		}

		if err != nil {
			code := errors.Code(err)
			fields[lfCode] = errors.CodeToHttp(code)
			fields[lfStatus] = errors.CodeToStr(code)
			fields[lfError] = err.Error()
		}

		message := fmt.Sprintf("[%v] %v(%v) - %v %v - %v",
			time.Since(st).Truncate(time.Millisecond),
			fields[lfStatus], fields[lfCode],
			ctx.Method(), ctx.OriginalURL(),
			string(ctx.Body()))

		logger := logrus.WithFields(fields)
		if fields[lfCode] == http.StatusOK {
			logger.Info(message)
		} else {
			logger.Warn(message)
		}
		return err
	}
}
