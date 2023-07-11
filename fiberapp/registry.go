package fiberapp

import "github.com/gofiber/fiber/v2"

type HttpHandler interface {
	Register(app *fiber.App)
}

type Registry struct {
	HttpHandlers []HttpHandler
}

func NewRegistry() *Registry {
	return &Registry{
		HttpHandlers: []HttpHandler{},
	}
}

func (r *Registry) AddHttpHandler(handler HttpHandler) {
	r.HttpHandlers = append(r.HttpHandlers, handler)
}

func (r *Registry) RegisterHandlers(app *fiber.App) {
	for _, handler := range r.HttpHandlers {
		handler.Register(app)
	}
}
