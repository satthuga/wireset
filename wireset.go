package wireset

import (
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/fireauthsvc"
	"github.com/aiocean/wireset/logsvc"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/pubsub/router"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/server"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/aiocean/wireset/tracersvc"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	repository.ShopRepositoryWireset,
	repository.DefaultWireset,
	configsvc.Wireset,
	shopifysvc.Wireset,
	fiberapp.DefaultWireset,
	fireauthsvc.DefaultWireset,
	router.DefaultWireset,
	server.Wireset,
	tracersvc.TracerSvcWireset,
	logsvc.LogSvcWireset,
	pubsub.FirebasePubsubWireset,
	cachesvc.DefaultWireset,
)
