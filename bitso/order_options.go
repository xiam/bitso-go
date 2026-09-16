package bitso

// OrderTimeInForce controls how long an order remains active.
type OrderTimeInForce string

const (
	OrderTimeInForceGoodTillCancelled OrderTimeInForce = "goodtillcancelled"
	OrderTimeInForceFillOrKill        OrderTimeInForce = "fillorkill"
	OrderTimeInForceImmediateOrCancel OrderTimeInForce = "immediateorcancel"
	OrderTimeInForcePostOnly          OrderTimeInForce = "postonly"
)

func (t OrderTimeInForce) String() string {
	return string(t)
}

// MarginOrderType identifies the margin mode for an order.
type MarginOrderType string

const (
	MarginOrderTypeCrossMargin MarginOrderType = "CROSS_MARGIN"
)

func (m MarginOrderType) String() string {
	return string(m)
}
