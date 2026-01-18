package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceJSON(t *testing.T) {
	jsonData := `{
		"currency": "btc",
		"total": "1.5",
		"locked": "0.5",
		"available": "1.0",
		"pending_deposit": "0.1",
		"pending_withdrawal": "0.0"
	}`

	var balance Balance
	err := json.Unmarshal([]byte(jsonData), &balance)

	require.NoError(t, err)
	assert.Equal(t, Currency(BTC), balance.Currency)
	assert.Equal(t, "1.5", string(balance.Total))
	assert.Equal(t, "0.5", string(balance.Locked))
	assert.Equal(t, "1.0", string(balance.Available))
	assert.Equal(t, "0.1", string(balance.PendingDeposit))
	assert.Equal(t, "0.0", string(balance.PendingWithdrawal))
}

func TestTickerJSON(t *testing.T) {
	t.Run("unmarshal", func(t *testing.T) {
		jsonData := `{
			"book": "btc_mxn",
			"volume": "1000.5",
			"high": "550000.00",
			"last": "520000.00",
			"low": "480000.00",
			"vwap": "510000.00",
			"ask": "520100.00",
			"bid": "519900.00",
			"change_24": "14060",
			"rolling_average_change": {"6": "0.0919"},
			"created_at": "2024-01-15T10:30:00+00:00"
		}`

		var ticker Ticker
		err := json.Unmarshal([]byte(jsonData), &ticker)

		require.NoError(t, err)
		assert.Equal(t, "btc_mxn", ticker.Book.String())
		assert.Equal(t, "1000.5", string(ticker.Volume))
		assert.Equal(t, "550000.00", string(ticker.High))
		assert.Equal(t, "520000.00", string(ticker.Last))
		assert.Equal(t, "480000.00", string(ticker.Low))
		assert.Equal(t, "510000.00", string(ticker.Vwap))
		assert.Equal(t, "520100.00", string(ticker.Ask))
		assert.Equal(t, "519900.00", string(ticker.Bid))
		assert.Equal(t, "14060", string(ticker.Change24))
		assert.Equal(t, "0.0919", string(ticker.RollingAverageChange["6"]))
	})

	t.Run("marshal", func(t *testing.T) {
		ticker := Ticker{
			Book:     *NewBook(BTC, MXN),
			Volume:   ToMonetary(1000.5),
			High:     ToMonetary(550000),
			Change24: ToMonetary(14060),
			RollingAverageChange: map[string]Monetary{
				"6": "0.0919",
			},
		}

		data, err := json.Marshal(ticker)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"book":"btc_mxn"`)
		assert.Contains(t, string(data), `"change_24"`)
		assert.Contains(t, string(data), `"rolling_average_change"`)
	})
}

func TestTradeJSON(t *testing.T) {
	t.Run("with integer tid", func(t *testing.T) {
		jsonData := `{
			"book": "eth_mxn",
			"created_at": "2024-01-15T10:30:00+00:00",
			"amount": "2.5",
			"maker_side": "buy",
			"price": "35000.00",
			"tid": 12345
		}`

		var trade Trade
		err := json.Unmarshal([]byte(jsonData), &trade)

		require.NoError(t, err)
		assert.Equal(t, "eth_mxn", trade.Book.String())
		assert.Equal(t, "2.5", string(trade.Amount))
		assert.Equal(t, OrderSideBuy, trade.MakerSide)
		assert.Equal(t, "35000.00", string(trade.Price))
		assert.Equal(t, uint64(12345), trade.TID.Uint64())
	})

	t.Run("with string tid", func(t *testing.T) {
		jsonData := `{
			"book": "eth_mxn",
			"created_at": "2024-01-15T10:30:00+00:00",
			"amount": "2.5",
			"maker_side": "sell",
			"price": "35000.00",
			"tid": "67890"
		}`

		var trade Trade
		err := json.Unmarshal([]byte(jsonData), &trade)

		require.NoError(t, err)
		assert.Equal(t, OrderSideSell, trade.MakerSide)
		assert.Equal(t, uint64(67890), trade.TID.Uint64())
	})
}

func TestUserTradeJSON(t *testing.T) {
	jsonData := `{
		"book": "btc_mxn",
		"major": "-0.1",
		"created_at": "2024-01-15T10:30:00+00:00",
		"minor": "50000.00",
		"fees_amount": "325.00",
		"fees_currency": "mxn",
		"price": "500000.00",
		"tid": 12345,
		"oid": "order123",
		"side": "sell"
	}`

	var trade UserTrade
	err := json.Unmarshal([]byte(jsonData), &trade)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", trade.Book.String())
	assert.Equal(t, "-0.1", string(trade.Major))
	assert.Equal(t, "50000.00", string(trade.Minor))
	assert.Equal(t, "325.00", string(trade.FeesAmount))
	assert.Equal(t, Currency(MXN), trade.FeesCurrency)
	assert.Equal(t, "500000.00", string(trade.Price))
	assert.Equal(t, uint64(12345), trade.TID.Uint64())
	assert.Equal(t, "order123", trade.OID)
	assert.Equal(t, OrderSideSell, trade.Side)
}

func TestUserOrderTradeJSON(t *testing.T) {
	jsonData := `{
		"book": "btc_mxn",
		"major": "-0.05",
		"created_at": "2024-01-15T10:30:00+00:00",
		"minor": "25000.00",
		"fees_amount": "162.50",
		"currency": "mxn",
		"price": "500000.00",
		"tid": 12346,
		"oid": "order123",
		"side": "sell"
	}`

	var trade UserOrderTrade
	err := json.Unmarshal([]byte(jsonData), &trade)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", trade.Book.String())
	assert.Equal(t, Currency(MXN), trade.FeesCurrency)
	assert.Equal(t, "order123", trade.OID)
}

func TestFeeJSON(t *testing.T) {
	jsonData := `{
		"book": "btc_mxn",
		"fee_decimal": "0.0065",
		"fee_percent": "0.65"
	}`

	var fee Fee
	err := json.Unmarshal([]byte(jsonData), &fee)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", fee.Book.String())
	assert.Equal(t, "0.0065", string(fee.FeeDecimal))
	assert.Equal(t, "0.65", string(fee.FeePercent))
}

func TestCustomerFeesJSON(t *testing.T) {
	jsonData := `{
		"fees": [
			{"book": "btc_mxn", "fee_decimal": "0.0065", "fee_percent": "0.65"},
			{"book": "eth_mxn", "fee_decimal": "0.0065", "fee_percent": "0.65"}
		],
		"withdrawal_fees": {
			"btc": "0.0001",
			"eth": "0.005",
			"mxn": "0.00"
		}
	}`

	var fees CustomerFees
	err := json.Unmarshal([]byte(jsonData), &fees)

	require.NoError(t, err)
	assert.Len(t, fees.Fees, 2)
	assert.Equal(t, "btc_mxn", fees.Fees[0].Book.String())
	assert.Contains(t, fees.WithdrawalFees, "btc")
	assert.Contains(t, fees.WithdrawalFees, "eth")
	assert.Contains(t, fees.WithdrawalFees, "mxn")
	assert.Equal(t, "0.0001", string(fees.WithdrawalFees["btc"]))
}

func TestFundingJSON(t *testing.T) {
	jsonData := `{
		"fid": "fund123",
		"currency": "btc",
		"method": "Bitcoin Network",
		"amount": "0.5",
		"status": "complete",
		"created_at": "2024-01-15T10:30:00+00:00",
		"details": {
			"txid": "abc123def456",
			"confirmations": 6
		}
	}`

	var funding Funding
	err := json.Unmarshal([]byte(jsonData), &funding)

	require.NoError(t, err)
	assert.Equal(t, "fund123", funding.FID)
	assert.Equal(t, Currency(BTC), funding.Currency)
	assert.Equal(t, "Bitcoin Network", funding.Method)
	assert.Equal(t, "0.5", string(funding.Amount))
	assert.Equal(t, "complete", funding.Status)
	assert.NotNil(t, funding.Details)
	assert.Equal(t, "abc123def456", funding.Details["txid"])
}

func TestWithdrawalJSON(t *testing.T) {
	jsonData := `{
		"wid": "withdraw123",
		"currency": "mxn",
		"method": "SPEI",
		"amount": "10000.00",
		"status": "pending",
		"created_at": "2024-01-15T10:30:00+00:00",
		"details": {
			"clabe": "1234567890123456",
			"beneficiary": "John Doe"
		}
	}`

	var withdrawal Withdrawal
	err := json.Unmarshal([]byte(jsonData), &withdrawal)

	require.NoError(t, err)
	assert.Equal(t, "withdraw123", withdrawal.WID)
	assert.Equal(t, Currency(MXN), withdrawal.Currency)
	assert.Equal(t, "SPEI", withdrawal.Method)
	assert.Equal(t, "10000.00", string(withdrawal.Amount))
	assert.Equal(t, "pending", withdrawal.Status)
}

func TestExchangeOrderBookJSON(t *testing.T) {
	jsonData := `{
		"book": "btc_mxn",
		"default_chart": "depth",
		"minimum_amount": "0.001",
		"maximum_amount": "1000.0",
		"minimum_price": "100.0",
		"maximum_price": "10000000.0",
		"minimum_value": "10.0",
		"maximum_value": "1000000.0",
		"tick_size": "0.01",
		"fees": {
			"flat_rate": {
				"maker": "0.003",
				"taker": "0.0036"
			},
			"structure": [
				{"volume": "1000", "maker": "0.003", "taker": "0.0036"},
				{"volume": "5000", "maker": "0.00239", "taker": "0.00333"}
			]
		}
	}`

	var book ExchangeOrderBook
	err := json.Unmarshal([]byte(jsonData), &book)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", book.Book.String())
	assert.Equal(t, "depth", book.DefaultChart)
	assert.Equal(t, "0.001", string(book.MinimumAmount))
	assert.Equal(t, "1000.0", string(book.MaximumAmount))
	assert.Equal(t, "100.0", string(book.MinimumPrice))
	assert.Equal(t, "10000000.0", string(book.MaximumPrice))
	assert.Equal(t, "10.0", string(book.MinimumValue))
	assert.Equal(t, "1000000.0", string(book.MaximumValue))
	assert.Equal(t, "0.01", string(book.TickSize))

	// Fees
	assert.Equal(t, "0.003", string(book.Fees.FlatRate.Maker))
	assert.Equal(t, "0.0036", string(book.Fees.FlatRate.Taker))
	require.Len(t, book.Fees.Structure, 2)
	assert.Equal(t, "1000", string(book.Fees.Structure[0].Volume))
	assert.Equal(t, "0.003", string(book.Fees.Structure[0].Maker))
	assert.Equal(t, "0.0036", string(book.Fees.Structure[0].Taker))
}

func TestBookFeesJSON(t *testing.T) {
	jsonData := `{
		"flat_rate": {
			"maker": "0.003",
			"taker": "0.0036"
		},
		"structure": [
			{"volume": "1000", "maker": "0.003", "taker": "0.0036"},
			{"volume": "5000", "maker": "0.00239", "taker": "0.00333"},
			{"volume": "10000", "maker": "0.00205", "taker": "0.00282"}
		]
	}`

	var fees BookFees
	err := json.Unmarshal([]byte(jsonData), &fees)

	require.NoError(t, err)
	assert.Equal(t, "0.003", string(fees.FlatRate.Maker))
	assert.Equal(t, "0.0036", string(fees.FlatRate.Taker))
	require.Len(t, fees.Structure, 3)
	assert.Equal(t, "10000", string(fees.Structure[2].Volume))
}

func TestOrderJSON(t *testing.T) {
	jsonData := `{
		"book": "btc_mxn",
		"price": "500000.00",
		"amount": "0.5",
		"oid": "order123"
	}`

	var order Order
	err := json.Unmarshal([]byte(jsonData), &order)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", order.Book.String())
	assert.Equal(t, "500000.00", string(order.Price))
	assert.Equal(t, "0.5", string(order.Amount))
	assert.Equal(t, "order123", order.OID)
}

func TestOrderBookJSON(t *testing.T) {
	jsonData := `{
		"asks": [
			{"book": "btc_mxn", "price": "500100.00", "amount": "0.5"},
			{"book": "btc_mxn", "price": "500200.00", "amount": "1.0"}
		],
		"bids": [
			{"book": "btc_mxn", "price": "499900.00", "amount": "0.3"},
			{"book": "btc_mxn", "price": "499800.00", "amount": "0.8"}
		],
		"updated_at": "2024-01-15T10:30:00+00:00",
		"sequence": "12345"
	}`

	var orderBook OrderBook
	err := json.Unmarshal([]byte(jsonData), &orderBook)

	require.NoError(t, err)
	assert.Len(t, orderBook.Asks, 2)
	assert.Len(t, orderBook.Bids, 2)
	assert.Equal(t, "12345", orderBook.Sequence)
	assert.Equal(t, "500100.00", string(orderBook.Asks[0].Price))
	assert.Equal(t, "499900.00", string(orderBook.Bids[0].Price))
}

func TestUserOrderJSON(t *testing.T) {
	t.Run("complete order", func(t *testing.T) {
		jsonData := `{
			"book": "btc_mxn",
			"original_amount": "0.1",
			"unfilled_amount": "0.0",
			"original_value": "50000.00",
			"created_at": "2024-01-15T10:30:00+00:00",
			"updated_at": "2024-01-15T10:35:00+00:00",
			"price": "500000.00",
			"oid": "order123",
			"side": "buy",
			"status": "completed",
			"type": "limit"
		}`

		var order UserOrder
		err := json.Unmarshal([]byte(jsonData), &order)

		require.NoError(t, err)
		assert.Equal(t, "btc_mxn", order.Book.String())
		assert.Equal(t, "0.1", string(order.OriginalAmount))
		assert.Equal(t, "0.0", string(order.UnfilledAmount))
		assert.Equal(t, "order123", order.OID)
		assert.Equal(t, OrderSideBuy, order.Side)
		assert.Equal(t, OrderStatusCompleted, order.Status)
		assert.Equal(t, "limit", order.Type)
	})

	t.Run("partially filled order", func(t *testing.T) {
		jsonData := `{
			"book": "eth_mxn",
			"original_amount": "5.0",
			"unfilled_amount": "2.5",
			"original_value": "175000.00",
			"created_at": "2024-01-15T10:30:00+00:00",
			"updated_at": "2024-01-15T10:35:00+00:00",
			"price": "35000.00",
			"oid": "order456",
			"side": "sell",
			"status": "partially filled",
			"type": "limit"
		}`

		var order UserOrder
		err := json.Unmarshal([]byte(jsonData), &order)

		require.NoError(t, err)
		assert.Equal(t, OrderStatusPartialFill, order.Status)
		assert.Equal(t, OrderSideSell, order.Side)
		assert.Equal(t, "2.5", string(order.UnfilledAmount))
	})
}

func TestOrderPlacementJSON(t *testing.T) {
	t.Run("limit order", func(t *testing.T) {
		order := OrderPlacement{
			Book:  *NewBook(BTC, MXN),
			Side:  OrderSideBuy,
			Type:  OrderTypeLimit,
			Major: ToMonetary(0.1),
			Price: ToMonetary(500000),
		}

		data, err := json.Marshal(order)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"book":"btc_mxn"`)
		assert.Contains(t, string(data), `"side":"buy"`)
		assert.Contains(t, string(data), `"type":"limit"`)
		assert.Contains(t, string(data), `"major":"0.1`)
		assert.Contains(t, string(data), `"price":"500000`)
	})

	t.Run("market order", func(t *testing.T) {
		order := OrderPlacement{
			Book:  *NewBook(ETH, MXN),
			Side:  OrderSideSell,
			Type:  OrderTypeMarket,
			Major: ToMonetary(2.0),
		}

		data, err := json.Marshal(order)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"type":"market"`)
		assert.Contains(t, string(data), `"side":"sell"`)
	})

	t.Run("unmarshal", func(t *testing.T) {
		jsonData := `{
			"book": "btc_mxn",
			"side": "buy",
			"type": "limit",
			"major": "0.5",
			"price": "480000.00"
		}`

		var order OrderPlacement
		err := json.Unmarshal([]byte(jsonData), &order)

		require.NoError(t, err)
		assert.Equal(t, "btc_mxn", order.Book.String())
		assert.Equal(t, OrderSideBuy, order.Side)
		assert.Equal(t, OrderTypeLimit, order.Type)
	})
}

func TestTransactionJSON(t *testing.T) {
	jsonData := `{
		"eid": "txn123",
		"operation": "trade",
		"created_at": "2024-01-15T10:30:00+00:00",
		"balance_updates": [
			{"currency": "btc", "amount": "-0.1"},
			{"currency": "mxn", "amount": "50000.00"}
		],
		"details": {
			"tid": "12345",
			"oid": "order123"
		}
	}`

	var txn Transaction
	err := json.Unmarshal([]byte(jsonData), &txn)

	require.NoError(t, err)
	assert.Equal(t, "txn123", txn.EID)
	assert.Equal(t, OperationTrade, txn.Operation)
	require.Len(t, txn.BalanceUpdates, 2)
	assert.Equal(t, Currency(BTC), txn.BalanceUpdates[0].Currency)
	assert.Equal(t, "-0.1", string(txn.BalanceUpdates[0].Amount))
	assert.Equal(t, Currency(MXN), txn.BalanceUpdates[1].Currency)
	assert.NotNil(t, txn.Details)
	assert.Equal(t, "12345", txn.Details["tid"])
}

func TestEnvelopeJSON(t *testing.T) {
	t.Run("success envelope", func(t *testing.T) {
		jsonData := `{
			"success": true,
			"payload": {"key": "value"}
		}`

		var env Envelope
		err := json.Unmarshal([]byte(jsonData), &env)

		require.NoError(t, err)
		assert.True(t, env.Success)
		// Error struct is always present but Code will be nil for success
		assert.Nil(t, env.Error.Code)
		assert.Empty(t, env.Error.Message)
	})

	t.Run("error envelope", func(t *testing.T) {
		jsonData := `{
			"success": false,
			"error": {
				"code": 101,
				"message": "Invalid API key"
			}
		}`

		var env Envelope
		err := json.Unmarshal([]byte(jsonData), &env)

		require.NoError(t, err)
		assert.False(t, env.Success)
		// Code is interface{} so JSON numbers are float64
		assert.Equal(t, float64(101), env.Error.Code)
		assert.Equal(t, "Invalid API key", env.Error.Message)
	})
}

func TestErrorJSON(t *testing.T) {
	jsonData := `{
		"code": 303,
		"message": "The field book is missing"
	}`

	var bitsoErr struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	err := json.Unmarshal([]byte(jsonData), &bitsoErr)

	require.NoError(t, err)
	assert.Equal(t, 303, bitsoErr.Code)
	assert.Equal(t, "The field book is missing", bitsoErr.Message)
}

func TestEmptyPayloads(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		jsonData := `{"success": true, "payload": []}`

		var result struct {
			Success bool    `json:"success"`
			Payload []Trade `json:"payload"`
		}
		err := json.Unmarshal([]byte(jsonData), &result)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Empty(t, result.Payload)
	})

	t.Run("null payload", func(t *testing.T) {
		jsonData := `{"success": true, "payload": null}`

		var result struct {
			Success bool    `json:"success"`
			Payload []Trade `json:"payload"`
		}
		err := json.Unmarshal([]byte(jsonData), &result)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Nil(t, result.Payload)
	})
}
