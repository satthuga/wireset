package firebasesvc

import (
	"encoding/base64"
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

	//credential is base64 encoded

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(credential)))
	_, err := base64.StdEncoding.Decode(decoded, []byte(credential))
	if err != nil {
		return nil, fmt.Errorf("failed to decode FIREBASE_CREDENTIAL: %w", err)
	}

	return &FirebaseCfg{
		Credentials: decoded,
	}, nil
}
