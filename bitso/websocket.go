package bitso

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

const wssURL = `wss://ws.bitso.com`

// WebSocketReply represents a generic reply from a channel.
type WebSocketReply struct {
	Action   string      `json:"action"`
	Response string      `json:"response"`
	Time     uint64      `json:"time"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload,omitempty"`
}

// WebSocketTrade represents a message from the "trades" channel.
type WebSocketTrade struct {
	Book    Book
	Payload []struct {
		TID               uint64   `json:"i"`
		Amount            Monetary `json:"a"`
		Price             Monetary `json:"r"`
		Value             Monetary `json:"v"`
		MakerSide         string   `json:"t"`
		CreationTimestamp uint64   `json:"x"`
		MakerOrderID      string   `json:"mo"`
		TakerOrderID      string   `json:"to"`
	}
}

// WebSocketDiffOrder represents a message from the "diff-orders" channel.
type WebSocketDiffOrder struct {
	Book    Book
	Payload []struct {
		Timestamp           uint64   `json:"d"`
		Price               Monetary `json:"r"`
		Status              string   `json:"s"`
		Position            int      `json:"t"`
		Amount              Monetary `json:"a"`
		Value               Monetary `json:"v"`
		LastUpdateTimestamp uint64   `json:"z"`
		OrderID             string   `json:"o"`
	}
}

// WebSocketOrder represents a message from the "diff-orders" channel.
type WebSocketOrder struct {
	Book    Book
	Payload struct {
		Bids []struct {
			Amount    Monetary `json:"a"`
			OrderID   string   `json:"o"`
			Position  int      `json:"t"`
			Price     Monetary `json:"r"`
			Status    string   `json:"s"`
			Timestamp uint64   `json:"d"`
			Value     Monetary `json:"v"`
		} `json:"bids"`
		Asks []struct {
			Amount    Monetary `json:"a"`
			OrderID   string   `json:"o"`
			Position  int      `json:"t"`
			Price     Monetary `json:"r"`
			Status    string   `json:"s"`
			Timestamp uint64   `json:"d"`
			Value     Monetary `json:"v"`
		} `json:"asks"`
	} `json:"payload"`
}

// WebSocketMessage represents a message that can be sent to channel.
type WebSocketMessage struct {
	Action string `json:"action"`
	Book   *Book  `json:"book"`
	Type   string `json:"type"`
}

// A WebSocketConn establishes a connection with Bitso's websocket service to
// send and receive messages over the ws protocol.
type WebSocketConn struct {
	endpoint string
	conn     *websocket.Conn

	inbox chan interface{}
}

// Receive returns a channel where received messages are sent.
func (ws *WebSocketConn) Receive() chan interface{} {
	return ws.inbox
}

// WebSocketConn creates a websocket handler and establishes a connection with
// Bitso's websocket servers.
func NewWebSocketConn() (*WebSocketConn, error) {
	ws := &WebSocketConn{
		endpoint: wssURL,
		inbox:    make(chan interface{}, 8),
	}

	var err error
	ws.conn, _, err = websocket.DefaultDialer.Dial(wssURL, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		defer ws.Close()
		for {
			_, data, err := ws.conn.ReadMessage()
			if err != nil {
				log.Printf("failed to read message: %v", err)
				return
			}

			var reply WebSocketReply
			if err := json.Unmarshal(data, &reply); err != nil {
				log.Printf("failed to unmarshal message: %v", err)
				return
			}

			switch reply.Type {
			case "diff-orders":
				if reply.Payload != nil {
					var diff WebSocketDiffOrder
					if err := json.Unmarshal(data, &diff); err != nil {
						log.Printf("failed to unmarshal diff order: %v", err)
						return
					}
					ws.inbox <- diff
					continue
				}
			case "ka":
				// keep alive
				continue
			case "orders":
				if reply.Payload != nil {
					var order WebSocketOrder
					if err := json.Unmarshal(data, &order); err != nil {
						log.Printf("failed to unmarshal order: %v", err)
						return
					}
					ws.inbox <- order
					continue
				}
			case "trades":
				if reply.Payload != nil {
					var trade WebSocketTrade
					if err := json.Unmarshal(data, &trade); err != nil {
						log.Printf("failed to unmarshal trade: %v", err)
						return
					}
					ws.inbox <- trade
					continue
				}
			}

			ws.inbox <- reply
		}
	}()

	return ws, nil
}

// Close closes the active connection with Bitso's websocket servers.
func (ws *WebSocketConn) Close() error {
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

// Subscribe subscribes to a messages channel.
func (ws *WebSocketConn) Subscribe(book *Book, channelName string) error {
	m := WebSocketMessage{
		Action: "subscribe",
		Book:   book,
		Type:   channelName,
	}
	return ws.conn.WriteJSON(m)
}
