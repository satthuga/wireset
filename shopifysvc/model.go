package shopifysvc

type Shop struct {
	ID                   string `json:"id" firestore:"id"`
	Name                 string `json:"name" firestore:"name"`
	Email                string `json:"email" firestore:"email"`
	CountryCode          string `json:"countryCode" firestore:"countryCode"`
	Domain               string `json:"domain" firestore:"domain"`
	MyshopifyDomain      string `json:"myshopifyDomain" firestore:"myshopifyDomain"`
	TimezoneAbbreviation string `json:"timezoneAbbreviation" firestore:"timezoneAbbreviation"`
	IanaTimezone         string `json:"ianaTimezone" firestore:"ianaTimezone"`
	CurrencyCode         string `json:"currencyCode" firestore:"currencyCode"`
}

type Product struct {
	ID string `json:"id" firestore:"id"`
}
