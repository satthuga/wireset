package command

import (
	"api/pkg/model"
	"api/pkg/pubsub"
	"api/pkg/shopifysvc"
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type InstallWebhookHandler struct {
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
	shopifySvc *shopifysvc.ShopifyService
}

func NewInstallWebhookHandler(
	shopifySvc *shopifysvc.ShopifyService,
	registry *pubsub.HandlerRegistry,

) *InstallWebhookHandler {
	handler := &InstallWebhookHandler{
		shopifySvc: shopifySvc,
	}

	registry.AddCommandHandler(handler)

	return handler
}

func (h *InstallWebhookHandler) HandlerName() string {
	return "core.InstallWebhookCmd"
}

func (h *InstallWebhookHandler) NewCommand() interface{} {
	return &model.InstallWebhookCmd{}
}

func (h *InstallWebhookHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *InstallWebhookHandler) Handle(ctx context.Context, cmdItf interface{}) error {
	cmd := cmdItf.(*model.InstallWebhookCmd)

	shopClient := h.shopifySvc.GetShopifyClient(cmd.MyshopifyDomain, cmd.AccessToken)

	if err := shopClient.InstallAppUninstalledWebhook(); err != nil {
		return err
	}

	return nil
}
