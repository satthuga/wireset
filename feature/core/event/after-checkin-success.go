package event

import (
	"context"
	model2 "github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/pubsub"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"go.uber.org/zap"
)

type CheckinHandler struct {
	logger     *zap.Logger
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus
}

// this handler used to init the wallet when user registered

func NewCheckinHandler(
	logger *zap.Logger,
	registry *pubsub.HandlerRegistry,
) *CheckinHandler {
	handler := &CheckinHandler{
		logger: logger,
	}

	registry.AddEventHandler(handler)

	return handler
}

func (h *CheckinHandler) HandlerName() string {
	return "CheckinHandler"
}

func (h *CheckinHandler) NewEvent() interface{} {
	return &model2.ShopCheckedInEvt{}
}

func (h *CheckinHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *CheckinHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model2.ShopCheckedInEvt)
	h.logger.Info("CoreCheckinHandler", zap.Any("event", event))

	return h.commandBus.Send(ctx, &model2.InstallWebhookCmd{
		MyshopifyDomain: cmd.MyshopifyDomain,
		AccessToken:     cmd.AccessToken,
	})
}
