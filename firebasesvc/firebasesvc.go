package firebasesvc

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"github.com/google/wire"
	"google.golang.org/api/option"
)

var DefaultWireset = wire.NewSet(
	NewFirebaseSvc,
	NewFirebaseCfg,
)

func NewFirebaseSvc(cfg *FirebaseCfg) (*firebase.App, error) {
	opt := option.WithCredentialsJSON(cfg.Credentials)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	return app, nil
}
