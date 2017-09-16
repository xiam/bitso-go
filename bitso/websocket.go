package bitso

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

const wssURL = `wss://ws.bitso.com`

// WebsocketReply represents a generic reply from a channel.
type WebsocketReply struct {
	Action   string      `json:"action"`
	Response string      `json:"response"`
	Time     uint64      `json:"time"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload,omitempty"`
}

// WebsocketTrade represents a message from the "trades" channel.
type WebsocketTrade struct {
	Book    Book
	Payload []struct {
		TID          uint64   `json:"i"`
		Amount       Monetary `json:"a"`
		Price        Monetary `json:"r"`
		Value        Monetary `json:"v"`
		MakerOrderID string   `json:"mo"`
		TakerOrderID string   `json:"to"`
	}
}

// WebsocketDiffOrder represents a message from the "diff-orders" channel.
type WebsocketDiffOrder struct {
	Book    Book
	Payload []struct {
		Timestamp uint64   `json:"d"`
		Price     Monetary `json:"r"`
		Position  int      `json:"t"`
		Amount    Monetary `json:"a"`
		Value     Monetary `json:"v"`
		OrderID   string   `json:"o"`
	}
}

// WebsocketOrder represents a message from the "diff-orders" channel.
type WebsocketOrder struct {
	Book    Book
	Payload struct {
		Bids []struct {
			Price     float64 `json:"r"`
			Amount    float64 `json:"a"`
			Position  int     `json:"t"`
			Value     float64 `json:"v"`
			Timestamp uint64  `json:"d"`
		} `json:"bids"`
		Asks []struct {
			Price     float64 `json:"r"`
			Amount    float64 `json:"a"`
			Position  int     `json:"t"`
			Value     float64 `json:"v"`
			Timestamp uint64  `json:"d"`
		} `json:"asks"`
	} `json:"payload"`
}

// WebsocketMessage represents a message that can be sent to channel.
type WebsocketMessage struct {
	Action string `json:"action"`
	Book   *Book  `json:"book"`
	Type   string `json:"type"`
}

// A Websocket establishes a connection with Bitso's websocket service to send
// and receive messages over the ws protocol.
type Websocket struct {
	endpoint string
	conn     *websocket.Conn

	inbox chan interface{}
}

// Receive returns a channel where received messages are sent.
func (ws *Websocket) Receive() chan interface{} {
	return ws.inbox
}

// NewWebsocket creates a websocket handler and establishes a connection with
// Bitso's websocket servers.
func NewWebsocket() (*Websocket, error) {
	ws := &Websocket{
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

			var reply WebsocketReply
			if err := json.Unmarshal(data, &reply); err != nil {
				log.Printf("failed to unmarshal message: %v", err)
				return
			}

			switch reply.Type {
			case "diff-orders":
				if reply.Payload != nil {
					var diff WebsocketDiffOrder
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
					var order WebsocketOrder
					if err := json.Unmarshal(data, &order); err != nil {
						log.Printf("failed to unmarshal order: %v", err)
						return
					}
					ws.inbox <- order
					continue
				}
			case "trades":
				if reply.Payload != nil {
					var trade WebsocketTrade
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
func (ws *Websocket) Close() error {
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

// Subscribe subscribes to a messages channel.
func (ws *Websocket) Subscribe(book *Book, channelName string) error {
	m := WebsocketMessage{
		Action: "subscribe",
		Book:   book,
		Type:   channelName,
	}
	return ws.conn.WriteJSON(m)
}
