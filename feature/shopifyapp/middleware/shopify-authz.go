package middleware

import (
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/model"
	"github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"net/http"
	"strings"

	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/gofiber/fiber/v2"
)

type ShopifyAuthzMiddleware struct {
	configService   *configsvc.ConfigService
	shopifyConfig   *shopifysvc.Config
	tokenRepository *repository.TokenRepository
	shopRepository  *repository.ShopRepository
	cacheSvc        *cachesvc.CacheService
	shopifyApp      *goshopify.App
	logger          *zap.Logger
}

func NewAuthzController(
	configSvc *configsvc.ConfigService,
	shopifyConfig *shopifysvc.Config,
	tokenRepository *repository.TokenRepository,
	shopRepository *repository.ShopRepository,
	shopifyApp *goshopify.App,
	logger *zap.Logger,
	cacheSvc *cachesvc.CacheService,
) *ShopifyAuthzMiddleware {
	localLogger := logger.Named("shopifyAuthzMiddleware")
	controller := &ShopifyAuthzMiddleware{
		logger:          localLogger,
		configService:   configSvc,
		shopifyConfig:   shopifyConfig,
		tokenRepository: tokenRepository,
		shopRepository:  shopRepository,
		shopifyApp:      shopifyApp,
		cacheSvc:        cacheSvc,
	}

	return controller
}

// IsAuthRequired check if the path is required authentication
// TODO by hard code this path, it's become not flexible, hard to maintain. Maybe let's feature to register it into the http handler registry is better
func (s *ShopifyAuthzMiddleware) IsAuthRequired(path string) bool {
	if strings.HasPrefix(path, "/auth") {
		return false
	}

	if strings.HasPrefix(path, "/metrics") {
		return false
	}

	if strings.HasPrefix(path, "/app") {
		return false
	}

	return true
}

// TODO the token which sent from shopify have expired time, we can use this time to cache the authz result, so that we do not need to query database every time
func (s *ShopifyAuthzMiddleware) Handle(c *fiber.Ctx) error {
	if !s.IsAuthRequired(c.OriginalURL()) {
		return c.Next()
	}

	authentication := strings.TrimPrefix(c.Get("authorization"), "Bearer ")
	if authentication == "" {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized",
		})
	}

	var claims model.CustomJwtClaims
	token, err := jwt.ParseWithClaims(authentication, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.shopifyConfig.ClientSecret), nil
	})
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	if !token.Valid {
		return fiber.NewError(http.StatusUnauthorized, "Couldn't handle this token")
	}

	adminUrl := claims.Dest
	host := strings.Split(adminUrl, "/")[2]

	authUrl := s.shopifyApp.AuthorizeUrl(host, s.shopifyConfig.LoginNonce)

	shop, err := s.shopRepository.GetByDomain(c.UserContext(), host)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message:           "Unauthorized: " + err.Error(),
			AuthenticationUrl: authUrl,
		})
	}

	if shop == nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message:           "Unauthorized: " + err.Error(),
			AuthenticationUrl: authUrl,
		})
	}

	shopifyToken, err := s.tokenRepository.GetToken(c.UserContext(), shop.ID)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(model.AuthResponse{
			Message:           "Unauthorized: " + err.Error(),
			AuthenticationUrl: authUrl,
		})
	}

	c.Locals("shop", shop)
	c.Locals("shop_id", shop.ID)
	c.Locals("access_token", shopifyToken.AccessToken)

	return c.Next()

}
