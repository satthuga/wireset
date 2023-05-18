package shopifysvc

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/tidwall/gjson"
)

const endpointTemplate = "https://%s.myshopify.com/admin/api/%s/graphql.json"

var DefaultWireset = wire.NewSet(
	wire.Struct(new(ShopifyService), "*"),
	NewShopifyApp,
)

type ShopifyService struct {
	ConfigService *configsvc.ConfigService
	ShopifyConfig *Config
	CacheSvc      *cachesvc.CacheService
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

func (c *ShopifyClient) doRequest(body string) (*gjson.Result, error) {

	url := fmt.Sprintf(endpointTemplate, c.ShopifyDomain, c.ApiVersion)

	var jsonStr = []byte(body)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/graphql")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := gjson.GetBytes(respBody, "data")

	return &data, nil
}

func (c *ShopifyClient) GetShopDetails() (*Shop, error) {
	requestBody := `{
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
	response, err := c.doRequest(requestBody)
	if err != nil {
		return nil, err
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

	requestBody := `{"query":"mutation scriptTagCreate($input: ScriptTagInput!) {
							scriptTagCreate(input: $input) {
								userErrors {
									field
									message
								}
								scriptTag {
									src
								}
							}}","variables":{"input":{"cache":false,"displayScope":"ALL","src":"` + scriptUrl + `"}},"operationName":"scriptTagCreate"}'`
	if _, err := c.doRequest(requestBody); err != nil {
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
	response, err := c.doRequest(requestBody)
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
	if _, err := c.doRequest(requestBody); err != nil {
		return err
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
	response, err := c.doRequest(requestBody)
	if err != nil {
		return false, err
	}

	total := response.Get("webhookSubscriptions.edges.#").Int()
	return total > 0, nil
}

var ErrProductNotFound = errors.New("product not found")

// GetProductByHandle returns a product by handle
func (c *ShopifyClient) GetProductByHandle(handle string) (*Product, error) {
	panic("implement me")
}
