package fireauthsvc

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(
	NewFirebaseAuthSvc,
)

func NewFirebaseAuthSvc(
	firebaseApp *firebase.App,
) (*auth.Client, error) {
	client, err := firebaseApp.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return client, nil
}
