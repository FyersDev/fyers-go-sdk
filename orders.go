package fyersgosdk

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *FyersModel) SingleOrderAction(orderRequest OrderRequest) (string, error) {
	if err := orderRequest.validateTPSL(); err != nil {
		return "", err
	}
	body, err := json.Marshal(orderRequest)
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	headers := m.authHeader()
	headers.Set("Content-Type", "application/json")
	resp, err := m.httpClient.DoRaw(http.MethodPost, SingleOrderActionURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) MultiOrderAction(orderRequests []OrderRequest) (string, error) {
	for i, req := range orderRequests {
		if err := req.validateTPSL(); err != nil {
			return "", fmt.Errorf("order[%d]: %w", i, err)
		}
	}
	body, err := json.Marshal(orderRequests)
	if err != nil {
		return "", fmt.Errorf("marshal order requests: %w", err)
	}
	headers := m.authHeader()
	headers.Set("Content-Type", "application/json")
	resp, err := m.httpClient.DoRaw(http.MethodPost, MultipleOrderActionURL, body, headers)
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
	headers := m.authHeader()
	headers.Set("Content-Type", "application/json")
	resp, err := m.httpClient.DoRaw(http.MethodPost, MultiLegOrderURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
