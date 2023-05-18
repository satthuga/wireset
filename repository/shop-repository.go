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
	snapshot, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(shopID)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}

		return false, err
	}

	return snapshot.Exists(), nil
}

func (r *ShopRepository) Create(ctx context.Context, shop *shopifysvc.Shop) error {
	_, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(shop.ID)).Set(ctx, shop)
	if err != nil {
		return err
	}

	return nil
}

func (r *ShopRepository) Update(ctx context.Context, shop *shopifysvc.Shop) error {

	updates := []firestore.Update{
		{Path: "ID", Value: shop.ID},
		{Path: "Domain", Value: shop.Domain},
		{Path: "MyshopifyDomain", Value: shop.MyshopifyDomain},
		{Path: "Name", Value: shop.Name},
		{Path: "Email", Value: shop.Email},
		{Path: "CountryCode", Value: shop.CountryCode},
		{Path: "TimezoneAbbreviation", Value: shop.TimezoneAbbreviation},
		{Path: "IanaTimezone", Value: shop.IanaTimezone},
		{Path: "CurrencyCode", Value: shop.CurrencyCode},
	}
	_, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(shop.ID)).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

func (r *ShopRepository) Get(ctx context.Context, shopID string) (*shopifysvc.Shop, error) {
	snapshot, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(shopID)).Get(ctx)
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

	cur := r.firestoreClient.Collection("shops").Where("MyshopifyDomain", "==", domain).Documents(ctx)
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
		{Path: "LastLoginTime", Value: at},
	}
	_, err := r.firestoreClient.Collection("shops").Doc(normalizeShopID(shopID)).Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStoreState updates the store state
func (r *ShopRepository) UpdateStoreState(ctx context.Context, shopID string, key string, value interface{}) error {
	panic("implement me")
}
