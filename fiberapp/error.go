package fiberapp

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
