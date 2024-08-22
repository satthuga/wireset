package fiberapp

import (
	"encoding/json"
	"errors"
	"github.com/aiocean/wireset/configsvc"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/wire"
	"go.uber.org/zap"
	"net/url"
	"os"
	"time"
)

var DefaultWireset = wire.NewSet(
	NewFiberApp,
	NewRegistry,
	NewHealthRegistry,
)

type FiberAppConfig struct {
	BodyLimit   int
	ServiceName string
	IdleTimeout time.Duration
	MaxRequests int
	RateLimit   time.Duration
	ProxyURL    string
}

func NewFiberApp(
	logsvc *zap.Logger,
	cfg *configsvc.ConfigService,
	healthRegistry *HealthRegistry,
) (*fiber.App, func(), error) {
	logger := logsvc.With(zap.Strings("tags", []string{"fiber"}))

	config := FiberAppConfig{
		BodyLimit:   50 * 1024 * 1024,
		ServiceName: cfg.ServiceName,
		IdleTimeout: 10 * time.Second,
		MaxRequests: 500,
		RateLimit:   30 * time.Second,
		ProxyURL:    os.Getenv("PROXY_URL"),
	}

	app := fiber.New(fiber.Config{
		BodyLimit:             config.BodyLimit,
		AppName:               config.ServiceName,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
		IdleTimeout:           config.IdleTimeout,
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
		SkipURIs: []string{"/healthz"},
	}))

	// compress
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(idempotency.New())
	app.Use(limiter.New(limiter.Config{
		Max:               config.MaxRequests,
		Expiration:        config.RateLimit,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(requestid.New())

	cleanup := func() {
		if err := app.Shutdown(); err != nil {
			logger.Error("failed to shut down fiber app", zap.Error(err))
			return
		}

		logger.Info("fiber app shut down")
	}

	// this is used for local development, to proxy to the real endpoint
	if config.ProxyURL != "" {
		app.Use(func(c *fiber.Ctx) error {
			if c.Path() == "/healthz" {
				return c.Next()
			}
			endpointUrl, _ := url.JoinPath(config.ProxyURL, c.Path())
			return proxy.Forward(endpointUrl)(c)
		})
	}

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			if err := healthRegistry.Check(); err != nil {
				return false
			}

			return true
		},
		LivenessEndpoint: "/healthz",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		ReadinessEndpoint: "/ready",
	}))

	return app, cleanup, nil
}
