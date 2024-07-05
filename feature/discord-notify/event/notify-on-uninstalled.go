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

type NotifyDiscordOnUninstallHandler struct {
	logger     *zap.Logger
	config     *config.Config
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

func NewNotifyDiscordOnUninstallHandler(
	logger *zap.Logger,
	config *config.Config,
) *NotifyDiscordOnUninstallHandler {
	handler := &NotifyDiscordOnUninstallHandler{
		logger: logger,
		config: config,
	}

	return handler
}

func (h *NotifyDiscordOnUninstallHandler) HandlerName() string {
	return "NotifyDiscordOnUninstallHandler"
}

func (h *NotifyDiscordOnUninstallHandler) NewEvent() interface{} {
	return &model.ShopUninstalledEvt{}
}

func (h *NotifyDiscordOnUninstallHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *NotifyDiscordOnUninstallHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopUninstalledEvt)
	payload := strings.NewReader(`{"content": "Shop uninstalled: ` + cmd.MyshopifyDomain + `"}`)
	req, _ := http.NewRequest("POST", h.config.NewInstallWebhook, payload)
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	return nil
}
