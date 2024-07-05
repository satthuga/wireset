package ws

import (
	"github.com/aiocean/wireset/feature/realtime/models"
	models2 "github.com/aiocean/wireset/feature/shopifyapp/models"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/gofiber/contrib/websocket"
	"github.com/tidwall/gjson"
)

type FetchActivateSubscriptionHandler struct {
	ShopifySvc *shopifysvc.ShopifyService
}

func (h *FetchActivateSubscriptionHandler) Handle(conn *websocket.Conn, payload *gjson.Result) error {

	shopifyDomain := conn.Locals("myshopifyDomain").(string)
	accessToken := conn.Locals("accessToken").(string)

	client := h.ShopifySvc.GetShopifyClient(shopifyDomain, accessToken)

	currentSubscription, err := client.GetSubscription()
	if err != nil {
		return conn.WriteJSON(models.WebsocketMessage{
			Topic: models.TopicError,
			Payload: models.ErrorPayload{
				Message: err.Error(),
			},
		})
	}

	return conn.WriteJSON(models.WebsocketMessage{
		Topic: models2.TopicSetActivateSubscription,
		Payload: models2.SetActivateSubscriptionPayload{
			ID:     currentSubscription.ID,
			Status: currentSubscription.Status,
		},
	})
}
