package fiberapp

import "github.com/gofiber/fiber/v2"

type HttpHandler struct {
	Method  string
	Path    string
	Handler fiber.Handler
	Meta    map[string]interface{}
}

type Registry struct {
	HttpHandlers    map[string]*HttpHandler
	HttpMiddlewares [][]interface{}
}

func NewRegistry() *Registry {
	return &Registry{
		HttpHandlers:    map[string]*HttpHandler{},
		HttpMiddlewares: [][]interface{}{},
	}
}

func (r *Registry) AddHttpHandlers(handlers []*HttpHandler) {
	for _, handler := range handlers {
		r.HttpHandlers[handler.Path] = handler
	}
}

func (r *Registry) AddHttpMiddleware(path string, handler interface{}) {
	r.HttpMiddlewares = append(r.HttpMiddlewares, []interface{}{path, handler})
}

func (r *Registry) GetHttpHandler(path string) *HttpHandler {
	return r.HttpHandlers[path]
}

func (r *Registry) RegisterHandlers(app *fiber.App) {
	for _, middleware := range r.HttpMiddlewares {
		app.Use(middleware...)
	}

	for _, handler := range r.HttpHandlers {
		app.Add(handler.Method, handler.Path, handler.Handler)
	}
}
