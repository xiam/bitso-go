package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrade(t *testing.T) {
	payload := []byte(`{
    "success": true,
    "payload": [{
        "book": "btc_mxn",
        "created_at": "2016-04-08T17:52:31.000+00:00",
        "amount": "0.02000000",
        "maker_side": "buy",
        "price": "5545.01",
        "tid": 55845
    }, {
        "book": "btc_mxn",
        "created_at": "2016-04-08T17:52:31.000+00:00",
        "amount": "0.33723939",
        "maker_side": "sell",
        "price": "5633.98",
        "tid": 55844
    }]
}`)
	var tradeResponse TradeResponse
	err := json.Unmarshal(payload, &tradeResponse)
	assert.NoError(t, err)

	assert.True(t, tradeResponse.Success)
	assert.Equal(t, BTC, tradeResponse.Payload[0].Book.Major())
	assert.Equal(t, MXN, tradeResponse.Payload[0].Book.Minor())
	assert.True(t, tradeResponse.Payload[0].Amount.Float64() > 0.001)
}
