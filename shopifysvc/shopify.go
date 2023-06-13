package shopifysvc

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/pkg/errors"

	"github.com/google/wire"
	"github.com/tidwall/gjson"
)

const graphQlEndpointTemplate = "https://%s.myshopify.com/admin/api/%s/graphql.json"
const restEndpointTemplate = "https://%s.myshopify.com/admin/api/%s"

var DefaultWireset = wire.NewSet(
	NewShopifyService,
	NewShopifyApp,
)

type ShopifyService struct {
	ConfigService *configsvc.ConfigService
	ShopifyConfig *Config
	CacheSvc      *cachesvc.CacheService
	Logger        *zap.Logger
}

func NewShopifyService(
	configService *configsvc.ConfigService,
	shopifyConfig *Config,
	cacheSvc *cachesvc.CacheService,
	logger *zap.Logger,
) (*ShopifyService, func(), error) {
	cleanup := func() {

	}

	return &ShopifyService{
		ConfigService: configService,
		ShopifyConfig: shopifyConfig,
		CacheSvc:      cacheSvc,
		Logger:        logger.With(zap.Strings("tags", []string{"shopify"})),
	}, cleanup, nil
}

type ShopifyClient struct {
	ShopifyDomain string
	ApiVersion    string
	AccessToken   string
	configSvc     *configsvc.ConfigService
}

func (s *ShopifyService) GetShopifyClient(shop, accessToken string) *ShopifyClient {
	shop = strings.Replace(shop, ".myshopify.com", "", -1)
	cacheKey := fmt.Sprintf("shopify_client_%s", shop)
	if client, ok := s.CacheSvc.Get(cacheKey); ok {
		c := client.(ShopifyClient)
		return &c
	}

	client := ShopifyClient{
		ShopifyDomain: shop,
		AccessToken:   accessToken,
		ApiVersion:    s.ShopifyConfig.ApiVersion,
		configSvc:     s.ConfigService,
	}
	s.CacheSvc.SetWithTTL(cacheKey, client, 1*time.Hour)

	return &client
}

// restRequest
func (c *ShopifyClient) DoRestRequest(method, path string, body io.Reader) (*gjson.Result, error) {
	endpoint := fmt.Sprintf(restEndpointTemplate, c.ShopifyDomain, c.ApiVersion) + path
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := gjson.ParseBytes(respBody)

	return &data, nil
}

func (c *ShopifyClient) DoGraphqlRequest(body string) (*gjson.Result, error) {

	endpoint := fmt.Sprintf(graphQlEndpointTemplate, c.ShopifyDomain, c.ApiVersion)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/graphql")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := gjson.GetBytes(respBody, "data")

	return &data, nil
}

func (c *ShopifyClient) GetShopDetails() (*Shop, error) {
	requestBody := `
{
  shop{
	id
    name
    email
    ianaTimezone
    timezoneOffset
	currencyCode
	myshopifyDomain
	primaryDomain {
      host
    }
  }
}`
	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get shop details")
	}

	shopData := response.Get("shop")
	shopDetails := &Shop{
		ID:                   shopData.Get("id").String(),
		Name:                 shopData.Get("name").String(),
		Email:                shopData.Get("email").String(),
		CountryCode:          shopData.Get("countryCode").String(),
		Domain:               shopData.Get("primaryDomain.host").String(),
		MyshopifyDomain:      shopData.Get("myshopifyDomain").String(),
		TimezoneAbbreviation: shopData.Get("timezoneOffset").String(),
		IanaTimezone:         shopData.Get("ianaTimezone").String(),
		CurrencyCode:         shopData.Get("currencyCode").String(),
	}

	return shopDetails, nil
}

func (c *ShopifyClient) InstallScript(scriptUrl string) error {
	isInstalled, err := c.IsScriptInstalled(scriptUrl)
	if err != nil {
		return err
	}

	if isInstalled {
		return nil
	}

	requestBody := `
{
	"query":"mutation scriptTagCreate($input: ScriptTagInput!) {
		scriptTagCreate(input: $input) {
	userErrors {
		field
		message
	}
	scriptTag {
		src
	}
}}","variables":{"input":{"cache":false,"displayScope":"ALL","src":"` + scriptUrl + `"}},"operationName":"scriptTagCreate"}'
`
	if _, err := c.DoGraphqlRequest(requestBody); err != nil {
		return err
	}
	return nil
}

func (c *ShopifyClient) IsScriptInstalled(scriptUrl string) (bool, error) {
	requestBody := `{"query":"{
		scriptTags(first: 10, src: "` + scriptUrl + `"){
			edges{
				node {
					src
				}
			}
		}
	}"}`
	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return false, err
	}

	total := response.Get("scriptTags.edges.#").Int()
	return total > 0, nil
}

func (c *ShopifyClient) InstallAppUninstalledWebhook() error {
	isInstalled, err := c.IsAppUninstalledWebhookInstalled()
	if err != nil {
		return err
	}

	if isInstalled {
		return nil
	}

	requestBody := `{"query":"mutation webhookSubscriptionCreate($input: WebhookSubscriptionInput!) {
							webhookSubscriptionCreate(input: $input) {
								userErrors {
									field
									message
								}
								webhookSubscription {
									id
								}
							}}","variables":{"input":{"topic":"APP_UNINSTALLED","format":"JSON","address":"` + c.configSvc.ServiceUrl + `/webhook/shopify/app-uninstalled"}},"operationName":"webhookSubscriptionCreate"}'`
	if _, err := c.DoGraphqlRequest(requestBody); err != nil {
		return errors.WithMessage(err, "failed to install app uninstalled webhook")
	}

	return nil
}

func (c *ShopifyClient) IsAppUninstalledWebhookInstalled() (bool, error) {
	requestBody := `{"query":"{
		webhookSubscriptions(first: 10, topic: APP_UNINSTALLED){
			edges{
				node {
					id
				}
			}
		}
	}"}`
	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return false, errors.WithMessage(err, "failed to check if app uninstalled webhook is installed")
	}

	total := response.Get("webhookSubscriptions.edges.#").Int()
	return total > 0, nil
}

// GetCurrentTheme returns the current theme
func (c *ShopifyClient) GetCurrentTheme() (string, error) {
	requestBody := `
query {
  theme {
    id
    name
    role
  }
}`
	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get current theme")
	}

	themeData := response.Get("theme")
	return themeData.Get("id").String(), nil
}
