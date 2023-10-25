package command

import (
	"context"

	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/shopifysvc"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type InstallWebhookHandler struct {
	EventBus   *cqrs.EventBus
	CommandBus *cqrs.CommandBus
	ShopifySvc *shopifysvc.ShopifyService
}

// NewInstallWebhookHandler creates a new InstallWebhookHandler.
func NewInstallWebhookHandler(shopifySvc *shopifysvc.ShopifyService) *InstallWebhookHandler {
	return &InstallWebhookHandler{
		ShopifySvc: shopifySvc,
	}
}

func (h *InstallWebhookHandler) HandlerName() string {
	return "core.InstallWebhookCmd"
}

func (h *InstallWebhookHandler) NewCommand() interface{} {
	return &model.InstallWebhookCmd{}
}

func (h *InstallWebhookHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.EventBus = eventBus
	h.CommandBus = commandBus
}

func (h *InstallWebhookHandler) Handle(ctx context.Context, cmdItf interface{}) error {
	cmd := cmdItf.(*model.InstallWebhookCmd)

	shopClient := h.ShopifySvc.GetShopifyClient(cmd.MyshopifyDomain, cmd.AccessToken)

	if err := shopClient.InstallAppUninstalledWebhook(); err != nil {
		return err
	}

	return nil
}
