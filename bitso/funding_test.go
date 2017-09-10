package bitso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFundings(t *testing.T) {
	payload := []byte(`{
    "success": true,
    "payload": [{
        "fid": "c5b8d7f0768ee91d3b33bee648318688",
        "status": "pending",
        "created_at": "2016-04-08T17:52:31.000+00:00",
        "currency": "btc",
        "method": "btc",
        "amount": "0.48650929",
        "details": {
            "funding_address": "18MsnATiNiKLqUHDTRKjurwMg7inCrdNEp",
            "tx_hash": "d4f28394693e9fb5fffcaf730c11f32d1922e5837f76ca82189d3bfe30ded433"
        }
    }, {
        "fid": "p4u8d7f0768ee91d3b33bee6483132i8",
        "status": "complete",
        "created_at": "2016-04-08T17:52:31.000+00:00",
        "currency": "mxn",
        "method": "sp",
        "amount": "300.15",
        "details": {
            "sender_name": "BERTRAND RUSSELL",
            "sender_bank": "BBVA Bancomer",
            "sender_clabe": "012610001967722183",
            "receive_clabe": "646180115400467548",
            "numeric_reference": "80416",
            "concepto": "Para el 🐖",
            "clave_rastreo": "BNET01001604080002076841",
            "beneficiary_name": "ALFRED NORTH WHITEHEAD"
        }
    }]
}`)
	var fundingsResponse FundingsResponse
	err := json.Unmarshal(payload, &fundingsResponse)
	assert.NoError(t, err)

	assert.True(t, fundingsResponse.Success)
	assert.NotNil(t, fundingsResponse.Payload)
	assert.NotEmpty(t, fundingsResponse.Payload[0].FID)
}
