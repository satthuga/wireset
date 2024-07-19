package wireset

import (
	"github.com/google/wire"

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
)

var Common = wire.NewSet(
	fiberapp.DefaultWireset,
	server.DefaultWireset,
	logsvc.DefaultWireset,
	pubsub.DefaultWireset,
	cachesvc.DefaultWireset,
)

var ShopifyApp = wire.NewSet(
	Common,
	repository.ShopRepoWireset,
	repository.TokenRepoWireset,
	repository.StateRepoWireset,
	shopifysvc.DefaultWireset,
	firestoresvc.DefaultWireset,
	fireauthsvc.DefaultWireset,
	shopifyapp.DefaultWireset,
)

var NormalApp = wire.NewSet(
	Common,
	fireauthsvc.DefaultWireset,
	firestoresvc.DefaultWireset,
)

// MinimalApp provides minimal dependencies for a basic app
var MinimalApp = Common

// CliApp provides dependencies for a CLI app
var CliApp = wire.NewSet(
	logsvc.DefaultWireset,
)
