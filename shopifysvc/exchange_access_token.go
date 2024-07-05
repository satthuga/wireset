package shopifysvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// exchange the session token with access token
type AccessTokenRequest struct {
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	GrantType          string `json:"grant_type"`
	SubjectToken       string `json:"subject_token"`
	SubjectTokenType   string `json:"subject_token_type"`
	RequestedTokenType string `json:"requested_token_type"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func ExchangeAccessToken(shop, clientID, clientSecret, sessionToken string) (*AccessTokenResponse, error) {
	url := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)

	data := AccessTokenRequest{
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       sessionToken,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:id_token",
		RequestedTokenType: "urn:shopify:params:oauth:token-type:offline-access-token",
	}

	payload, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var accessTokenResponse AccessTokenResponse
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		return nil, err
	}

	return &accessTokenResponse, nil
}
