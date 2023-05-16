package firestoresvc

import (
	"api/pkg/firebasesvc"
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var DefaultWireset = wire.NewSet(
	NewFirestoreSvc,
	firebasesvc.DefaultWireset,
)

func NewFirestoreSvc(
	app *firebase.App,
	logger *zap.Logger,
) (*firestore.Client, func(), error) {
	ctx := context.Background()

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error initializing app: %v", err)
	}

	localLogger := logger.With(zap.Strings("tags", []string{"FirestoreSvc"}))

	cleanup := func() {
		localLogger.Info("FirestoreSvc: Cleaning up")
		if err := client.Close(); err != nil {
			localLogger.Error("FirestoreSvc: error closing firestore client", zap.Error(err))
		} else {
			localLogger.Info("FirestoreSvc: firestore client closed")
		}

	}

	return client, cleanup, nil
}
