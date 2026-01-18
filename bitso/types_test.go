package bitso

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Monetary tests

func TestMonetary_String(t *testing.T) {
	m := Monetary("123.456")
	assert.Equal(t, "123.456", string(m))
}

func TestMonetary_Float64(t *testing.T) {
	tests := []struct {
		monetary Monetary
		expected float64
	}{
		{"100.5", 100.5},
		{"0.00001", 0.00001},
		{"-50.25", -50.25},
		{"0", 0},
		{"1000000.123456", 1000000.123456},
	}

	for _, tc := range tests {
		t.Run(string(tc.monetary), func(t *testing.T) {
			assert.InDelta(t, tc.expected, tc.monetary.Float64(), 0.0000001)
		})
	}
}

func TestMonetary_Float64_Invalid(t *testing.T) {
	m := Monetary("not-a-number")
	result := m.Float64()
	assert.Equal(t, float64(0), result)
}

func TestMonetary_Decimal(t *testing.T) {
	t.Run("valid decimal", func(t *testing.T) {
		m := Monetary("123.456789")
		d, err := m.Decimal()

		require.NoError(t, err)
		expected, _ := decimal.NewFromString("123.456789")
		assert.True(t, d.Equal(expected))
	})

	t.Run("large decimal", func(t *testing.T) {
		m := Monetary("9999999999.999999999999")
		d, err := m.Decimal()

		require.NoError(t, err)
		assert.False(t, d.IsZero())
	})

	t.Run("invalid decimal", func(t *testing.T) {
		m := Monetary("invalid")
		_, err := m.Decimal()

		require.Error(t, err)
	})
}

func TestToMonetary(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{100.5, "100.500000"},
		{0.1, "0.100000"},
		{1000000, "1000000.000000"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			m := ToMonetary(tc.input)
			assert.Equal(t, tc.expected, string(m))
		})
	}
}

func TestMonetary_JSONRoundtrip(t *testing.T) {
	type wrapper struct {
		Value Monetary `json:"value"`
	}

	original := wrapper{Value: Monetary("123.456")}

	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"value":"123.456"`)

	var decoded wrapper
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original.Value, decoded.Value)
}

// Time tests

func TestTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected time.Time
	}{
		{
			name:     "format with colon in timezone",
			json:     `"2024-01-15T10:30:00+00:00"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:     "format without colon in timezone",
			json:     `"2024-01-15T10:30:00-0600"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.FixedZone("", -6*60*60)),
		},
		{
			name:     "format with milliseconds",
			json:     `"2024-01-15T10:30:00.123-06:00"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 123000000, time.FixedZone("", -6*60*60)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result Time
			err := json.Unmarshal([]byte(tc.json), &result)

			require.NoError(t, err)
			assert.True(t, tc.expected.Equal(result.Time()))
		})
	}
}

func TestTime_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"invalid format", `"not-a-time"`},
		{"wrong structure", `12345`},
		{"empty string", `""`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result Time
			err := json.Unmarshal([]byte(tc.json), &result)
			require.Error(t, err)
		})
	}
}

func TestTime_String(t *testing.T) {
	tm := Time(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	result := tm.String()

	assert.Equal(t, "2024-01-15T10:30:00+0000", result)
}

func TestTime_Time(t *testing.T) {
	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tm := Time(expected)

	result := tm.Time()

	assert.Equal(t, expected, result)
}

// TID tests

func TestTID_UnmarshalJSON_Integer(t *testing.T) {
	jsonData := `12345`

	var tid TID
	err := json.Unmarshal([]byte(jsonData), &tid)

	require.NoError(t, err)
	assert.Equal(t, uint64(12345), tid.Uint64())
}

func TestTID_UnmarshalJSON_String(t *testing.T) {
	jsonData := `"67890"`

	var tid TID
	err := json.Unmarshal([]byte(jsonData), &tid)

	require.NoError(t, err)
	assert.Equal(t, uint64(67890), tid.Uint64())
}

func TestTID_UnmarshalJSON_LargeNumber(t *testing.T) {
	jsonData := `"18446744073709551615"` // max uint64

	var tid TID
	err := json.Unmarshal([]byte(jsonData), &tid)

	require.NoError(t, err)
	assert.Equal(t, uint64(18446744073709551615), tid.Uint64())
}

func TestTID_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"invalid string", `"not-a-number"`},
		{"negative number", `"-123"`},
		{"float", `"123.456"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tid TID
			err := json.Unmarshal([]byte(tc.json), &tid)
			require.Error(t, err)
		})
	}
}

func TestTID_Uint64(t *testing.T) {
	t.Run("normal value", func(t *testing.T) {
		tid := TID(12345)
		assert.Equal(t, uint64(12345), tid.Uint64())
	})

	t.Run("nil pointer", func(t *testing.T) {
		var tid *TID
		assert.Equal(t, uint64(0), tid.Uint64())
	})

	t.Run("zero value", func(t *testing.T) {
		tid := TID(0)
		assert.Equal(t, uint64(0), tid.Uint64())
	})
}

// Book tests

func TestNewBook(t *testing.T) {
	book := NewBook(BTC, MXN)

	require.NotNil(t, book)
	assert.Equal(t, Currency(BTC), book.Major())
	assert.Equal(t, Currency(MXN), book.Minor())
}

func TestBook_String(t *testing.T) {
	tests := []struct {
		major    Currency
		minor    Currency
		expected string
	}{
		{BTC, MXN, "btc_mxn"},
		{ETH, USD, "eth_usd"},
		{XRP, BRL, "xrp_brl"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			book := NewBook(tc.major, tc.minor)
			assert.Equal(t, tc.expected, book.String())
		})
	}
}

func TestBook_Major(t *testing.T) {
	book := NewBook(ETH, MXN)
	assert.Equal(t, Currency(ETH), book.Major())
}

func TestBook_Minor(t *testing.T) {
	book := NewBook(ETH, MXN)
	assert.Equal(t, Currency(MXN), book.Minor())
}

func TestBook_MarshalJSON(t *testing.T) {
	book := NewBook(BTC, MXN)

	data, err := json.Marshal(book)

	require.NoError(t, err)
	assert.Equal(t, `"btc_mxn"`, string(data))
}

func TestBook_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json          string
		expectedMajor Currency
		expectedMinor Currency
	}{
		{`"btc_mxn"`, BTC, MXN},
		{`"eth_usd"`, ETH, USD},
		{`"sol_brl"`, SOL, BRL},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var book Book
			err := json.Unmarshal([]byte(tc.json), &book)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedMajor, book.Major())
			assert.Equal(t, tc.expectedMinor, book.Minor())
		})
	}
}

func TestBook_UnmarshalJSON_Invalid(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"missing underscore", `"btcmxn"`},
		{"too many parts", `"btc_mxn_extra"`},
		{"empty", `""`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var book Book
			err := json.Unmarshal([]byte(tc.json), &book)
			require.Error(t, err)
		})
	}
}

func TestBook_SQLValue(t *testing.T) {
	book := NewBook(BTC, MXN)

	value, err := book.Value()

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", value)
}

func TestBook_SQLScan(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		var book Book
		err := book.Scan("eth_mxn")

		require.NoError(t, err)
		assert.Equal(t, Currency(ETH), book.Major())
		assert.Equal(t, Currency(MXN), book.Minor())
	})

	t.Run("nil value", func(t *testing.T) {
		var book Book
		err := book.Scan(nil)

		require.NoError(t, err)
	})
}

func TestBook_JSONRoundtrip(t *testing.T) {
	original := NewBook(SOL, USD)

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Book
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Major(), decoded.Major())
	assert.Equal(t, original.Minor(), decoded.Minor())
}

// Currency tests

func TestToCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected Currency
	}{
		{"BTC", BTC},
		{"btc", BTC},
		{"Btc", BTC},
		{"mxn", MXN},
		{"ETH", ETH},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ToCurrency(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCurrency_String(t *testing.T) {
	assert.Equal(t, "btc", Currency(BTC).String())
	assert.Equal(t, "mxn", Currency(MXN).String())
	assert.Equal(t, "eth", Currency(ETH).String())
}

func TestCurrency_MarshalJSON(t *testing.T) {
	data, err := json.Marshal(BTC)

	require.NoError(t, err)
	assert.Equal(t, `"btc"`, string(data))
}

func TestCurrency_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected Currency
	}{
		{`"btc"`, BTC},
		{`"BTC"`, BTC}, // should be normalized to lowercase
		{`"mxn"`, MXN},
		{`"ETH"`, ETH},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			var currency Currency
			err := json.Unmarshal([]byte(tc.json), &currency)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, currency)
		})
	}
}

func TestCurrency_SQLValue(t *testing.T) {
	currency := Currency(BTC)
	value, err := currency.Value()

	require.NoError(t, err)
	assert.Equal(t, "btc", value)
}

func TestCurrency_SQLScan(t *testing.T) {
	var currency Currency
	err := currency.Scan("eth")

	require.NoError(t, err)
	assert.Equal(t, Currency(ETH), currency)
}

func TestCurrency_Constants(t *testing.T) {
	// Verify some key currency constants are defined correctly
	assert.Equal(t, "btc", string(BTC))
	assert.Equal(t, "eth", string(ETH))
	assert.Equal(t, "mxn", string(MXN))
	assert.Equal(t, "usd", string(USD))
	assert.Equal(t, "brl", string(BRL))
	assert.Equal(t, "usdt", string(USDT))
	assert.Equal(t, "sol", string(SOL))
	assert.Equal(t, "doge", string(DOGE))
}

func TestCurrencyNone(t *testing.T) {
	assert.Equal(t, "", CurrencyNone.String())
}
