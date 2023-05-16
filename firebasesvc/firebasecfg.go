package firebasesvc

import (
	"fmt"
	"os"
)

type FirebaseCfg struct {
	Credentials []byte
}

func NewFirebaseCfg() (*FirebaseCfg, error) {
	credential := os.Getenv("FIREBASE_CREDENTIAL")
	if credential == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIAL is empty")
	}

	return &FirebaseCfg{
		Credentials: []byte(credential),
	}, nil
}
