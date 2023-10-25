package config

import (
	"github.com/pkg/errors"
	"os"
)

type Config struct {
	NewInstallWebhook string
}

var ErrMissingNewInstallWebhook = errors.New("DISCORD_WEBHOOK_URL is missing")

func NewNotifyConfigFromEnv() (*Config, error) {
	conf := &Config{}

	if newInstallWebhook, ok := os.LookupEnv("DISCORD_WEBHOOK_URL"); ok {
		conf.NewInstallWebhook = newInstallWebhook
	} else {
		return nil, ErrMissingNewInstallWebhook
	}

	return conf, nil
}
