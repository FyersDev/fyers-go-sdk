package fyersgosdk

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *FyersModel) SingleOrderAction(orderRequest OrderRequest) (string, error) {
	body, err := json.Marshal(orderRequest)
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, SingleOrderActionURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) MultiOrderAction(orderRequests []OrderRequest) (string, error) {
	body, err := json.Marshal(orderRequests)
	if err != nil {
		return "", fmt.Errorf("marshal order requests: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, MultipleOrderActionURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) MultiLegOrderAction(orderRequests []MultiLegOrderRequest) (string, error) {
	body, err := json.Marshal(orderRequests)
	if err != nil {
		return "", fmt.Errorf("marshal order requests: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, MultiLegOrderURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
