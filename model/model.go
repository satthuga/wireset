package model

type AuthResponse struct {
	Message             string `json:"message,omitempty"`
	AuthenticationUrl   string `json:"authenticationUrl,omitempty"`
	FirebaseCustomToken string `json:"firebaseCustomToken,omitempty"`
}

type ShopifyToken struct {
	ShopID      string
	AccessToken string
}
