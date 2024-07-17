package fiberapp

import (
	"encoding/json"
	"errors"
	"github.com/aiocean/wireset/configsvc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
	"net/url"
	"os"
	"time"

	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	NewFiberApp,
	NewRegistry,
	NewHealthRegistry,
)

func NewFiberApp(
	logsvc *zap.Logger,
	cfg *configsvc.ConfigService,
	healthRegistry *HealthRegistry,
) (*fiber.App, func(), error) {

	logger := logsvc.With(zap.Strings("tags", []string{"fiber"}))

	app := fiber.New(fiber.Config{
		BodyLimit:             50 * 1024 * 1024,
		AppName:               cfg.ServiceName,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
		IdleTimeout:           10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			logger.Error("error", zap.Error(err))

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// enable middlewares
	app.Use(cors.New())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger,
	}))

	// compress
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(idempotency.New())
	app.Use(limiter.New(limiter.Config{
		Max:               500,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(requestid.New())

	cleanup := func() {
		logger.Info("Cleaning up")
		if err := app.Shutdown(); err != nil {
			logger.Error("failed to shut down fiber app", zap.Error(err))
			return
		}

		logger.Info("fiber app shut down")
	}

	// this is used for local development, to proxy to the real endpoint
	proxyUrl := os.Getenv("PROXY_URL")
	if proxyUrl != "" {

		app.Use(func(c *fiber.Ctx) error {
			if c.Path() == "/healthz" {
				return c.Next()
			}
			endpointUrl, _ := url.JoinPath(proxyUrl, c.Path())
			return proxy.Forward(endpointUrl)(c)
		})
	}

	app.Get("/healthz", func(c *fiber.Ctx) error {
		if err := healthRegistry.Check(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusOK)
	})

	return app, cleanup, nil
}
