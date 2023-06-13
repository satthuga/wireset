package handler

import (
	"net/http"

	"github.com/aiocean/wireset/model"
	"github.com/gofiber/fiber/v2"
)

// getAccessToken is a helper function exchange code with access token
func (s *AuthHandler) getAccessToken(context *fiber.Ctx) (string, error) {
	shopDomain := context.Query("shop")
	accessToken, err := s.ShopifyApp.GetAccessToken(shopDomain, context.Query("code"))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// loginCallback is a handler to handle login callback from shopify
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

		s.PubsubSvc.Publish(ctx.UserContext(), &model.ShopInstalledEvt{
			MyshopifyDomain: shopName,
			AccessToken:     accessToken,
			ShopID:          shopDetails.ID,
		})
	}

	if err := s.TokenRepo.SaveAccessToken(ctx.UserContext(), &model.ShopifyToken{
		ShopID:      shopDetails.ID,
		AccessToken: accessToken,
	}); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	redirectUrl := "https://" + shopName + "/admin/apps/" + s.ShopifyConfig.ClientId

	return ctx.Redirect(redirectUrl)
}

func (s *AuthHandler) verifyLoginRequest(context *fiber.Ctx) error {
	code := context.Query("code")
	messageMAC := context.Query("hmac")
	shopDomain := context.Query("shop")
	state := context.Query("state")
	timestamp := context.Query("timestamp")
	host := context.Query("host")

	if state != s.ShopifyConfig.LoginNonce {
		return fiber.NewError(http.StatusBadRequest, "Invalid nonce")
	}

	message := "code=" + code + "&host=" + host + "&shop=" + shopDomain + "&state=" + state + "&timestamp=" + timestamp
	s.ShopifyApp.VerifyMessage(message, messageMAC)

	return nil
}
