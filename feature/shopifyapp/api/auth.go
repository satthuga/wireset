package api

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	goshopify "github.com/bold-commerce/go-shopify/v3"
	"go.uber.org/zap"
)

type AuthHandler struct {
	ShopRepo       *repository.ShopRepository
	ShopifyService *shopifysvc.ShopifyService
	ConfigSvc      *configsvc.ConfigService
	ShopifyConfig  *shopifysvc.Config
	ShopifyApp     *goshopify.App
	TokenRepo      *repository.TokenRepository
	EventBus       *cqrs.EventBus
	CommandBus     *cqrs.CommandBus
	LogSvc         *zap.Logger
	CacheSvc       *cachesvc.CacheService
}
