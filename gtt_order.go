package fyersgosdk

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *FyersModel) GTTSingleOrderAction(orderRequest GTTOrderRequest) (string, error) {
	body, err := json.Marshal(orderRequest)
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, GTTOrderURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GTTMultiOrderAction(orderRequests []GTTOrderRequest) (string, error) {
	if len(orderRequests) == 0 {
		return "", fmt.Errorf("at least one order request required")
	}
	body, err := json.Marshal(orderRequests[0])
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, GTTOrderURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ModifyGTTOrder(orderRequests []ModifyGTTOrderRequest) (string, error) {
	if len(orderRequests) == 0 {
		return "", fmt.Errorf("at least one order request required")
	}

	body, err := json.Marshal(orderRequests[0])
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	headers := m.authHeader()
	resp, err := m.httpClient.DoRaw(http.MethodPatch, GTTOrderURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) CancelGTTOrder(orderId string) (string, error) {
	body, err := json.Marshal(CancelGTTOrderRequest{Id: orderId})
	if err != nil {
		return "", fmt.Errorf("marshal cancel request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodDelete, GTTOrderURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) GetGTTOrderBook() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, GTTOrderBookURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
