package event

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/realtime/command"
	"github.com/aiocean/wireset/feature/realtime/models"
	models2 "github.com/aiocean/wireset/feature/shopifyapp/models"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
)

type OnUserConnectedHandler struct {
	CommandBus *cqrs.CommandBus
	TokenRepo  *repository.TokenRepository
	ShopifySvc *shopifysvc.ShopifyService
}

func (h *OnUserConnectedHandler) HandlerName() string {
	return "OnUserConnectedHandler"
}

func (h *OnUserConnectedHandler) NewEvent() interface{} {
	return &models.UserJoinedEvt{}
}

func (h *OnUserConnectedHandler) Handle(ctx context.Context, event interface{}) error {
	evt := event.(*models.UserJoinedEvt)
	shopifyDomain := evt.RoomID

	return h.CommandBus.Send(ctx, &command.SendWsMessageCmd{
		RoomID:   evt.RoomID,
		Username: evt.UserName,
		Payload: models.WebsocketMessage{
			Topic: models2.TopicSetActivateSubscription,
			Payload: models2.SetActivateSubscriptionPayload{
				ID:   shopifyDomain,
				Name: evt.UserName,
			},
		},
	})
}
