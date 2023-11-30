package shopifysvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/wire"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aiocean/wireset/cachesvc"
	"github.com/aiocean/wireset/configsvc"
	"github.com/pkg/errors"

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
	httpClient    *http.Client
	logger        *zap.Logger
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
		logger:        s.Logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	s.CacheSvc.SetWithTTL(cacheKey, client, 1*time.Hour)

	return &client
}
func (c *ShopifyClient) DoRestRequest(method, path string, body io.Reader) (*gjson.Result, error) {
	endpoint := fmt.Sprintf(restEndpointTemplate, c.ShopifyDomain, c.ApiVersion) + path
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, errors.Wrap(err, "DoRestRequest: failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "DoRestRequest: failed to do request")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.logger.Error("DoRestRequest: failed to close response body", zap.Error(err))
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "DoRestRequest: failed to read response body")
	}

	data := gjson.ParseBytes(respBody)

	return &data, nil
}

type GraphQlRequest struct {
	Query     string `json:"query"`
	Operation string `json:"operationName,omitempty"`
	Variables any    `json:"variables,omitempty"`
}

func (c *ShopifyClient) DoGraphqlRequest(request *GraphQlRequest) (*gjson.Result, error) {

	jsonPayload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	endpoint := fmt.Sprintf(graphQlEndpointTemplate, c.ShopifyDomain, c.ApiVersion)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("failed to close response body", zap.Error(err))
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to do request, status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	result := gjson.GetManyBytes(respBody, "data", "errors")

	if result[1].Exists() {
		return nil, &GraphQLError{Errors: result[1].Array()}
	}

	return &result[0], nil
}

func (c *ShopifyClient) GetShopDetails() (*Shop, error) {
	requestBody := `{shop{
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
        }}`

	response, err := c.DoGraphqlRequest(&GraphQlRequest{Query: requestBody})
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

	requestBody := &GraphQlRequest{
		Query: `mutation scriptTagCreate($input: ScriptTagInput!) {
            scriptTagCreate(input: $input) {
                userErrors {
                    field
                    message
                }
                scriptTag {
                    src
                }
            }
        }`,
		Operation: "scriptTagCreate",
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"cache":        false,
				"displayScope": "ALL",
				"src":          scriptUrl,
			},
		},
	}

	if _, err := c.DoGraphqlRequest(requestBody); err != nil {
		return err
	}
	return nil
}

func (c *ShopifyClient) IsScriptInstalled(scriptUrl string) (bool, error) {
	requestBody := `{
		scriptTags(first: 10, src: "` + scriptUrl + `"){
			edges{
				node {
					src
				}
			}
		}
	}`
	response, err := c.DoGraphqlRequest(&GraphQlRequest{Query: requestBody})
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

	requestBody := &GraphQlRequest{
		Query: `mutation webhookSubscriptionCreate($input: WebhookSubscriptionInput!) {
			webhookSubscriptionCreate(input: $input) {
				userErrors {
					field
					message
				}
				webhookSubscription {
					id
				}
			}
		}`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"topic":   "APP_UNINSTALLED",
				"format":  "JSON",
				"address": c.configSvc.ServiceUrl + "/webhook/shopify/app-uninstalled",
			},
		},
	}

	if _, err := c.DoGraphqlRequest(requestBody); err != nil {
		return errors.WithMessage(err, "failed to install app uninstalled webhook")
	}

	return nil
}

func (c *ShopifyClient) IsAppUninstalledWebhookInstalled() (bool, error) {
	requestBody := &GraphQlRequest{
		Query: `{
   webhookSubscriptions(first: 10, topic: APP_UNINSTALLED){
    edges{
     node {
      id
     }
    }
   }
  }`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return false, errors.WithMessage(err, "failed to check if app uninstalled webhook is installed")
	}

	total := response.Get("webhookSubscriptions.edges.#").Int()
	return total > 0, nil
}

// GetCurrentTheme returns the current theme
func (c *ShopifyClient) GetCurrentTheme() (string, error) {
	requestBody := &GraphQlRequest{
		Query: `{
		theme {
		  id
		  name
		  role
		}
	  }`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.WithMessage(err, "failed to get current theme")
	}

	themeData := response.Get("theme")
	return themeData.Get("id").String(), nil
}

func (c *ShopifyClient) GetCurrentApplicationInstallationID() (string, error) {
	requestBody := &GraphQlRequest{
		Query: `{
        currentAppInstallation {
            id
        }
    }`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to get current application installation ID")
	}

	installationID := response.Get("currentAppInstallation.id").String()
	return installationID, nil
}

const appDataMetafieldNamespace = "aio_decor"

// GetAppDataMetaField returns the value of the app data metafield
func (c *ShopifyClient) GetAppDataMetaField(ownerId, key string) (string, error) {
	requestBody := &GraphQlRequest{
		Query: `query GetAppDataMetafield($metafieldsQueryInput: MetafieldsQueryInput!) {
		  metafields(query: $metafieldsQueryInput) {
			edges {
		   node {
			 id
			 namespace
			 key
			 value
		   }
			}
		  }
   		}`,
		Operation: "GetAppDataMetafield",
		Variables: map[string]interface{}{
			"metafieldsQueryInput": map[string]interface{}{
				"namespace": appDataMetafieldNamespace,
				"key":       key,
				"ownerId":   ownerId,
			},
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to get app data metafield")
	}

	metafields := response.Get("metafields.edges.#.node")
	if metafields.Exists() {
		return metafields.Array()[0].Get("value").String(), nil
	}

	return "", nil
}
func (c *ShopifyClient) SetAppDataMetaField(ownerId, key, value string) error {
	requestBody := &GraphQlRequest{
		Query: `mutation CreateAppDataMetafield($metafieldsSetInput: [MetafieldsSetInput!]!) {
   metafieldsSet(metafields: $metafieldsSetInput) {
    metafields {
     id
     namespace
     key
     value
    }
    userErrors {
     field
     message
    }
   }
  }`,
		Operation: "CreateAppDataMetafield",
		Variables: map[string]interface{}{
			"metafieldsSetInput": []map[string]interface{}{
				{
					"namespace": appDataMetafieldNamespace,
					"key":       key,
					"type":      "single_line_text_field",
					"value":     value,
					"ownerId":   ownerId,
				},
			},
		},
	}

	_, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return errors.Wrap(err, "failed to create app data metafield")
	}

	return nil
}
