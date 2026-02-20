package fyersgosdk

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

func SetClientData(appId, appSecret, redirectUrl string) *Client {
	client := &Client{
		appId:       appId,
		appSecret:   appSecret,
		redirectUrl: redirectUrl,
		httpClient:  NewHTTPClient(nil, nil, false),
	}

	return client
}

func (c *Client) SetAccessToken(accessToken string) *Client {
	c.accessToken = accessToken
	return c
}

func (c *Client) GetLoginURL() string {
	return fmt.Sprintf("%s&client_id=%s&redirect_uri=%s&response_type=%s&state=%s", GenerateAuthCodeURL, c.appId, c.redirectUrl, "code", "sample_state")
}

func NewFyersModel(appId, accessToken string) *FyersModel {
	return &FyersModel{
		appId:       appId,
		accessToken: accessToken,
		httpClient:  NewHTTPClient(nil, nil, false),
	}
}

func (m *FyersModel) authHeader() http.Header {
	h := http.Header{}
	h.Set("Authorization", fmt.Sprintf("%s:%s", m.appId, m.accessToken))
	return h
}

func (c *Client) GenerateAccessToken(authToken string, fyClient *Client) (string, error) {

	h := sha256.New()
	h.Write([]byte(fyClient.appId + ":" + fyClient.appSecret))

	requestBody := fmt.Sprintf(`{"code":"%s","appIdHash":"%s","grant_type":"authorization_code"}`, authToken, fmt.Sprintf("%x", h.Sum(nil)))

	response, err := c.httpClient.DoRaw(http.MethodPost, ValidateAuthCodeURL, []byte(requestBody), nil)
	if err != nil {
		return "", err
	}

	return string(response.Body), nil
}

func (c *Client) GenerateAccessTokenFromRefreshToken(refreshToken, pin string, fyClient *Client) (string, error) {

	h := sha256.New()
	h.Write([]byte(fyClient.appId + ":" + fyClient.appSecret))

	requestBody := fmt.Sprintf(`{"refresh_token":"%s","appIdHash":"%s","grant_type":"refresh_token","pin":"%s"}`, refreshToken, fmt.Sprintf("%x", h.Sum(nil)), pin)

	headers := make(http.Header)
	response, err := c.httpClient.DoRaw(http.MethodPost, ValidateRefreshTokenURL, []byte(requestBody), headers)
	if err != nil {
		return "", err
	}

	return string(response.Body), nil
}
