package bitso

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrency_SQLScanSupportedValues(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var currency Currency

		err := currency.Scan("ETH")

		require.NoError(t, err)
		assert.Equal(t, Currency(ETH), currency)
	})

	t.Run("bytes", func(t *testing.T) {
		var currency Currency

		err := currency.Scan([]byte("MXN"))

		require.NoError(t, err)
		assert.Equal(t, Currency(MXN), currency)
	})

	t.Run("nil preserves value", func(t *testing.T) {
		currency := Currency(BTC)

		err := currency.Scan(nil)

		require.NoError(t, err)
		assert.Equal(t, Currency(BTC), currency)
	})
}

func TestCurrency_SQLScanUnsupportedType(t *testing.T) {
	var currency Currency

	err := currency.Scan(123)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot scan int into Currency")
}

func TestBook_SQLScanSupportedValues(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var book Book

		err := book.Scan("btc_mxn")

		require.NoError(t, err)
		assert.Equal(t, Currency(BTC), book.Major())
		assert.Equal(t, Currency(MXN), book.Minor())
	})

	t.Run("bytes", func(t *testing.T) {
		var book Book

		err := book.Scan([]byte("eth_usd"))

		require.NoError(t, err)
		assert.Equal(t, Currency(ETH), book.Major())
		assert.Equal(t, Currency(USD), book.Minor())
	})

	t.Run("nil preserves value", func(t *testing.T) {
		book := NewBook(SOL, BRL)

		err := book.Scan(nil)

		require.NoError(t, err)
		assert.Equal(t, Currency(SOL), book.Major())
		assert.Equal(t, Currency(BRL), book.Minor())
	})
}

func TestBook_SQLScanInvalidAndUnsupportedValues(t *testing.T) {
	t.Run("invalid string", func(t *testing.T) {
		var book Book

		err := book.Scan("btcmxn")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected book format")
	})

	t.Run("unsupported type", func(t *testing.T) {
		var book Book

		err := book.Scan(123)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot scan int into Book")
	})
}

func TestOrderSide_SQLScanSupportedValues(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var side OrderSide

		err := side.Scan("buy")

		require.NoError(t, err)
		assert.Equal(t, OrderSideBuy, side)
	})

	t.Run("bytes", func(t *testing.T) {
		var side OrderSide

		err := side.Scan([]byte("sell"))

		require.NoError(t, err)
		assert.Equal(t, OrderSideSell, side)
	})

	t.Run("nil preserves value", func(t *testing.T) {
		side := OrderSideSell

		err := side.Scan(nil)

		require.NoError(t, err)
		assert.Equal(t, OrderSideSell, side)
	})
}

func TestOrderSide_SQLScanInvalidAndUnsupportedValues(t *testing.T) {
	t.Run("invalid string", func(t *testing.T) {
		var side OrderSide

		err := side.Scan("hold")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported order side")
	})

	t.Run("unsupported type", func(t *testing.T) {
		var side OrderSide

		err := side.Scan(123)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot scan int into OrderSide")
	})
}

func TestOrderStatus_SQLScanSupportedValues(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var status OrderStatus

		err := status.Scan("open")

		require.NoError(t, err)
		assert.Equal(t, OrderStatusOpen, status)
	})

	t.Run("bytes", func(t *testing.T) {
		var status OrderStatus

		err := status.Scan([]byte("partially filled"))

		require.NoError(t, err)
		assert.Equal(t, OrderStatusPartialFill, status)
	})

	t.Run("nil preserves value", func(t *testing.T) {
		status := OrderStatusCompleted

		err := status.Scan(nil)

		require.NoError(t, err)
		assert.Equal(t, OrderStatusCompleted, status)
	})
}

func TestOrderStatus_SQLScanInvalidAndUnsupportedValues(t *testing.T) {
	t.Run("invalid string", func(t *testing.T) {
		var status OrderStatus

		err := status.Scan("settled")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported status")
	})

	t.Run("unsupported type", func(t *testing.T) {
		var status OrderStatus

		err := status.Scan(123)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot scan int into OrderStatus")
	})
}

func TestOperation_SQLScanSupportedValues(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var op Operation

		err := op.Scan("funding")

		require.NoError(t, err)
		assert.Equal(t, OperationFunding, op)
	})

	t.Run("bytes", func(t *testing.T) {
		var op Operation

		err := op.Scan([]byte("withdrawal"))

		require.NoError(t, err)
		assert.Equal(t, OperationWithdrawal, op)
	})

	t.Run("nil preserves value", func(t *testing.T) {
		op := OperationTrade

		err := op.Scan(nil)

		require.NoError(t, err)
		assert.Equal(t, OperationTrade, op)
	})
}

func TestOperation_SQLScanInvalidAndUnsupportedValues(t *testing.T) {
	t.Run("invalid string", func(t *testing.T) {
		var op Operation

		err := op.Scan("rebate")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported operation")
	})

	t.Run("unsupported type", func(t *testing.T) {
		var op Operation

		err := op.Scan(123)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot scan int into Operation")
	})
}
