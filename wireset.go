package wireset

import (
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/fireauthsvc"
	"github.com/aiocean/wireset/logsvc"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/server"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/aiocean/wireset/tracersvc"
	"github.com/google/wire"
)

var ShopifyAppWireset = wire.NewSet(
	repository.ShopRepoWireset,
	repository.TokenRepoWireset,
	shopifysvc.DefaultWireset,
	fiberapp.DefaultWireset,
	fireauthsvc.DefaultWireset,
	server.DefaultWireset,
	tracersvc.TracerSvcWireset,
	logsvc.DefaultWireset,
	pubsub.DefaultWireset,
	cachesvc.DefaultWireset,
)
var NormalAppWireset = wire.NewSet(
	fiberapp.DefaultWireset,
	fireauthsvc.DefaultWireset,
	server.DefaultWireset,
	tracersvc.TracerSvcWireset,
	logsvc.DefaultWireset,
	pubsub.DefaultWireset,
	cachesvc.DefaultWireset,
)
