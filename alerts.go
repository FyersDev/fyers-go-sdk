package fyersgosdk

import (
	"encoding/json"
	"net/http"
)

func (m *FyersModel) GetAlerts() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, AlertsURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) ToggleAlert(alertId string) (string, error) {
	headers := m.authHeader()
	headers.Set("Content-Type", "application/json")
	body, _ := json.Marshal(map[string]string{"alertId": alertId})
	resp, err := m.httpClient.DoRaw(http.MethodPut, ToggleAlertURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) CreateAlert(alertRequest AlertRequest) (string, error) {
	headers := m.authHeader()
	if alertRequest.Agent == "" {
		alertRequest.Agent = "fyers-api"
	}
	body, err := json.Marshal(alertRequest)
	if err != nil {
		return "", err
	}
	resp, err := m.httpClient.DoRaw(http.MethodPost, AlertsURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) DeleteAlert(alertId string) (string, error) {
	headers := m.authHeader()
	body, _ := json.Marshal(map[string]string{"alertId": alertId, "agent": "fyers-api"})
	resp, err := m.httpClient.DoRaw(http.MethodDelete, AlertsURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

func (m *FyersModel) UpdateAlert(alertId string, alertRequest AlertRequest) (string, error) {
	headers := m.authHeader()
	if alertRequest.Agent == "" {
		alertRequest.Agent = "fyers-api"
	}
	type updateRequest struct {
		AlertId string `json:"alertId"`
		AlertRequest
	}
	req := updateRequest{AlertId: alertId, AlertRequest: alertRequest}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	resp, err := m.httpClient.DoRaw(http.MethodPut, AlertsURL, body, headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
