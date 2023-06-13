package repository

import (
	"context"
	"github.com/pkg/errors"

	"github.com/aiocean/wireset/model"

	"cloud.google.com/go/firestore"
	"github.com/google/wire"
)

// repository get shopify token from database
type TokenRepository struct {
	firestoreClient *firestore.Client
}

func NewTokenRepository(
	firestoreClient *firestore.Client,
) *TokenRepository {
	return &TokenRepository{
		firestoreClient: firestoreClient,
	}
}

var TokenRepoWireset = wire.NewSet(NewTokenRepository)

func (r *TokenRepository) GetToken(ctx context.Context, shopID string) (*model.ShopifyToken, error) {
	if shopID == "" {
		return nil, errors.New("shop id is empty")
	}

	normalizedShopID, err := NormalizeShopID(shopID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to normalize shop id")
	}

	snapshot, err := r.firestoreClient.Collection("shops").Doc(normalizedShopID).Get(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get shop")
	}

	if !snapshot.Exists() {
		return nil, errors.New("shop not found")
	}

	tokenString, err := snapshot.DataAtPath(firestore.FieldPath{"shopifyToken"})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get shopify token")
	}

	token := model.ShopifyToken{
		ShopID:      shopID,
		AccessToken: tokenString.(string),
	}

	return &token, nil
}

func (r *TokenRepository) SaveAccessToken(ctx context.Context, token *model.ShopifyToken) error {
	if token == nil {
		return errors.New("token is nil")
	}

	updates := []firestore.Update{
		{
			Path:  "shopifyToken",
			Value: token.AccessToken,
		},
	}

	normalizedShopID, err := NormalizeShopID(token.ShopID)
	if err != nil {
		return errors.WithMessage(err, "failed to normalize shop id")
	}

	if _, err := r.firestoreClient.Collection("shops").Doc(normalizedShopID).Update(ctx, updates); err != nil {
		return errors.WithMessage(err, "failed to update shop")
	}

	return nil
}
