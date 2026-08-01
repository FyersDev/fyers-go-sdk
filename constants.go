package fyersgosdk

// Product type constants for order placement.
const (
	ProductCNC      = "CNC"
	ProductIntraday = "INTRADAY"
	ProductMargin   = "MARGIN"
	ProductMTF      = "MTF"
)

// LegType values for TP/SL offset measurement.
const (
	// LegTypePoints measures takeProfit/stopLoss as price-point offsets (default).
	LegTypePoints = 1
	// LegTypePercent measures takeProfit/stopLoss as a percentage of entry price.
	LegTypePercent = 2
)
