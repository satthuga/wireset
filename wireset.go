package wireset

import (
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/feature/shopifyapp"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/fireauthsvc"
	"github.com/aiocean/wireset/firestoresvc"
	"github.com/aiocean/wireset/logsvc"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/server"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/aiocean/wireset/tracersvc"
	"github.com/google/wire"
)

// ShopifyAppWireset is a wire set that provides dependencies for a Shopify app.
// It includes dependencies for repositories, Shopify service, Fiber app, Firebase authentication service,
// server, tracing service, logging service, pubsub service, and cache service.
var ShopifyAppWireset = wire.NewSet(
	repository.ShopRepoWireset,
	repository.TokenRepoWireset,
	repository.StateRepoWireset,
	shopifysvc.DefaultWireset,
	fiberapp.DefaultWireset,
	firestoresvc.DefaultWireset,
	fireauthsvc.DefaultWireset,
	server.DefaultWireset,
	tracersvc.TracerSvcWireset,
	logsvc.DefaultWireset,
	pubsub.DefaultWireset,
	cachesvc.DefaultWireset,
	shopifyapp.DefaultWireset,
)

// NormalAppWireset is a wire set that provides dependencies for a normal app.
// It includes dependencies for Fiber app, Firebase authentication service, server,
// tracing service, logging service, pubsub service, and cache service.
var NormalAppWireset = wire.NewSet(
	fiberapp.DefaultWireset,
	fireauthsvc.DefaultWireset,
	firestoresvc.DefaultWireset,
	server.DefaultWireset,
	tracersvc.TracerSvcWireset,
	logsvc.DefaultWireset,
	pubsub.DefaultWireset,
	cachesvc.DefaultWireset,
)
