package fiberapp

import "github.com/gofiber/fiber/v2"
import "strings"

type HttpHandler struct {
	Method   string
	Path     string
	Handlers []fiber.Handler
}

type Registry struct {
	HttpHandlers    map[string]*HttpHandler
	HttpMiddlewares map[string]interface{}
}

func NewRegistry() *Registry {
	return &Registry{
		HttpHandlers:    map[string]*HttpHandler{},
		HttpMiddlewares: map[string]interface{}{},
	}
}

func (r *Registry) AddHttpHandlers(handlers []*HttpHandler) {
	for _, handler := range handlers {
		r.HttpHandlers[createHandlerID(handler.Method, handler.Path)] = handler
	}
}

func (r *Registry) AddHttpMiddleware(path string, handler interface{}) {
	r.HttpMiddlewares[path] = handler
}

func (r *Registry) GetHttpHandler(method, path string) *HttpHandler {
	id := createHandlerID(method, path)
	return r.HttpHandlers[id]
}

func createHandlerID(method, path string) string {
	return strings.ToLower(method + " " + path)
}

func (r *Registry) RegisterHandlers(app *fiber.App) {
	for _, handler := range r.HttpHandlers {
		app.Add(handler.Method, handler.Path, handler.Handlers...)
	}
}

func (r *Registry) RegisterMiddlewares(app *fiber.App) {
	for path, middleware := range r.HttpMiddlewares {
		app.Use(path, middleware)
	}
}
