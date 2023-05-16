package pkg

import (
	"api/pkg/cachesvc"
	"api/pkg/configsvc"
	"api/pkg/fiberapp"
	"api/pkg/fireauthsvc"
	"api/pkg/logsvc"
	"api/pkg/pubsub"
	"api/pkg/pubsub/router"
	"api/pkg/repository"
	"api/pkg/server"
	"api/pkg/shopifysvc"
	"api/pkg/tracersvc"
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
