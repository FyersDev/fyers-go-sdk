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
	}
	resp, err := m.httpClient.Do(http.MethodGet, OptionChainURl, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
