package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/discord-notify/config"
	"github.com/aiocean/wireset/model"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type NotifyDiscordOnInstallHandler struct {
	logger     *zap.Logger
	config     *config.Config
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

// this handler used to init the wallet when user registered

func NewNotifyDiscordOnInstallHandler(
	logger *zap.Logger,
	config *config.Config,
) *NotifyDiscordOnInstallHandler {
	handler := &NotifyDiscordOnInstallHandler{
		logger: logger,
		config: config,
	}

	return handler
}

func (h *NotifyDiscordOnInstallHandler) HandlerName() string {
	return "NotifyDiscordOnInstallHandler"
}

func (h *NotifyDiscordOnInstallHandler) NewEvent() interface{} {
	return &model.ShopInstalledEvt{}
}

func (h *NotifyDiscordOnInstallHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *NotifyDiscordOnInstallHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopInstalledEvt)
	payload := strings.NewReader(`{"content": "New shop installed: ` + cmd.MyshopifyDomain + `"}`)
	req, _ := http.NewRequest("POST", h.config.NewInstallWebhook, payload)
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	return nil
}
