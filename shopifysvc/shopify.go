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
	ShopifyConfig *Config
	ApiVersion    string
	AccessToken   string
	configSvc     *configsvc.ConfigService
	httpClient    *http.Client
	logger        *zap.Logger
}

func (s *ShopifyService) GetShopifyClient(shop, accessToken string) *ShopifyClient {
	shop = strings.Replace(shop, ".myshopify.com", "", -1)
	cacheKey := fmt.Sprintf("shopify_client_%s_%s", shop, accessToken)
	if client, ok := s.CacheSvc.Get(cacheKey); ok {
		c := client.(ShopifyClient)
		return &c
	}

	client := ShopifyClient{
		ShopifyDomain: shop,
		AccessToken:   accessToken,
		ApiVersion:    s.ShopifyConfig.ApiVersion,
		configSvc:     s.ConfigService,
		ShopifyConfig: s.ShopifyConfig,
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
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
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
		OperationName: "scriptTagCreate",
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
		Query: `query GetAppDataMetafield($metafieldsQueryInput: [MetafieldsQueryInput!]!) {
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
		OperationName: "GetAppDataMetafield",
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

// GetShopMetaField accept ownerId, key
func (c *ShopifyClient) GetShopMetaField(namespace, key string) (string, error) {
	requestBody := &GraphQlRequest{
		Query: `query GetShopMetafield($namespace: String!, $key: String!) {
			shop {
				metafield(namespace: $namespace, key: $key) {
					value
				}
			}
		}`,
		Variables: map[string]interface{}{
			"namespace": namespace,
			"key":       key,
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to get shop metafield")
	}

	metafield := response.Get("shop.metafield")
	if metafield.Exists() {
		return metafield.Get("value").String(), nil
	}

	return "", nil
}

func (c *ShopifyClient) SetShopMetaField(ownerId, key, value string) error {
	requestBody := &GraphQlRequest{
		Query: `mutation CreateShopMetafield($metafieldsSetInput: [MetafieldsSetInput!]!) {
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
		OperationName: "CreateShopMetafield",
		Variables: map[string]interface{}{
			"metafieldsSetInput": []map[string]interface{}{
				{
					"namespace": "aio_decor",
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
		return errors.Wrap(err, "failed to create shop metafield")
	}

	return nil
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
		OperationName: "CreateAppDataMetafield",
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

type Subscription struct {
	ID                        string
	TrialDays                 int
	CurrentPeriodEnd          string
	Status                    string
	Test                      bool
	CurrentPeriodEndFormatted string
}

var ErrorSubscriptionNotFound = errors.New("subscription not found")

func (c *ShopifyClient) GetSubscription() (*Subscription, error) {
	requestBody := &GraphQlRequest{
		Query: `{
		  currentAppInstallation {
			activeSubscriptions{
				id
				name
				trialDays
				status
				test
				currentPeriodEnd
			}
		  }
		}`,
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get subscription")
	}

	subscriptionData := response.Get("currentAppInstallation.activeSubscriptions.0")
	if !subscriptionData.Exists() {
		return nil, ErrorSubscriptionNotFound
	}

	subscription := &Subscription{
		ID:               subscriptionData.Get("id").String(),
		TrialDays:        int(subscriptionData.Get("trialDays").Int()),
		CurrentPeriodEnd: subscriptionData.Get("currentPeriodEnd").String(),
		Status:           subscriptionData.Get("status").String(),
	}

	return subscription, nil
}

func (c *ShopifyClient) CreateSubscription(price float32) (*gjson.Result, error) {
	name := "premium"
	returnUrl := "https://admin.shopify.com/store/" + c.ShopifyDomain + "/apps/" + c.ShopifyConfig.ClientId
	lineItems := []map[string]interface{}{
		{
			"plan": map[string]interface{}{
				"appRecurringPricingDetails": map[string]interface{}{
					"price": map[string]interface{}{
						"amount":       price,
						"currencyCode": "USD",
					},
					"interval": "EVERY_30_DAYS",
				},
			},
		},
	}

	requestBody := &GraphQlRequest{
		Query: `mutation AppSubscriptionCreate(
            $name: String!
            $lineItems: [AppSubscriptionLineItemInput!]!
            $returnUrl: URL!
        ) {
            appSubscriptionCreate(
                name: $name
                returnUrl: $returnUrl
                lineItems: $lineItems
                test: true
            ) {
                userErrors {
                    field
                    message
                }
                appSubscription {
                    id
                }
                confirmationUrl
            }
        }`,
		Variables: map[string]interface{}{
			"name":      name,
			"returnUrl": returnUrl,
			"lineItems": lineItems,
		},
	}

	response, err := c.DoGraphqlRequest(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create subscription")
	}

	return response, nil
}

func NormaizeShopifyDomain(shopifyDomain string) string {
	return strings.Replace(shopifyDomain, ".myshopify.com", "", -1) + ".myshopify.com"
}
