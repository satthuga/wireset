package middleware

import (
	"github.com/aiocean/wireset/configsvc"
	model2 "github.com/aiocean/wireset/model"
	repository2 "github.com/aiocean/wireset/repository"
	"github.com/aiocean/wireset/shopifysvc"
	"go.uber.org/zap"
	"net/http"
	"strings"

	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

type AuthzController struct {
	configService   *configsvc.ConfigService
	shopifyConfig   *shopifysvc.Config
	tokenRepository *repository2.TokenRepository
	shopRepository  *repository2.ShopRepository
	shopifyApp      *goshopify.App
	logger          *zap.Logger
}

func NewAuthzController(
	configSvc *configsvc.ConfigService,
	tokenRepository *repository2.TokenRepository,
	shopRepository *repository2.ShopRepository,
	shopifyApp *goshopify.App,
	logger *zap.Logger,
) *AuthzController {
	localLogger := logger.With(zap.Strings("tags", []string{"AuthzController"}))
	controller := &AuthzController{
		logger:          localLogger,
		configService:   configSvc,
		tokenRepository: tokenRepository,
		shopRepository:  shopRepository,
		shopifyApp:      shopifyApp,
	}
	return controller
}

func (s *AuthzController) IsAuthRequired(path string) bool {
	if strings.HasPrefix(path, "/auth") {
		return false
	}

	if strings.HasPrefix(path, "/ws") {
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

func (s *AuthzController) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !s.IsAuthRequired(c.OriginalURL()) {
			return c.Next()
		}

		authentication := strings.TrimPrefix(c.Get("authorization"), "Bearer ")
		if authentication == "" {
			return c.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
				Message: "Unauthorized",
			})
		}

		var claims model2.CustomJwtClaims
		token, err := jwt.ParseWithClaims(authentication, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.shopifyConfig.ClientSecret), nil
		})

		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
				Message: "Unauthorized",
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
			return c.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
				Message:           "Unauthorized",
				AuthenticationUrl: authUrl,
			})
		}

		shopifyToken, err := s.tokenRepository.GetToken(c.UserContext(), shop.ID)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(model2.AuthResponse{
				Message:           "Unauthorized",
				AuthenticationUrl: authUrl,
			})
		}

		s.logger.Debug("shopifyToken", zap.Any("shopifyToken", shopifyToken))

		c.Locals("shop", shop)
		c.Locals("shop_id", shop.ID)
		c.Locals("access_token", shopifyToken.AccessToken)
		return c.Next()

	}
}
