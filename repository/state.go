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

	normalizedID, err := NormalizeShopID(shopID)
	if err != nil {
		return err
	}

	if _, err := r.FirestoreClient.Collection("states").Doc(normalizedID).Set(ctx, state, firestore.MergeAll); err != nil {
		return err
	}

	return nil
}
