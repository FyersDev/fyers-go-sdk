package fyersgosdk

// Product type constants for order placement.
const (
	ProductCNC      = "CNC"
	ProductIntraday = "INTRADAY"
	ProductMargin   = "MARGIN"
	ProductMTF      = "MTF"

	// Deprecated: ProductBO is deprecated and rejected by the API (HTTP 400, code -1800).
	// Use ProductIntraday or ProductMargin with TakeProfit/StopLoss offsets instead.
	ProductBO = "BO"

	// Deprecated: ProductCO is deprecated and rejected by the API (HTTP 400, code -1800).
	// Use a standard product type with TakeProfit/StopLoss offsets instead.
	ProductCO = "CO"
)

// LegType values for TP/SL offset measurement.
const (
	// LegTypePoints measures takeProfit/stopLoss as price-point offsets (default).
	LegTypePoints = 1
	// LegTypePercent measures takeProfit/stopLoss as a percentage of entry price.
	LegTypePercent = 2
)

const (
	minTPSLOffset      = 0.0025
	maxTPSLPercent     = 100.0
	deprecatedBOCOCode = -1800
)
