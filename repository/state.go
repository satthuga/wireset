package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var StateRepoWireset = wire.NewSet(wire.Struct(new(StateRepository), "*"))

type StateRepository struct {
	FirestoreClient *firestore.Client
	Logger          *zap.Logger
}

// SetState set state to firestore
func (r *StateRepository) SetShopState(ctx context.Context, shopID string, state map[string]interface{}) error {
	_, err := r.FirestoreClient.Collection("states").Doc(NormalizeShopID(shopID)).Set(ctx, state, firestore.MergeAll)
	if err != nil {
		return err
	}

	return nil
}
