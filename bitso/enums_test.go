package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// OrderSide tests

func TestOrderSide_String(t *testing.T) {
	tests := []struct {
		side     OrderSide
		expected string
	}{
		{OrderSideBuy, "buy"},
		{OrderSideSell, "sell"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.side.String())
		})
	}
}

func TestOrderSide_StringUnknown(t *testing.T) {
	// Unknown values return a descriptive string instead of panicking
	t.Run("invalid value", func(t *testing.T) {
		invalid := OrderSide(99)
		assert.Equal(t, "OrderSide(99)", invalid.String())
	})

	t.Run("none value", func(t *testing.T) {
		none := OrderSideNone
		assert.Equal(t, "OrderSide(0)", none.String())
	})
}

func TestOrderSide_MarshalJSON(t *testing.T) {
	tests := []struct {
		side     OrderSide
		expected string
	}{
		{OrderSideBuy, `"buy"`},
		{OrderSideSell, `"sell"`},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			data, err := json.Marshal(tc.side)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestOrderSide_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected OrderSide
	}{
		{`"buy"`, OrderSideBuy},
		{`"sell"`, OrderSideSell},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var side OrderSide
			err := json.Unmarshal([]byte(tc.json), &side)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, side)
		})
	}
}

func TestOrderSide_UnmarshalJSON_Invalid(t *testing.T) {
	var side OrderSide
	err := json.Unmarshal([]byte(`"invalid"`), &side)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported order side")
}

func TestOrderSide_SQLValue(t *testing.T) {
	side := OrderSideBuy
	value, err := side.Value()

	require.NoError(t, err)
	assert.Equal(t, "buy", value)
}

func TestOrderSide_SQLScan(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		var side OrderSide
		err := side.Scan("sell")
		require.NoError(t, err)
		assert.Equal(t, OrderSideSell, side)
	})

	t.Run("nil value", func(t *testing.T) {
		var side OrderSide
		err := side.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("invalid value", func(t *testing.T) {
		var side OrderSide
		err := side.Scan("invalid")
		require.Error(t, err)
	})
}

// OrderType tests

func TestOrderType_String(t *testing.T) {
	tests := []struct {
		orderType OrderType
		expected  string
	}{
		{OrderTypeMarket, "market"},
		{OrderTypeLimit, "limit"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.orderType.String())
		})
	}
}

func TestOrderType_StringUnknown(t *testing.T) {
	// Unknown values return a descriptive string instead of panicking
	t.Run("invalid value", func(t *testing.T) {
		invalid := OrderType(99)
		assert.Equal(t, "OrderType(99)", invalid.String())
	})

	t.Run("none value", func(t *testing.T) {
		none := OrderTypeNone
		assert.Equal(t, "OrderType(0)", none.String())
	})
}

func TestOrderType_MarshalJSON(t *testing.T) {
	tests := []struct {
		orderType OrderType
		expected  string
	}{
		{OrderTypeMarket, `"market"`},
		{OrderTypeLimit, `"limit"`},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			data, err := json.Marshal(tc.orderType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestOrderType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected OrderType
	}{
		{`"market"`, OrderTypeMarket},
		{`"limit"`, OrderTypeLimit},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var orderType OrderType
			err := json.Unmarshal([]byte(tc.json), &orderType)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, orderType)
		})
	}
}

func TestOrderType_UnmarshalJSON_Invalid(t *testing.T) {
	var orderType OrderType
	err := json.Unmarshal([]byte(`"invalid"`), &orderType)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported order type")
}

// OrderStatus tests

func TestOrderStatus_String(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected string
	}{
		{OrderStatusOpen, "open"},
		{OrderStatusQueued, "queued"},
		{OrderStatusPartialFill, "partially filled"},
		{OrderStatusCancelled, "cancelled"},
		{OrderStatusCompleted, "completed"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestOrderStatus_StringUnknown(t *testing.T) {
	// Unknown values return a descriptive string instead of panicking
	t.Run("invalid value", func(t *testing.T) {
		invalid := OrderStatus(99)
		assert.Equal(t, "OrderStatus(99)", invalid.String())
	})

	t.Run("none value", func(t *testing.T) {
		none := OrderStatusNone
		assert.Equal(t, "OrderStatus(0)", none.String())
	})
}

func TestOrderStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected string
	}{
		{OrderStatusOpen, `"open"`},
		{OrderStatusQueued, `"queued"`},
		{OrderStatusPartialFill, `"partially filled"`},
		{OrderStatusCancelled, `"cancelled"`},
		{OrderStatusCompleted, `"completed"`},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			data, err := json.Marshal(tc.status)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestOrderStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected OrderStatus
	}{
		{`"open"`, OrderStatusOpen},
		{`"queued"`, OrderStatusQueued},
		{`"partially filled"`, OrderStatusPartialFill},
		{`"cancelled"`, OrderStatusCancelled},
		{`"completed"`, OrderStatusCompleted},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var status OrderStatus
			err := json.Unmarshal([]byte(tc.json), &status)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, status)
		})
	}
}

func TestOrderStatus_UnmarshalJSON_Invalid(t *testing.T) {
	var status OrderStatus
	err := json.Unmarshal([]byte(`"invalid"`), &status)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported status")
}

func TestOrderStatus_SQLValue(t *testing.T) {
	status := OrderStatusCompleted
	value, err := status.Value()

	require.NoError(t, err)
	assert.Equal(t, "completed", value)
}

func TestOrderStatus_SQLScan(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		var status OrderStatus
		err := status.Scan("partially filled")
		require.NoError(t, err)
		assert.Equal(t, OrderStatusPartialFill, status)
	})

	t.Run("nil value", func(t *testing.T) {
		var status OrderStatus
		err := status.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("invalid value", func(t *testing.T) {
		var status OrderStatus
		err := status.Scan("invalid")
		require.Error(t, err)
	})
}

// Operation tests

func TestOperation_String(t *testing.T) {
	tests := []struct {
		op       Operation
		expected string
	}{
		{OperationFunding, "funding"},
		{OperationWithdrawal, "withdrawal"},
		{OperationTrade, "trade"},
		{OperationFee, "fee"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.op.String())
		})
	}
}

func TestOperation_StringUnknown(t *testing.T) {
	// Unknown values return a descriptive string instead of panicking
	t.Run("invalid value", func(t *testing.T) {
		invalid := Operation(99)
		assert.Equal(t, "Operation(99)", invalid.String())
	})

	t.Run("none value", func(t *testing.T) {
		none := OperationNone
		assert.Equal(t, "Operation(0)", none.String())
	})
}

func TestOperation_MarshalJSON(t *testing.T) {
	tests := []struct {
		op       Operation
		expected string
	}{
		{OperationFunding, `"funding"`},
		{OperationWithdrawal, `"withdrawal"`},
		{OperationTrade, `"trade"`},
		{OperationFee, `"fee"`},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			data, err := json.Marshal(tc.op)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestOperation_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected Operation
	}{
		{`"funding"`, OperationFunding},
		{`"withdrawal"`, OperationWithdrawal},
		{`"trade"`, OperationTrade},
		{`"fee"`, OperationFee},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var op Operation
			err := json.Unmarshal([]byte(tc.json), &op)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, op)
		})
	}
}

func TestOperation_UnmarshalJSON_Invalid(t *testing.T) {
	var op Operation
	err := json.Unmarshal([]byte(`"invalid"`), &op)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported operation")
}

func TestOperation_SQLValue(t *testing.T) {
	op := OperationTrade
	value, err := op.Value()

	require.NoError(t, err)
	assert.Equal(t, "trade", value)
}

func TestOperation_SQLScan(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		var op Operation
		err := op.Scan("withdrawal")
		require.NoError(t, err)
		assert.Equal(t, OperationWithdrawal, op)
	})

	t.Run("nil value", func(t *testing.T) {
		var op Operation
		err := op.Scan(nil)
		require.NoError(t, err)
	})

	t.Run("invalid value", func(t *testing.T) {
		var op Operation
		err := op.Scan("invalid")
		require.Error(t, err)
	})
}

// Integration tests with structs

func TestEnums_InStruct(t *testing.T) {
	t.Run("marshal struct with enums", func(t *testing.T) {
		order := OrderPlacement{
			Book:  *NewBook(BTC, MXN),
			Side:  OrderSideBuy,
			Type:  OrderTypeLimit,
			Major: ToMonetary(0.1),
			Price: ToMonetary(500000),
		}

		data, err := json.Marshal(order)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"side":"buy"`)
		assert.Contains(t, string(data), `"type":"limit"`)
	})

	t.Run("unmarshal struct with enums", func(t *testing.T) {
		jsonData := `{
			"book": "btc_mxn",
			"original_amount": "0.1",
			"unfilled_amount": "0.0",
			"original_value": "50000.00",
			"created_at": "2024-01-15T10:30:00+00:00",
			"updated_at": "2024-01-15T10:30:00+00:00",
			"price": "500000.00",
			"oid": "order123",
			"side": "sell",
			"status": "completed",
			"type": "market"
		}`

		var order UserOrder
		err := json.Unmarshal([]byte(jsonData), &order)

		require.NoError(t, err)
		assert.Equal(t, OrderSideSell, order.Side)
		assert.Equal(t, OrderStatusCompleted, order.Status)
	})
}
