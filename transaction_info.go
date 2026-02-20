package fyersgosdk

import (
	"net/http"
	"net/url"
)

func (m *FyersModel) GetOrderBook() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, OrderBookURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetOrderBookByTag(tag string) (string, error) {
	params := url.Values{}
	params.Set("order_tag", tag)
	resp, err := m.httpClient.Do(http.MethodGet, OrdersByTagURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetOrderById(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)
	resp, err := m.httpClient.Do(http.MethodGet, OrderByIdURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetPositions() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, PositionURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetTradeBook() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, TradeBookURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetTradeBookByTag(tag string) (string, error) {
	params := url.Values{}
	params.Set("order_tag", tag)
	resp, err := m.httpClient.Do(http.MethodGet, TradeBookByTagURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
