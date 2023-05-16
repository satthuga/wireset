package configsvc

import (
	"errors"
	"os"

	"github.com/google/wire"
)

type ConfigService struct {
	ServiceName string
	ServiceUrl  string
	AppUrl      string
	Address     string
	Port        string
	Environment string
}

type DatabaseConfig struct {
	ConnectionUri string
}

var Wireset = wire.NewSet(NewConfigService)

func NewConfigService() (*ConfigService, error) {
	configService := &ConfigService{
		ServiceName: os.Getenv("SERVICE_NAME"),
		ServiceUrl:  os.Getenv("SERVICE_URL"),
		AppUrl:      os.Getenv("APP_URL"),
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

	if value, ok := os.LookupEnv("APP_URL"); ok {
		configService.AppUrl = value
	} else {
		return nil, errors.New("APP_URL is required")
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
