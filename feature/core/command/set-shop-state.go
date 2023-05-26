package command

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
)

type SetShopStateHandler struct {
	eventBus   *cqrs.EventBus
	commandBus *cqrs.CommandBus

	ShopStateRepo *repository.StateRepository
}

func NewSetShopStateHandler(
	ShopStateRepo *repository.StateRepository,
) *SetShopStateHandler {
	return &SetShopStateHandler{
		ShopStateRepo: ShopStateRepo,
	}
}

func (h *SetShopStateHandler) HandlerName() string {
	return "SetShopStateHandler"
}

func (h *SetShopStateHandler) NewCommand() interface{} {
	return &model.SetShopStateCmd{}
}

func (h *SetShopStateHandler) RegisterBus(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) {
	h.eventBus = eventBus
	h.commandBus = commandBus
}

func (h *SetShopStateHandler) Handle(ctx context.Context, raw interface{}) error {
	cmd := raw.(*model.SetShopStateCmd)

	if cmd.ShopID == "" {
		return nil
	}

	return h.ShopStateRepo.SetShopState(ctx, cmd.ShopID, cmd.State)
}
