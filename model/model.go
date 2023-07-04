package model

type AuthResponse struct {
	Message             string `json:"message,omitempty"`
	AuthenticationUrl   string `json:"authenticationUrl,omitempty"`
	FirebaseCustomToken string `json:"firebaseCustomToken,omitempty"`
}

type ShopifyToken struct {
	ShopID      string `json:"shopId" firestore:"shopId"`
	AccessToken string `json:"accessToken" firestore:"accessToken"`
}

type Plan struct {
	PlanID   string    `json:"planId" firestore:"planId"`
	Features []Feature `json:"features" firestore:"features"`
}

type Feature struct {
	FeatureID string `json:"featureId" firestore:"featureId"`
}
