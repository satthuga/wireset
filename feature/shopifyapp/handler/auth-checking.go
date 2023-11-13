package handler

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"

	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (s *AuthHandler) Checkin(ctx *fiber.Ctx) error {
	// scenario 0: this request was sent with authorization header, usually from shopify app bridge
	authentication := strings.TrimPrefix(ctx.Get("authorization"), "Bearer ")

	if authentication == "" {
		// scenario 1: this request was sent with shop query parameter, usually from shopify app listing page
		shop := ctx.Query("shop")
		if shop == "" {
			return fiber.NewError(http.StatusBadRequest, "shop is required")
		}

		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: s.ShopifyApp.AuthorizeUrl(shop, s.ShopifyConfig.LoginNonce),
		})
	}

	if authResponse, ok := s.CacheSvc.Get(authentication); ok {
		s.LogSvc.Info("Checkin cache hit")
		return ctx.Status(http.StatusOK).JSON(authResponse)
	}

	var claims model.CustomJwtClaims
	token, err := jwt.ParseWithClaims(authentication, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.ShopifyConfig.ClientSecret), nil
	})
	if err != nil {
		s.LogSvc.Error("error while parsing jwt token", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized",
		})
	}

	host := strings.Split(claims.Dest, "/")[2]
	authUrl := s.ShopifyApp.AuthorizeUrl(host, s.ShopifyConfig.LoginNonce)

	if !token.Valid {
		s.LogSvc.Error("invalid jwt token")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: authUrl,
		})
	}

	shop, err := s.ShopRepo.GetByDomain(ctx.UserContext(), host)
	if err != nil {
		s.LogSvc.Error("error while getting shop by domain", zap.Error(err))
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Shop is not found in database",
			AuthenticationUrl: authUrl,
		})
	}

	if _, err := s.TokenRepo.GetToken(ctx.UserContext(), shop.ID); err != nil {
		s.LogSvc.Error("error while getting token", zap.Error(err))
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Token is not found in database",
			AuthenticationUrl: authUrl,
		})
	}

	accessToken, err := s.TokenRepo.GetToken(ctx.UserContext(), shop.ID)
	if err != nil {
		s.LogSvc.Error("error while getting token", zap.Error(err))
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Token is not found in database",
			AuthenticationUrl: authUrl,
		})
	}

	if err := s.EventBus.Publish(ctx.UserContext(), &model.ShopCheckedInEvt{
		MyshopifyDomain: shop.MyshopifyDomain,
		AccessToken:     accessToken.AccessToken,
		ShopID:          shop.ID,
	}); err != nil {
		s.LogSvc.Error("error while publishing event", zap.Error(err))
	}

	// This user is authorized, create custom firebase token
	normalizedId, err := repository.NormalizeShopID(shop.ID)
	if err != nil {
		s.LogSvc.Error("error while normalizing shop id", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(model.AuthResponse{
			Message: "Internal server error",
		})
	}
	firebaseToken, err := s.FireAuth.CustomToken(ctx.UserContext(), normalizedId)
	if err != nil {
		s.LogSvc.Error("error while creating custom token", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(model.AuthResponse{
			Message: "Internal server error",
		})
	}

	authResponse := model.AuthResponse{
		Message:             "Authorized",
		FirebaseCustomToken: firebaseToken,
	}

	s.CacheSvc.SetWithTTL(authentication, authResponse, time.Second*5)
	return ctx.Status(http.StatusOK).JSON(authResponse)
}

// CreateFirebaseToken creates a custom Firebase token for the given user ID.
func (s *AuthHandler) CreateFirebaseToken(ctx context.Context, shopID string) (string, error) {

	token, err := s.FireAuth.CustomToken(ctx, shopID)
	if err != nil {
		return "", err
	}

	return token, nil
}
