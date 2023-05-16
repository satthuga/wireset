package shopifysvc

type Shop struct {
	ID                   string
	Name                 string
	Email                string
	CountryCode          string
	Domain               string
	MyshopifyDomain      string
	TimezoneAbbreviation string
	IanaTimezone         string
	CurrencyCode         string
}

type Product struct {
	ID string
}
