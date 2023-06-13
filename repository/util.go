package repository

import (
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
)

func NormalizeShopID(shopID string) (string, error) {

	if len(shopID) == 0 {
		return "", errors.New("shopID is empty")
	}

	if strings.Count(shopID, "/") == 0 {
		return shopID, nil
	}

	return base64.StdEncoding.EncodeToString([]byte(shopID)), nil
}

func DenormalizeShopID(shopID string) (string, error) {
	if !strings.HasPrefix(shopID, "gid://") {
		return shopID, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(shopID)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode shopID")
	}

	return string(decoded), nil
}
