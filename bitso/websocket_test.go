package bitso

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketReply_UnmarshalJSON(t *testing.T) {
	t.Run("subscribe response", func(t *testing.T) {
		jsonData := `{
			"action": "subscribe",
			"response": "ok",
			"time": 1705312200000,
			"type": "trades"
		}`

		var reply WebSocketReply
		err := json.Unmarshal([]byte(jsonData), &reply)

		require.NoError(t, err)
		assert.Equal(t, "subscribe", reply.Action)
		assert.Equal(t, "ok", reply.Response)
		assert.Equal(t, uint64(1705312200000), reply.Time)
		assert.Equal(t, "trades", reply.Type)
		assert.Nil(t, reply.Payload)
	})

	t.Run("with payload", func(t *testing.T) {
		jsonData := `{
			"action": "",
			"response": "",
			"time": 1705312200000,
			"type": "trades",
			"payload": [{"key": "value"}]
		}`

		var reply WebSocketReply
		err := json.Unmarshal([]byte(jsonData), &reply)

		require.NoError(t, err)
		assert.Equal(t, "trades", reply.Type)
		assert.NotNil(t, reply.Payload)
	})
}

func TestWebSocketTrade_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"type": "trades",
		"book": "btc_mxn",
		"payload": [
			{
				"i": 12345,
				"a": "0.5",
				"r": "500000.00",
				"v": "250000.00",
				"t": "0",
				"x": 1705312200000,
				"mo": "maker-order-123",
				"to": "taker-order-456"
			},
			{
				"i": 12346,
				"a": "1.0",
				"r": "500100.00",
				"v": "500100.00",
				"t": "1",
				"x": 1705312201000,
				"mo": "maker-order-124",
				"to": "taker-order-457"
			}
		]
	}`

	var trade WebSocketTrade
	err := json.Unmarshal([]byte(jsonData), &trade)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", trade.Book.String())
	require.Len(t, trade.Payload, 2)

	// First trade
	assert.Equal(t, uint64(12345), trade.Payload[0].TID)
	assert.Equal(t, "0.5", string(trade.Payload[0].Amount))
	assert.Equal(t, "500000.00", string(trade.Payload[0].Price))
	assert.Equal(t, "250000.00", string(trade.Payload[0].Value))
	assert.Equal(t, "0", trade.Payload[0].MakerSide)
	assert.Equal(t, uint64(1705312200000), trade.Payload[0].CreationTimestamp)
	assert.Equal(t, "maker-order-123", trade.Payload[0].MakerOrderID)
	assert.Equal(t, "taker-order-456", trade.Payload[0].TakerOrderID)

	// Second trade
	assert.Equal(t, uint64(12346), trade.Payload[1].TID)
	assert.Equal(t, "1", trade.Payload[1].MakerSide)
}

func TestWebSocketDiffOrder_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"type": "diff-orders",
		"book": "eth_mxn",
		"sequence": 12345,
		"payload": [
			{
				"d": 1705312200000,
				"r": "35000.00",
				"s": "open",
				"t": 0,
				"a": "2.5",
				"v": "87500.00",
				"z": 1705312200500,
				"o": "order-123"
			},
			{
				"d": 1705312201000,
				"r": "35100.00",
				"s": "cancelled",
				"t": 1,
				"a": "1.0",
				"v": "35100.00",
				"z": 1705312201500,
				"o": "order-124"
			}
		]
	}`

	var diff WebSocketDiffOrder
	err := json.Unmarshal([]byte(jsonData), &diff)

	require.NoError(t, err)
	assert.Equal(t, "eth_mxn", diff.Book.String())
	require.Len(t, diff.Payload, 2)

	// First order
	assert.Equal(t, uint64(1705312200000), diff.Payload[0].Timestamp)
	assert.Equal(t, "35000.00", string(diff.Payload[0].Price))
	assert.Equal(t, "open", diff.Payload[0].Status)
	assert.Equal(t, 0, diff.Payload[0].Position)
	assert.Equal(t, "2.5", string(diff.Payload[0].Amount))
	assert.Equal(t, "87500.00", string(diff.Payload[0].Value))
	assert.Equal(t, uint64(1705312200500), diff.Payload[0].LastUpdateTimestamp)
	assert.Equal(t, "order-123", diff.Payload[0].OrderID)

	// Second order (cancelled)
	assert.Equal(t, "cancelled", diff.Payload[1].Status)
	assert.Equal(t, 1, diff.Payload[1].Position)
}

func TestWebSocketOrder_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"type": "orders",
		"book": "btc_mxn",
		"payload": {
			"bids": [
				{
					"a": "0.5",
					"o": "bid-order-1",
					"t": 0,
					"r": "499000.00",
					"s": "open",
					"d": 1705312200000,
					"v": "249500.00"
				},
				{
					"a": "1.0",
					"o": "bid-order-2",
					"t": 0,
					"r": "498000.00",
					"s": "open",
					"d": 1705312199000,
					"v": "498000.00"
				}
			],
			"asks": [
				{
					"a": "0.3",
					"o": "ask-order-1",
					"t": 1,
					"r": "500000.00",
					"s": "open",
					"d": 1705312200000,
					"v": "150000.00"
				}
			]
		}
	}`

	var order WebSocketOrder
	err := json.Unmarshal([]byte(jsonData), &order)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", order.Book.String())

	// Bids
	require.Len(t, order.Payload.Bids, 2)
	assert.Equal(t, "0.5", string(order.Payload.Bids[0].Amount))
	assert.Equal(t, "bid-order-1", order.Payload.Bids[0].OrderID)
	assert.Equal(t, 0, order.Payload.Bids[0].Position)
	assert.Equal(t, "499000.00", string(order.Payload.Bids[0].Price))
	assert.Equal(t, "open", order.Payload.Bids[0].Status)
	assert.Equal(t, uint64(1705312200000), order.Payload.Bids[0].Timestamp)
	assert.Equal(t, "249500.00", string(order.Payload.Bids[0].Value))

	// Asks
	require.Len(t, order.Payload.Asks, 1)
	assert.Equal(t, "0.3", string(order.Payload.Asks[0].Amount))
	assert.Equal(t, "ask-order-1", order.Payload.Asks[0].OrderID)
	assert.Equal(t, 1, order.Payload.Asks[0].Position)
}

func TestWebSocketMessage_MarshalJSON(t *testing.T) {
	t.Run("subscribe to trades", func(t *testing.T) {
		book := NewBook(BTC, MXN)
		msg := WebSocketMessage{
			Action: "subscribe",
			Book:   book,
			Type:   "trades",
		}

		data, err := json.Marshal(msg)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"action":"subscribe"`)
		assert.Contains(t, string(data), `"book":"btc_mxn"`)
		assert.Contains(t, string(data), `"type":"trades"`)
	})

	t.Run("subscribe to diff-orders", func(t *testing.T) {
		book := NewBook(ETH, MXN)
		msg := WebSocketMessage{
			Action: "subscribe",
			Book:   book,
			Type:   "diff-orders",
		}

		data, err := json.Marshal(msg)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"type":"diff-orders"`)
	})

	t.Run("subscribe to orders", func(t *testing.T) {
		book := NewBook(SOL, USD)
		msg := WebSocketMessage{
			Action: "subscribe",
			Book:   book,
			Type:   "orders",
		}

		data, err := json.Marshal(msg)

		require.NoError(t, err)
		assert.Contains(t, string(data), `"type":"orders"`)
	})
}

func TestWebSocketMessage_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"action": "subscribe",
		"book": "btc_mxn",
		"type": "trades"
	}`

	var msg WebSocketMessage
	err := json.Unmarshal([]byte(jsonData), &msg)

	require.NoError(t, err)
	assert.Equal(t, "subscribe", msg.Action)
	assert.Equal(t, "btc_mxn", msg.Book.String())
	assert.Equal(t, "trades", msg.Type)
}

func TestWebSocketTrade_EmptyPayload(t *testing.T) {
	jsonData := `{
		"type": "trades",
		"book": "btc_mxn",
		"payload": []
	}`

	var trade WebSocketTrade
	err := json.Unmarshal([]byte(jsonData), &trade)

	require.NoError(t, err)
	assert.Equal(t, "btc_mxn", trade.Book.String())
	assert.Empty(t, trade.Payload)
}

func TestWebSocketOrder_EmptyBidsAsks(t *testing.T) {
	jsonData := `{
		"type": "orders",
		"book": "eth_mxn",
		"payload": {
			"bids": [],
			"asks": []
		}
	}`

	var order WebSocketOrder
	err := json.Unmarshal([]byte(jsonData), &order)

	require.NoError(t, err)
	assert.Empty(t, order.Payload.Bids)
	assert.Empty(t, order.Payload.Asks)
}

func TestWebSocketConn_Receive(t *testing.T) {
	// Test that Receive returns a channel
	// Note: We can't easily test the full WebSocket connection without a mock server
	// This test verifies the basic structure

	ws := &WebSocketConn{
		inbox: make(chan interface{}, 8),
	}

	ch := ws.Receive()
	require.NotNil(t, ch)

	// Verify we can send to the channel
	go func() {
		ws.inbox <- WebSocketReply{Type: "test"}
	}()

	select {
	case msg := <-ch:
		reply, ok := msg.(WebSocketReply)
		require.True(t, ok)
		assert.Equal(t, "test", reply.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestWebSocketConn_Close_Nil(t *testing.T) {
	ws := &WebSocketConn{
		conn: nil,
	}

	err := ws.Close()
	assert.NoError(t, err)
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Test the different message type strings that can be received
	messageTypes := []string{
		"trades",
		"diff-orders",
		"orders",
		"ka", // keep-alive
	}

	for _, msgType := range messageTypes {
		t.Run(msgType, func(t *testing.T) {
			jsonData := `{"type": "` + msgType + `", "action": "", "response": "", "time": 0}`
			var reply WebSocketReply
			err := json.Unmarshal([]byte(jsonData), &reply)

			require.NoError(t, err)
			assert.Equal(t, msgType, reply.Type)
		})
	}
}

func TestWebSocketDiffOrder_StatusValues(t *testing.T) {
	// Test different status values in diff-orders
	statuses := []string{"open", "cancelled", "completed"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			jsonData := `{
				"type": "diff-orders",
				"book": "btc_mxn",
				"payload": [
					{
						"d": 1705312200000,
						"r": "500000.00",
						"s": "` + status + `",
						"t": 0,
						"a": "1.0",
						"v": "500000.00",
						"z": 1705312200000,
						"o": "order-1"
					}
				]
			}`

			var diff WebSocketDiffOrder
			err := json.Unmarshal([]byte(jsonData), &diff)

			require.NoError(t, err)
			assert.Equal(t, status, diff.Payload[0].Status)
		})
	}
}

func TestWebSocketTrade_MakerSideValues(t *testing.T) {
	// Test maker side values (0=buy, 1=sell per API docs)
	makerSides := []string{"0", "1"}

	for _, side := range makerSides {
		t.Run("side_"+side, func(t *testing.T) {
			jsonData := `{
				"type": "trades",
				"book": "btc_mxn",
				"payload": [
					{
						"i": 12345,
						"a": "1.0",
						"r": "500000.00",
						"v": "500000.00",
						"t": "` + side + `",
						"x": 1705312200000,
						"mo": "maker-1",
						"to": "taker-1"
					}
				]
			}`

			var trade WebSocketTrade
			err := json.Unmarshal([]byte(jsonData), &trade)

			require.NoError(t, err)
			assert.Equal(t, side, trade.Payload[0].MakerSide)
		})
	}
}
