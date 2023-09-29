package configsvc

import (
	"errors"
	"os"

	"github.com/google/wire"
)

type ConfigService struct {
	ServiceName         string
	ServiceUrl          string
	Address             string
	Port                string
	Environment         string
	DataDogAgentAddress string
}

type DatabaseConfig struct {
	ConnectionUri string
}

var EnvWireset = wire.NewSet(NewConfigFromEnv)

// NewConfigFromEnv creates a new ConfigService from environment variables.
// It returns an error if any of the required environment variables are missing.
func NewConfigFromEnv() (*ConfigService, error) {
	configService := &ConfigService{
		ServiceName: os.Getenv("SERVICE_NAME"),
		ServiceUrl:  os.Getenv("SERVICE_URL"),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	if value, ok := os.LookupEnv("ENVIRONMENT"); ok {
		configService.Environment = value
	} else {
		return nil, errors.New("ENVIRONMENT is required")
	}

	if value, ok := os.LookupEnv("SERVICE_NAME"); ok {
		configService.ServiceName = value
	} else {
		return nil, errors.New("SERVICE_NAME is required")
	}

	if value, ok := os.LookupEnv("SERVICE_URL"); ok {
		configService.ServiceUrl = value
	} else {
		return nil, errors.New("SERVICE_URL is required")
	}

	if value, ok := os.LookupEnv("ADDRESS"); ok {
		configService.Address = value
	} else {
		configService.Address = ""
	}

	if value, ok := os.LookupEnv("PORT"); ok {
		configService.Port = value
	} else {
		configService.Port = "8080"
	}

	return configService, nil
}

// IsProduction
func (c *ConfigService) IsProduction() bool {
	return c.Environment == "production"
}
