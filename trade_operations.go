package fyersgosdk

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (m *FyersModel) ModifyOrder(orderRequest ModifyOrderRequest) (string, error) {
	body, err := json.Marshal(orderRequest)
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPatch, SingleOrderActionURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ModifyMutliOrder(requests []ModifyMultiOrderItem) (string, error) {
	if len(requests) == 0 {
		return "", fmt.Errorf("at least one order modification required")
	}
	body, err := json.Marshal(requests)
	if err != nil {
		return "", fmt.Errorf("marshal multi order request: %w", err)
	}
	headers := m.authHeader()
	resp, err := m.httpClient.DoRaw(http.MethodPatch, MultipleOrderActionURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) CancelOrder(Id string) (string, error) {
	body, err := json.Marshal(CancelOrderRequest{Id: Id})
	if err != nil {
		return "", fmt.Errorf("marshal cancel order request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodDelete, SingleOrderActionURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) CancelMutliOrder(orderIds []string) (string, error) {
	if len(orderIds) == 0 {
		return "", fmt.Errorf("at least one order cancellation required")
	}
	bodyItems := make([]CancelOrderRequest, len(orderIds))
	for i, id := range orderIds {
		bodyItems[i] = CancelOrderRequest{Id: id}
	}
	body, err := json.Marshal(bodyItems)
	if err != nil {
		return "", fmt.Errorf("marshal multi order cancellation request: %w", err)
	}
	headers := m.authHeader()
	resp, err := m.httpClient.DoRaw(http.MethodDelete, MultipleOrderActionURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ExitPosition() (string, error) {
	body, _ := json.Marshal(map[string]int{"exit_all": 1})
	resp, err := m.httpClient.DoRaw(http.MethodDelete, PositionURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ExitPositionById(orderId []string) (string, error) {
	if len(orderId) == 0 {
		return "", fmt.Errorf("at least one order id required")
	}
	bodyItems := make([]map[string]string, len(orderId))
	for i, id := range orderId {
		bodyItems[i] = map[string]string{"orderId": id}
	}
	body, err := json.Marshal(bodyItems)
	if err != nil {
		return "", fmt.Errorf("marshal exit position by id request: %w", err)
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, PositionURL, body, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ExitPositionByProductType(req ExitPositionByProductTypeRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal exit position by product type request: %w", err)
	}
	headers := m.authHeader()
	resp, err := m.httpClient.DoRaw(http.MethodDelete, PositionURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) CancelPendingOrders(req CancelPendingOrdersRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal cancel pending orders request: %w", err)
	}
	headers := m.authHeader()
	resp, err := m.httpClient.DoRaw(http.MethodDelete, PositionURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ConvertPosition(req ConvertPositionRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal convert position request: %w", err)
	}
	headers := m.authHeader()
	headers.Set("Content-Type", "application/json")
	resp, err := m.httpClient.DoRaw(http.MethodPost, PositionURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
