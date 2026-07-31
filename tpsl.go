package fyersgosdk

import (
	"encoding/json"
	"fmt"
)

func validateLegType(legType int) error {
	if legType != 0 && legType != LegTypePoints && legType != LegTypePercent {
		return fmt.Errorf("legType must be 1 (points) or 2 (percent)")
	}
	return nil
}

func (r OrderRequest) validateTPSL() error {
	return validateLegType(r.LegType)
}

func (r ModifyOrderRequest) validateTPSL() error {
	return validateLegType(r.LegType)
}

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

func (r AttachPositionLegsRequest) validateTPSL() error {
	if r.PositionID == "" {
		return fmt.Errorf("positionId is required")
	}
	return validateLegType(r.LegType)
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
