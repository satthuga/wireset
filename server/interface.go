package server

import "context"

type Feature interface {
	Init() error
}

type Server interface {
	Start(ctx context.Context) chan error
}
