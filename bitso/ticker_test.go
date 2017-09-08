package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTicker(t *testing.T) {
	payload := []byte(`{
    "success": true,
    "payload": {
        "book": "btc_mxn",
        "volume": "22.31349615",
        "high": "5750.00",
        "last": "5633.98",
        "low": "5450.00",
        "vwap": "5393.45",
        "ask": "5632.24",
        "bid": "5520.01",
        "created_at": "2016-04-08T17:52:31.000+00:00"
    }
}`)
	var tickerResponse TickerResponse
	err := json.Unmarshal(payload, &tickerResponse)
	assert.NoError(t, err)

	assert.True(t, tickerResponse.Success)
	assert.Equal(t, BTC, tickerResponse.Payload.Book.Major())
	assert.Equal(t, MXN, tickerResponse.Payload.Book.Minor())
	assert.True(t, tickerResponse.Payload.High.Float64() > 5000.0)
}
