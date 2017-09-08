package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvailableBooks(t *testing.T) {
	payload := []byte(`{
    "success": true,
    "payload": [{
        "book": "btc_mxn",
        "minimum_amount": ".003",
        "maximum_amount": "1000.00",
        "minimum_price": "100.00",
        "maximum_price": "1000000.00",
        "minimum_value": "25.00",
        "maximum_value": "1000000.00"
    }, {
        "book": "eth_mxn",
        "minimum_amount": ".003",
        "maximum_amount": "1000.00",
        "minimum_price": "100.0",
        "maximum_price": "1000000.0",
        "minimum_value": "25.0",
        "maximum_value": "1000000.0"
    }]
}`)
	var availableBooksResponse AvailableBooksResponse
	err := json.Unmarshal(payload, &availableBooksResponse)
	assert.NoError(t, err)

	assert.True(t, availableBooksResponse.Success)
	assert.True(t, len(availableBooksResponse.Payload) > 0)
	assert.Equal(t, BTC, availableBooksResponse.Payload[0].Book.Major())
	assert.Equal(t, MXN, availableBooksResponse.Payload[0].Book.Minor())
	assert.True(t, availableBooksResponse.Payload[0].MinimumPrice.Float64() == 100.0)
}
