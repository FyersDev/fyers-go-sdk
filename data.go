package fyersgosdk

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (m *FyersModel) GetHistory(historyRequest HistoryRequest) (string, error) {
	params := url.Values{}
	params.Set("symbol", historyRequest.Symbol)
	params.Set("resolution", historyRequest.Resolution)
	params.Set("date_format", historyRequest.DateFormat)
	params.Set("range_from", historyRequest.RangeFrom)
	params.Set("range_to", historyRequest.RangeTo)
	params.Set("cont_flag", historyRequest.ContFlag)
	resp, err := m.httpClient.Do(http.MethodGet, StockHistoryURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetStockQuotes(symbols []string) (string, error) {
	if len(symbols) == 0 {
		return "", fmt.Errorf("at least one symbol required")
	}

	params := url.Values{"symbols": {strings.Join(symbols, ",")}}
	resp, err := m.httpClient.Do(http.MethodGet, StockQuotesURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetMarketDepth(req MarketDepthRequest) (string, error) {
	params := url.Values{
		"symbol":     {req.Symbol},
		"ohlcv_flag": {req.OHLCV},
	}
	resp, err := m.httpClient.Do(http.MethodGet, MarketDepthURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetOptionChain(req OptionChainRequest) (string, error) {
	params := url.Values{
		"symbol":      {req.Symbol},
		"strikecount": {strconv.Itoa(req.StrikeCount)},
		"timestamp":   {req.Timestamp},
		"greeks":      {req.Greeks},
	}
	resp, err := m.httpClient.Do(http.MethodGet, OptionChainURl, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetExpiryDates(req ExpiryDatesRequest) (string, error) {
	params := url.Values{}
	params.Set("symbol", req.UnderlyingSymbol)
	params.Set("range_from", req.RangeFrom)
	params.Set("range_to", req.RangeTo)
	params.Set("date_format", req.DateFormat)

	resp, err := m.httpClient.Do(http.MethodGet, ExpiryDatesURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetHistoryUnderlyingSymbols(req HistoryUnderlyingSymbolsRequest) (string, error) {
	params := url.Values{}
	params.Set("symbol", req.UnderlyingSymbol)
	params.Set("expiry_date", req.ExpiryDate)

	resp, err := m.httpClient.Do(http.MethodGet, HistoryUnderlyingSymbolsURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetFuturesChain(req FuturesChainRequest) (string, error) {
	if req.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	params := url.Values{}
	params.Set("symbol", req.Symbol)

	resp, err := m.httpClient.Do(http.MethodGet, FuturesChainURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetFNOHistoricalData(req FNOHistoricalDataRequest) (string, error) {
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("resolution", req.Resolution)
	params.Set("date_format", req.DateFormat)
	params.Set("range_from", req.RangeFrom)
	params.Set("range_to", req.RangeTo)
	if req.OiFlag != "" {
		params.Set("oi_flag", req.OiFlag)
	}
	if req.Greeks != "" {
		params.Set("greeks", req.Greeks)
	}

	resp, err := m.httpClient.Do(http.MethodGet, FNOHistoricalDataURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
