package fyersgosdk

import (
	"net/http"
)

func (m *FyersModel) GetProfile() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, ProfileURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetFunds() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, FundURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetHoldings() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, HoldingsURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) Logout() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodPost, LogoutURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
