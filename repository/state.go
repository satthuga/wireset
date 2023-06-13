package repository

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var StateRepoWireset = wire.NewSet(wire.Struct(new(StateRepository), "*"))

type StateRepository struct {
	FirestoreClient *firestore.Client
	Logger          *zap.Logger
}

// SetShopState set state to firestore
func (r *StateRepository) SetShopState(ctx context.Context, shopID string, state map[string]interface{}) error {

	normalizedID, err := NormalizeShopID(shopID)
	if err != nil {
		return errors.WithMessage(err, "normalize shop id")
	}

	if _, err := r.FirestoreClient.Collection("states").Doc(normalizedID).Set(ctx, state, firestore.MergeAll); err != nil {
		return errors.WithMessage(err, "set state to firestore")
	}

	return nil
}
