package handler

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusHandler struct {
}

func (g *PrometheusHandler) prometheus(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}

func (g *PrometheusHandler) Register(fiberApp *fiber.App) {
	fiberApp.All("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}
