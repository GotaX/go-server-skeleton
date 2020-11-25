package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/xid"

	"github.com/GotaX/go-server-skeleton/pkg/errors"
)

func Fiber(config func(r *fiber.App)) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: handleError})

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(recover.New())

	config(app)
	return app
}

func handleError(ctx *fiber.Ctx, err error) error {
	requestId := xid.New().String()
	response := errors.Http(requestId, err)
	return ctx.Status(response.Error.Code).JSON(response)
}
