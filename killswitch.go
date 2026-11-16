package fyersgosdk

import (
	"net/http"
)

// ActivateKillswitch places the account on break (POST /killswitch/activate).
// Activating the kill switch blocks trading on the account.
//
// Like every other method in this SDK, an API-level rejection is not a Go
// error: the response body is returned unchanged as the string, so check "s".
// For example, when the kill switch is already active the API answers HTTP 400
// with {"code": 1500, "message": "Already on break, Please enjoy the break", "s": "error"}.
// A non-nil error is returned only for transport failures.
func (m *FyersModel) ActivateKillswitch() (string, error) {
	headers := m.authHeader()
	headers.Set("Content-Type", "application/json")
	resp, err := m.httpClient.DoRaw(http.MethodPost, KillswitchActivateURL, []byte("{}"), headers)
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}

// GetKillswitchStatus returns whether the kill switch is active, along with
// open positions and pending orders (GET /killswitch/status).
// The response body is returned unchanged as a JSON string.
func (m *FyersModel) GetKillswitchStatus() (string, error) {
	resp, err := m.httpClient.DoRaw(http.MethodGet, KillswitchStatusURL, nil, m.authHeader())
	if err != nil {
		return "", err
	}
	return string(resp.Body), nil
}
