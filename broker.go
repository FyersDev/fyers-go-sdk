package fyersgosdk

import (
	"net/http"
)

func (m *FyersModel) GetMarketStatus() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, MarketStatusURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
