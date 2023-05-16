package shopifysvc

import (
	"github.com/pkg/errors"
	"os"
)

type Config struct {
	ClientId     string
	ClientSecret string
	Scope        string
	RedirectUrl  string
	ApiVersion   string
	LoginNonce   string
}

func ConfigFromEnv() (*Config, error) {
	config := &Config{}

	if value, ok := os.LookupEnv("SCOPES"); ok {
		config.Scope = value
	} else {
		return nil, errors.New("SCOPES is required")
	}

	if value, ok := os.LookupEnv("SHOPIFY_CLIENT_SECRET"); ok {
		config.ClientSecret = value
	} else {
		return nil, errors.New("SHOPIFY_CLIENT_SECRET is required")
	}

	if value, ok := os.LookupEnv("SHOPIFY_CLIENT_ID"); ok {
		config.ClientId = value
	} else {
		return nil, errors.New("SHOPIFY_CLIENT_ID is required")
	}

	if value, ok := os.LookupEnv("SHOPIFY_API_VERSION"); ok {
		config.ApiVersion = value
	} else {
		return nil, errors.New("SHOPIFY_API_VERSION is required")
	}

	if value, ok := os.LookupEnv("LOGIN_NONCE"); ok {
		config.LoginNonce = value
	} else {
		return nil, errors.New("LOGIN_NONCE is required")
	}

	return config, nil
}
