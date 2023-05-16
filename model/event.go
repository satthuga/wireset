package model

type ShopInstalledEvt struct {
	MyshopifyDomain string
	AccessToken     string
}

type ShopUninstalledEvt struct {
	MyshopifyDomain string
}

type ShopCheckedInEvt struct {
	MyshopifyDomain string
	AccessToken     string
}

type ServerStartedEvt struct {
}
