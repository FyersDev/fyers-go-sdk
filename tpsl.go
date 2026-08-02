package fyersgosdk

import "encoding/json"

// MarshalJSON emits takeProfit/stopLoss as null when Clear* flags are set.
func (r ModifyOrderRequest) MarshalJSON() ([]byte, error) {
	type alias ModifyOrderRequest
	raw, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	if !r.ClearTakeProfit && !r.ClearStopLoss {
		return raw, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if r.ClearTakeProfit {
		m["takeProfit"] = nil
	}
	if r.ClearStopLoss {
		m["stopLoss"] = nil
	}
	return json.Marshal(m)
}

// MarshalJSON emits takeProfit/stopLoss as null when Clear* flags are set.
func (r AttachPositionLegsRequest) MarshalJSON() ([]byte, error) {
	type alias AttachPositionLegsRequest
	raw, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	if !r.ClearTakeProfit && !r.ClearStopLoss {
		return raw, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if r.ClearTakeProfit {
		m["takeProfit"] = nil
	}
	if r.ClearStopLoss {
		m["stopLoss"] = nil
	}
	return json.Marshal(m)
}
