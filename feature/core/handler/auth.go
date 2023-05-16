package handler

import (
	"context"
	"errors"
	"firebase.google.com/go/auth"
	"github.com/aiocean/wireset/configsvc"
	model2 "github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/pubsub"
	repository2 "github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/dgraph-io/ristretto"
	"net/http"
	"strings"
	"time"

	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHandler struct {
	ShopRepo       *repository2.ShopRepository
	ShopifyService *shopifysvc.ShopifyService
	ConfigSvc      *configsvc.ConfigService
	shopifyApp     *goshopify.App
	tokenRepo      *repository2.TokenRepository
	pubsubSvc      *pubsub.Pubsub
	logSvc         *zap.Logger
	cacheSvc       *ristretto.Cache
	fireAuth       *auth.Client
}

func NewAuthHandler(
	shopRepo *repository2.ShopRepository,
	shopifySvc *shopifysvc.ShopifyService,
	configSvc *configsvc.ConfigService,
	shopifyApp *goshopify.App,
	tokenRepo *repository2.TokenRepository,
	pubsubSvc *pubsub.Pubsub,
	logSvc *zap.Logger,
	cacheSvc *ristretto.Cache,
	fiberApp *fiber.App,
	fireAuth *auth.Client,
) *AuthHandler {

	handler := &AuthHandler{
		ShopRepo:       shopRepo,
		ShopifyService: shopifySvc,
		ConfigSvc:      configSvc,
		shopifyApp:     shopifyApp,
		tokenRepo:      tokenRepo,
		pubsubSvc:      pubsubSvc,
		logSvc:         logSvc,
		cacheSvc:       cacheSvc,
		fireAuth:       fireAuth,
	}

	authGroup := fiberApp.Group("/auth")
	{
		authGroup.Get("shopify/login-callback", handler.loginCallback)
		authGroup.Get("shopify/checkin", handler.checkin)
	}

	return handler
}

func (s *AuthHandler) checkin(ctx *fiber.Ctx) error {
	// scenario 0: this request was sent with authorization header, usually from shopify app bridge
	authentication := strings.TrimPrefix(ctx.Get("authorization"), "Bearer ")

	if authentication == "" {
		// scenario 1: this request was sent with shop query parameter, usually from shopify app listing page
		shop := ctx.Query("shop")
		if shop == "" {
			return fiber.NewError(http.StatusBadRequest, "shop is required")
		}

		return ctx.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: s.shopifyApp.AuthorizeUrl(shop, s.ConfigSvc.LoginNonce),
		})
	}

	if authResponse, ok := s.cacheSvc.Get(authentication); ok {
		s.logSvc.Info("checkin cache hit")
		return ctx.Status(http.StatusOK).JSON(authResponse)
	}

	var claims model2.CustomJwtClaims
	token, err := jwt.ParseWithClaims(authentication, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.ConfigSvc.ClientSecret), nil
	})
	if err != nil {
		s.logSvc.Error("error while parsing jwt token", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
			Message: "Unauthorized",
		})
	}

	host := strings.Split(claims.Dest, "/")[2]
	authUrl := s.shopifyApp.AuthorizeUrl(host, s.ConfigSvc.LoginNonce)

	if !token.Valid {
		s.logSvc.Error("invalid jwt token")
		return ctx.Status(http.StatusOK).JSON(model2.AuthResponse{
			Message:           "Unauthorized",
			AuthenticationUrl: authUrl,
		})
	}

	shop, err := s.ShopRepo.GetByDomain(ctx.UserContext(), host)
	if err != nil {
		s.logSvc.Error("error while getting shop by domain", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
			Message:           "Shop is not found in database",
			AuthenticationUrl: authUrl,
		})
	}

	if _, err := s.tokenRepo.GetToken(ctx.UserContext(), shop.ID); err != nil {
		s.logSvc.Error("error while getting token", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
			Message:           "Token is not found in database",
			AuthenticationUrl: authUrl,
		})
	}

	accessToken, err := s.tokenRepo.GetToken(ctx.UserContext(), shop.ID)
	if err != nil {
		s.logSvc.Error("error while getting token", zap.Error(err))
		return ctx.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
			Message:           "Token is not found in database",
			AuthenticationUrl: authUrl,
		})
	}
	if err := s.pubsubSvc.Send(ctx.UserContext(), &model2.ShopCheckedInEvt{
		MyshopifyDomain: shop.MyshopifyDomain,
		AccessToken:     accessToken.AccessToken,
	}); err != nil {
		s.logSvc.Error("error while publishing event", zap.Error(err))
	}

	// This user is authorized, create custom firebase token
	firebaseToken, err := s.fireAuth.CustomToken(ctx.UserContext(), shop.ID)
	if err != nil {
		s.logSvc.Error("error while creating custom token", zap.Error(err))
		return ctx.Status(http.StatusInternalServerError).JSON(model2.AuthResponse{
			Message: "Internal server error",
		})
	}

	authResponse := model2.AuthResponse{
		Message:             "Authorized",
		FirebaseCustomToken: firebaseToken,
	}

	s.cacheSvc.SetWithTTL(authentication, authResponse, 0, time.Minute*5)
	return ctx.Status(http.StatusOK).JSON(authResponse)
}

func (s *AuthHandler) loginCallback(ctx *fiber.Ctx) error {
	if err := s.verifyLoginRequest(ctx); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	shopName := ctx.Query("shop")
	shopClient := s.ShopifyService.GetShopifyClient(shopName, accessToken)

	shopDetails, err := shopClient.GetShopDetails()
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	if exists, err := s.ShopRepo.IsShopExists(ctx.UserContext(), shopDetails.ID); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	} else if exists {
		if err := s.ShopRepo.Update(ctx.UserContext(), shopDetails); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
	} else if !exists {

		if err := s.ShopRepo.Create(ctx.UserContext(), shopDetails); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}

		s.pubsubSvc.Publish(ctx.UserContext(), &model2.ShopInstalledEvt{
			MyshopifyDomain: shopName,
			AccessToken:     accessToken,
		})
	}

	if err := s.tokenRepo.SaveAccessToken(ctx.UserContext(), &model2.ShopifyToken{
		ShopID:      shopDetails.ID,
		AccessToken: accessToken,
	}); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	redirectUrl := "https://" + shopName + "/admin/apps/" + s.ConfigSvc.ClientId
	return ctx.Redirect(redirectUrl)
}

func (s *AuthHandler) verifyLoginRequest(context *fiber.Ctx) error {
	code := context.Query("code")
	messageMAC := context.Query("hmac")
	shopDomain := context.Query("shop")
	state := context.Query("state")
	timestamp := context.Query("timestamp")
	host := context.Query("host")

	if state != s.ConfigSvc.LoginNonce {
		return errors.New("nonce is not matched")
	}

	message := "code=" + code + "&host=" + host + "&shop=" + shopDomain + "&state=" + state + "&timestamp=" + timestamp
	s.shopifyApp.VerifyMessage(message, messageMAC)
	return nil
}

func (s *AuthHandler) getAccessToken(context *fiber.Ctx) (string, error) {
	shopDomain := context.Query("shop")
	accessToken, err := s.shopifyApp.GetAccessToken(shopDomain, context.Query("code"))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// CreateFirebaseToken
func (s *AuthHandler) CreateFirebaseToken(ctx context.Context, shopID string) (string, error) {

	token, err := s.fireAuth.CustomToken(ctx, shopID)
	if err != nil {
		return "", err
	}

	return token, nil
}
