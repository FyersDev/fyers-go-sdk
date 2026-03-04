package fyersgosdk

import (
	"net/http"
	"net/url"
	"strconv"
)

func (m *FyersModel) OrderHistory(req *OrderBookHistoryFilter) (string, error) {
	params := url.Values{}
	if req != nil {
		for _, e := range req.Exchange {
			params.Add("exchange_type", e)
		}
		for _, s := range req.Segment {
			params.Add("segment_type", s)
		}
		for _, f := range req.Status {
			params.Add("status", f)
		}
		if req.Symbol != "" {
			params.Set("symbol", req.Symbol)
		}
		if req.FromDate != "" {
			params.Set("from_date", req.FromDate)
		}
		if req.ToDate != "" {
			params.Set("to_date", req.ToDate)
		}
		if req.PageNo != 0 {
			params.Set("page_no", strconv.Itoa(req.PageNo))
		}
		if req.PageSize != 0 {
			params.Set("page_size", strconv.Itoa(req.PageSize))
		}
	}
	resp, err := m.httpClient.Do(http.MethodGet, OrderBookHistoryURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) TradeHistory(req *TradeBookHistoryFilter) (string, error) {
	params := url.Values{}
	if req != nil {
		for _, e := range req.Exchange {
			params.Add("exchange_type", e)
		}
		for _, s := range req.Segment {
			params.Add("segment_type", s)
		}
		for _, f := range req.Status {
			params.Add("status", f)
		}
		if req.Symbol != "" {
			params.Set("symbol", req.Symbol)
		}
		if req.FromDate != "" {
			params.Set("from_date", req.FromDate)
		}
		if req.ToDate != "" {
			params.Set("to_date", req.ToDate)
		}
		if req.PageNo != 0 {
			params.Set("page_no", strconv.Itoa(req.PageNo))
		}
		if req.PageSize != 0 {
			params.Set("page_size", strconv.Itoa(req.PageSize))
		}
	}
	resp, err := m.httpClient.Do(http.MethodGet, TradeBookHistoryURL, params, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
