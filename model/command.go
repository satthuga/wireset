package model

type InstallWebhookCmd struct {
	MyshopifyDomain string
	AccessToken     string
}

type CreateInsuranceProductCmd struct {
	MyshopifyDomain string
	AccessToken     string
}

type ExampleCmd struct{}

type SetShopStateCmd struct {
	ShopID string
	State  map[string]interface{}
}
