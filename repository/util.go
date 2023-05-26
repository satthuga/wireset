package repository

import (
	"encoding/base64"
	"strings"
)

func NormalizeShopID(shopID string) string {
	if _, err := base64.StdEncoding.DecodeString(shopID); err == nil {
		return shopID
	}

	return base64.StdEncoding.EncodeToString([]byte(shopID))
}

func DenormalizeShopID(shopID string) (string, error) {
	if !strings.HasPrefix(shopID, "gid://") {
		return shopID, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(shopID)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
