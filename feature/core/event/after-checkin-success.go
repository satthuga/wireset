package event

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/pubsub"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"go.uber.org/zap"
)

type CheckinHandler struct {
	Logger          *zap.Logger
	EventBus        *cqrs.EventBus
	CommandBus      *cqrs.CommandBus
	Registry        *pubsub.HandlerRegistry
	FireStoreClient *firestore.Client
}

func NewCheckinHandler(
	logger *zap.Logger,
	registry *pubsub.HandlerRegistry,
) *CheckinHandler {

	return &CheckinHandler{
		Logger:   logger,
		Registry: registry,
	}
}

func (h *CheckinHandler) Init() {
	h.Registry.AddEventHandler(h)
}

func (h *CheckinHandler) HandlerName() string {
	return "core.CheckinHandler"
}

func (h *CheckinHandler) NewEvent() interface{} {
	return &model.ShopCheckedInEvt{}
}

func (h *CheckinHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.EventBus = eventBus
	h.CommandBus = commandBus
}

func (h *CheckinHandler) Handle(ctx context.Context, event interface{}) error {
	cmd := event.(*model.ShopCheckedInEvt)

	_, _, err := h.FireStoreClient.Collection("todos").Add(ctx, map[string]interface{}{
		"first": "Ada",
		"last":  "Lovelace",
		"born":  1815,
	})
	if err != nil {
		return err
	}

	h.Logger.Info("CoreCheckinHandler", zap.Any("event", event))

	return h.CommandBus.Send(ctx, &model.InstallWebhookCmd{
		MyshopifyDomain: cmd.MyshopifyDomain,
		AccessToken:     cmd.AccessToken,
	})
}
