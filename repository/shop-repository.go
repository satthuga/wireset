package repository

import (
	"context"
	"errors"
	"github.com/aiocean/wireset/firestoresvc"
	"github.com/aiocean/wireset/shopifysvc"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/wire"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShopRepository struct {
	firestoreClient *firestore.Client
}

func NewShopRepository(
	firestoreClient *firestore.Client,
) *ShopRepository {
	return &ShopRepository{
		firestoreClient: firestoreClient,
	}
}

var ErrShopNotFound = errors.New("shop not found")

var ShopRepoWireset = wire.NewSet(
	NewShopRepository,
	firestoresvc.DefaultWireset,
)

func (r *ShopRepository) IsShopExists(ctx context.Context, shopID string) (bool, error) {
	normalizedID, err := NormalizeShopID(shopID)
	if err != nil {
		return false, err
	}
	snapshot, err := r.firestoreClient.Collection("shops").Doc(normalizedID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}

		return false, err
	}

	return snapshot.Exists(), nil
}

func (r *ShopRepository) Create(ctx context.Context, shop *shopifysvc.Shop) error {
	normalizedID, err := NormalizeShopID(shop.ID)
	if err != nil {
		return err
	}

	if _, err := r.firestoreClient.Collection("shops").Doc(normalizedID).Set(ctx, shop); err != nil {
		return err
	}

	return nil
}

func (r *ShopRepository) Update(ctx context.Context, shop *shopifysvc.Shop) error {

	updates := []firestore.Update{
		{Path: "id", Value: shop.ID},
		{Path: "domain", Value: shop.Domain},
		{Path: "myshopifyDomain", Value: shop.MyshopifyDomain},
		{Path: "name", Value: shop.Name},
		{Path: "email", Value: shop.Email},
		{Path: "countryCode", Value: shop.CountryCode},
		{Path: "timezoneAbbreviation", Value: shop.TimezoneAbbreviation},
		{Path: "ianaTimezone", Value: shop.IanaTimezone},
		{Path: "currencyCode", Value: shop.CurrencyCode},
	}

	normalizedID, err := NormalizeShopID(shop.ID)
	if err != nil {
		return err
	}

	if _, err := r.firestoreClient.Collection("shops").Doc(normalizedID).Update(ctx, updates); err != nil {
		return err
	}

	return nil
}

func (r *ShopRepository) Get(ctx context.Context, shopID string) (*shopifysvc.Shop, error) {

	normalizedID, err := NormalizeShopID(shopID)
	if err != nil {
		return nil, err
	}

	snapshot, err := r.firestoreClient.Collection("shops").Doc(normalizedID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrShopNotFound
		}
		return nil, err
	}

	shop := shopifysvc.Shop{}
	if err = snapshot.DataTo(&shop); err != nil {
		return nil, err
	}

	return &shop, nil
}

func (r *ShopRepository) GetByDomain(ctx context.Context, domain string) (*shopifysvc.Shop, error) {

	cur := r.firestoreClient.Collection("shops").Where("myshopifyDomain", "==", domain).Documents(ctx)
	defer cur.Stop()

	doc, err := cur.Next()
	if err != nil {
		return nil, err
	}

	shop := shopifysvc.Shop{}
	if err = doc.DataTo(&shop); err != nil {
		return nil, err
	}

	return &shop, nil
}

func (r *ShopRepository) UpdateLastLogin(ctx context.Context, shopID string, at *time.Time) error {
	updates := []firestore.Update{
		{Path: "lastLoginTime", Value: at},
	}

	normalizedID, err := NormalizeShopID(shopID)
	if err != nil {
		return err
	}

	if _, err := r.firestoreClient.Collection("shops").Doc(normalizedID).Update(ctx, updates); err != nil {
		return err
	}

	return nil
}

// UpdateStoreState updates the store state
func (r *ShopRepository) UpdateStoreState(ctx context.Context, shopID string, key string, value interface{}) error {
	panic("implement me")
}
