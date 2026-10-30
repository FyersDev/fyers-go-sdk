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

func (m *FyersModel) ScreenersStandard(req *ScreenersStandard) (string, error) {
	params := url.Values{}
	if req != nil {
		setParam(params, "screener_id", req.ScreenerId)
		setParam(params, "universe", req.Universe)
		setParam(params, "fields", req.Fields)
		setParam(params, "order_by", req.OrderBy)
		setParam(params, "order", req.Order)
		setParam(params, "timeframe", req.Timeframe)
		setParam(params, "expiry", req.Expiry)
		setParam(params, "option_type", req.OptionType)
	}
	resp, err := m.httpClient.Do(
		http.MethodGet,
		ScreenersStandardURL,
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
		setParam(params, "screener_id", req.ScreenerId)
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
		setParam(params, "screener_id", req.ScreenerId)
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

// setParam adds a query param only when it has a value. The API rejects a param
// it does not expect, so an empty one must not be sent.
func setParam(params url.Values, name, value string) {
	if value == "" {
		return
	}
	params.Set(name, value)
}
