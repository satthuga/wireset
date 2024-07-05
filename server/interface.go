package server

import "context"

type Feature interface {
	Init() error
	Name() string
}

type Server interface {
	Start(ctx context.Context) chan error
}
