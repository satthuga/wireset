package ws

import (
	"github.com/aiocean/wireset/feature/realtime/models"
	models2 "github.com/aiocean/wireset/feature/shopifyapp/models"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/gofiber/contrib/websocket"
	"github.com/tidwall/gjson"
)

type CreateSubscriptionHandler struct {
	ShopifySvc *shopifysvc.ShopifyService
	TokenRepo  *repository.TokenRepository
}

func (h *CreateSubscriptionHandler) Handle(conn *websocket.Conn, payload *gjson.Result) error {
	shopifyDomain := conn.Locals("myshopifyDomain").(string)
	accessToken := conn.Locals("accessToken").(string)

	client := h.ShopifySvc.GetShopifyClient(shopifyDomain, accessToken)

	result, err := client.CreateSubscription(1)
	if err != nil {
		return conn.WriteJSON(models.WebsocketMessage{
			Topic: models.TopicError,
			Payload: models.ErrorPayload{
				Message: err.Error(),
			},
		})
	}

	confirmUrl := result.Get("appSubscriptionCreate.confirmationUrl").String()
	if confirmUrl != "" {
		return conn.WriteJSON(models.WebsocketMessage{
			Topic: models2.TopicNavigateTo,
			Payload: models2.NavigateToPayload{
				URL: confirmUrl,
			},
		})
	}

	return conn.WriteJSON(models.WebsocketMessage{
		Topic: models.TopicError,
		Payload: models.ErrorPayload{
			Message: "Failed to create subscription",
		}})
}
