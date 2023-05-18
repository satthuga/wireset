package repository

import (
	"context"
	"errors"
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

	snapshot, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(shopID)).Get(ctx)
	if err != nil {
		return nil, err
	}

	if !snapshot.Exists() {
		return nil, errors.New("shop not found")
	}

	tokenString, err := snapshot.DataAtPath(firestore.FieldPath{"ShopifyToken"})
	if err != nil {
		return nil, err
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
		{Path: "ShopifyToken", Value: token.AccessToken},
	}
	_, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(token.ShopID)).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}
