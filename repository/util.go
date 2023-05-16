package repository

import "encoding/base64"

func normalizeShopID(shopID string) string {
	// convert to base64 string
	return base64.StdEncoding.EncodeToString([]byte(shopID))
}

func denormalizeShopID(shopID string) (string, error) {
	// decode base64 string
	decoded, err := base64.StdEncoding.DecodeString(shopID)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
