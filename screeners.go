package fyersgosdk

import (
	"net/http"
	"net/url"
)

func (m *FyersModel) ScreenersConfig() (string, error) {
	resp, err := m.httpClient.Do(
		http.MethodGet,
		ScreenersConfigURL,
		nil,
		m.authHeader(),
	)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ScreenersQuery(req *ScreenersQuery) (string, error) {
	params := url.Values{}
	if req != nil {
		params.Set("screener", req.Screener)
		params.Set("universe", req.Universe)
		params.Set("fields", req.Fields)
		params.Set("order_by", req.OrderBy)
		params.Set("order", req.Order)
	}
	resp, err := m.httpClient.Do(
		http.MethodGet,
		ScreenersQueryURL,
		params,
		m.authHeader(),
	)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ScreenersCandlestick(req *ScreenersCandlestick) (string, error) {
	params := url.Values{}
	if req != nil {
		params.Set("screener", req.Screener)
	}
	resp, err := m.httpClient.Do(
		http.MethodGet,
		ScreenersCandlestickURL,
		params,
		m.authHeader(),
	)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ScreenersTechnical(req *ScreenersTechnical) (string, error) {
	params := url.Values{}
	if req != nil {
		params.Set("screener", req.Screener)
	}
	resp, err := m.httpClient.Do(
		http.MethodGet,
		ScreenersTechnicalURL,
		params,
		m.authHeader(),
	)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
