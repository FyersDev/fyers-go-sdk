package fyersgosdk

import (
	"encoding/json"
	"fmt"
)

// FloatValue represents an optional float that can be omitted, set to a number,
// or explicitly serialized as JSON null (used to remove TP/SL legs on modify/attach).
type FloatValue struct {
	Value float64
	Null  bool
}

// FloatOffset returns a FloatValue set to the given offset.
func FloatOffset(v float64) *FloatValue {
	return &FloatValue{Value: v}
}

// NullOffset returns a FloatValue that serializes as JSON null (removes a TP/SL leg).
func NullOffset() *FloatValue {
	return &FloatValue{Null: true}
}

// MarshalJSON serializes the value as a JSON number or null.
func (f FloatValue) MarshalJSON() ([]byte, error) {
	if f.Null {
		return []byte("null"), nil
	}
	return json.Marshal(f.Value)
}

// UnmarshalJSON accepts a JSON number or null.
func (f *FloatValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.Null = true
		f.Value = 0
		return nil
	}
	f.Null = false
	return json.Unmarshal(data, &f.Value)
}

func validateProductType(productType string) error {
	if productType == ProductBO || productType == ProductCO {
		return fmt.Errorf("BO/CO product types are deprecated (code %d). Use standard product types with takeProfit/stopLoss offsets instead", deprecatedBOCOCode)
	}
	return nil
}

func validateOffset(name string, value float64, legType int) error {
	if value < 0 {
		return fmt.Errorf("takeProfit and stopLoss offsets must be positive numbers")
	}
	if value == 0 {
		return nil
	}
	if value < minTPSLOffset {
		return fmt.Errorf("%s offset must be greater than or equal to %g", name, minTPSLOffset)
	}
	if legType == LegTypePercent && value > maxTPSLPercent {
		return fmt.Errorf("takeProfit/stopLoss percentage must be between %g and %g", minTPSLOffset, maxTPSLPercent)
	}
	return nil
}

func validateOptionalFloatOffset(name string, v *FloatValue, legType int, hasLegType bool) error {
	if v == nil || v.Null {
		return nil
	}
	effectiveLegType := LegTypePoints
	if hasLegType {
		effectiveLegType = legType
	}
	if v.Value < 0 {
		return fmt.Errorf("takeProfit and stopLoss offsets must be positive numbers")
	}
	if v.Value == 0 {
		return nil
	}
	return validateOffset(name, v.Value, effectiveLegType)
}

func (r OrderRequest) validateTPSL() error {
	if err := validateProductType(r.ProductType); err != nil {
		return err
	}
	hasLegType := r.LegType != 0
	if hasLegType && r.LegType != LegTypePoints && r.LegType != LegTypePercent {
		return fmt.Errorf("legType must be 1 (points) or 2 (percent)")
	}
	if err := validateOptionalFloatOffset("takeProfit", r.TakeProfit, r.LegType, hasLegType); err != nil {
		return err
	}
	return validateOptionalFloatOffset("stopLoss", r.StopLoss, r.LegType, hasLegType)
}

func (r ModifyOrderRequest) validateTPSL() error {
	hasLegType := r.LegType != 0
	if hasLegType && r.LegType != LegTypePoints && r.LegType != LegTypePercent {
		return fmt.Errorf("legType must be 1 (points) or 2 (percent)")
	}
	if err := validateOptionalFloatOffset("takeProfit", r.TakeProfit, r.LegType, hasLegType); err != nil {
		return err
	}
	return validateOptionalFloatOffset("stopLoss", r.StopLoss, r.LegType, hasLegType)
}

func (r AttachPositionLegsRequest) validateTPSL() error {
	if r.PositionID == "" {
		return fmt.Errorf("positionId is required")
	}
	hasLegType := r.LegType != 0
	if hasLegType && r.LegType != LegTypePoints && r.LegType != LegTypePercent {
		return fmt.Errorf("legType must be 1 (points) or 2 (percent)")
	}
	if err := validateOptionalFloatOffset("takeProfit", r.TakeProfit, r.LegType, hasLegType); err != nil {
		return err
	}
	return validateOptionalFloatOffset("stopLoss", r.StopLoss, r.LegType, hasLegType)
}
