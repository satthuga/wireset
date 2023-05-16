package server

type Feature interface {
	GetName() string
	Register()
}
