package fiberapp

import (
	"api/pkg/configsvc"
	"api/pkg/fiberapp/middleware"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"time"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	NewFiberApp,
	middleware.NewAuthzController,
)

func NewFiberApp(
	logsvc *zap.Logger,
	authzMiddleware *middleware.AuthzController,
	cfg *configsvc.ConfigService,
) (*fiber.App, func(), error) {
	app := fiber.New(fiber.Config{
		AppName:               cfg.ServiceName,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: false,
		IdleTimeout:           10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	logger := logsvc.With(zap.Strings("tags", []string{"FiberApp"}))

	app.Use(cors.New())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger,
	}))
	app.Use(middleware.AllowEmbed())
	app.Use(authzMiddleware.Middleware())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(idempotency.New())

	cleanup := func() {
		logger.Info("Cleaning up")
		err := app.Shutdown()
		if err != nil {
			logger.Error("failed to shut down fiber app", zap.Error(err))
			return
		}

		logger.Info("fiber app shut down")
	}

	return app, cleanup, nil
}
