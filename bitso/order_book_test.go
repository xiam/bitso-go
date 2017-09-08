package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderBook(t *testing.T) {
	payload := []byte(`{
    "success": true,
    "payload": {
        "asks": [{
            "book": "btc_mxn",
            "price": "5632.24",
            "amount": "1.34491802"
        },{
            "book": "btc_mxn",
            "price": "5633.44",
            "amount": "0.4259"
        },{
            "book": "btc_mxn",
            "price": "5642.14",
            "amount": "1.21642"
        }],
        "bids": [{
            "book": "btc_mxn",
            "price": "6123.55",
            "amount": "1.12560000"
        },{
            "book": "btc_mxn",
            "price": "6121.55",
            "amount": "2.23976"
        }],
        "updated_at": "2016-04-08T17:52:31.000+00:00",
        "sequence": "27214"
    }
}`)
	var orderBookResponse OrderBookResponse
	err := json.Unmarshal(payload, &orderBookResponse)
	assert.NoError(t, err)

	assert.True(t, orderBookResponse.Success)
	assert.NotNil(t, orderBookResponse.Payload)
	assert.True(t, len(orderBookResponse.Payload.Bids) > 0)
	assert.True(t, len(orderBookResponse.Payload.Asks) > 0)
}
