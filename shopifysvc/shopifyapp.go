package shopifysvc

import (
	"github.com/aiocean/wireset/configsvc"
	goshopify "github.com/bold-commerce/go-shopify/v3"
)

func NewShopifyApp(shopifyConfig *Config, config *configsvc.ConfigService) *goshopify.App {
	return &goshopify.App{
		ApiKey:      shopifyConfig.ClientId,
		ApiSecret:   shopifyConfig.ClientSecret,
		RedirectUrl: config.ServiceUrl + "/auth/shopify/login-callback",
	}
}
