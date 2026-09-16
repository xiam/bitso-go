package bitso

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderTimeInForce_String(t *testing.T) {
	tests := []struct {
		name     string
		value    OrderTimeInForce
		expected string
	}{
		{"good till cancelled", OrderTimeInForceGoodTillCancelled, "goodtillcancelled"},
		{"fill or kill", OrderTimeInForceFillOrKill, "fillorkill"},
		{"immediate or cancel", OrderTimeInForceImmediateOrCancel, "immediateorcancel"},
		{"post only", OrderTimeInForcePostOnly, "postonly"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.value.String(), "time_in_force constant should match Bitso wire literal")
		})
	}
}

func TestMarginOrderType_String(t *testing.T) {
	assert.Equal(t, "CROSS_MARGIN", MarginOrderTypeCrossMargin.String(), "margin_order_type constant should match Bitso wire literal")
}
