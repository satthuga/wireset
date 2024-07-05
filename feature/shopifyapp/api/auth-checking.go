package api

import (
	"github.com/aiocean/wireset/shopifysvc"
	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aiocean/wireset/model"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (s *AuthHandler) sessionTokenKeyFunc(token *jwt.Token) (interface{}, error) {
	return []byte(s.ShopifyConfig.ClientSecret), nil
}

func (s *AuthHandler) Checkin(ctx *fiber.Ctx) error {
	authentication := strings.TrimPrefix(ctx.Get("authorization"), "Bearer ")

	if authentication == "" {
		// this case is rare, but it's possible
		// Happen when user access the app directly, or from development environment
		return s.handleNoAuth(ctx)
	}

	return s.handleAuth(ctx, authentication)
}

func (s *AuthHandler) handleNoAuth(ctx *fiber.Ctx) error {
	s.LogSvc.Info("authentication header is empty")
	shopQuery := ctx.Query("shop")
	if shopQuery == "" {
		s.LogSvc.Info("shop query parameter is empty")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: s.ShopifyConfig.AppListingUrl,
		})
	}

	shopName := strings.TrimSuffix(shopQuery, ".myshopify.com")
	shopDomain := shopName + ".myshopify.com"

	isExists, err := s.ShopRepo.IsDomainExists(ctx.UserContext(), shopDomain)
	if err != nil {
		s.LogSvc.Error("error while checking shop domain", zap.Error(err))
	}

	if isExists {
		s.LogSvc.Info("Shop exists in the database")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: "https://admin.shopify.com/store/" + shopName + "/apps/" + s.ShopifyConfig.ClientId,
		})
	}

	return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
		Message:           "Unauthorized",
		AuthenticationUrl: authorizeUrl(shopName, s.ShopifyConfig),
	})
}

func authorizeUrl(shopName string, shopifyConfig *shopifysvc.Config) string {
	shopUrl, _ := url.Parse(goshopify.ShopBaseUrl(shopName))
	shopUrl.Path = "/admin/oauth/authorize"
	query := shopUrl.Query()
	query.Set("client_id", shopifyConfig.ClientId)
	query.Set("state", shopifyConfig.LoginNonce)
	shopUrl.RawQuery = query.Encode()
	return shopUrl.String()
}

func (s *AuthHandler) handleAuth(ctx *fiber.Ctx, authentication string) error {
	if authResponse, ok := s.CacheSvc.Get(authentication); ok {
		return ctx.Status(http.StatusOK).JSON(authResponse)
	}

	var sessionClaim model.CustomJwtClaims
	sessionToken, err := jwt.ParseWithClaims(authentication, &sessionClaim, s.sessionTokenKeyFunc)
	if err != nil {
		s.LogSvc.Error("error parsing jwt sessionToken", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized",
		})
	}

	if !sessionToken.Valid {
		s.LogSvc.Error("invalid jwt sessionToken")
		return ctx.Status(http.StatusOK).JSON(model.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: s.ShopifyConfig.AppListingUrl,
		})
	}

	parsedMyshopifyDomain := strings.Split(sessionClaim.Dest, "/")[2]

	if err := s.EventBus.Publish(ctx.UserContext(), &model.ShopCheckedInEvt{
		MyshopifyDomain: parsedMyshopifyDomain,
		SessionToken:    authentication,
	}); err != nil {
		s.LogSvc.Error("error publishing event", zap.Error(err))
	}

	authResponse := model.AuthResponse{
		Message: "Authorized",
	}

	ttl := int64(sessionClaim.Exp) - time.Now().Unix()
	ttlDuration := time.Duration(ttl) * time.Second
	s.CacheSvc.SetWithTTL(authentication, authResponse, ttlDuration)

	return ctx.Status(http.StatusOK).JSON(authResponse)
}
