package bitso

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const wssURL = `wss://ws.bitso.com`

// ErrWebSocketClosed is returned when an operation needs an open WebSocket.
var ErrWebSocketClosed = errors.New("bitso: websocket connection closed")

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
	conn *websocket.Conn

	inbox   chan interface{}
	closing chan struct{}
	done    chan struct{}

	closeOnce sync.Once
	doneOnce  sync.Once
	mu        sync.RWMutex
	err       error
}

type webSocketDialer interface {
	Dial(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

type webSocketConfig struct {
	endpoint string
	dialer   webSocketDialer
}

// WebSocketOption configures a WebSocket connection.
type WebSocketOption func(*webSocketConfig)

// WithWebSocketEndpoint configures the WebSocket endpoint.
func WithWebSocketEndpoint(endpoint string) WebSocketOption {
	return func(cfg *webSocketConfig) {
		cfg.endpoint = endpoint
	}
}

// WithWebSocketDialer configures the dialer used to establish the WebSocket connection.
func WithWebSocketDialer(dialer webSocketDialer) WebSocketOption {
	return func(cfg *webSocketConfig) {
		if dialer != nil {
			cfg.dialer = dialer
		}
	}
}

// Receive returns a channel where received messages are sent.
//
// The channel is closed when the receive loop exits.
func (ws *WebSocketConn) Receive() chan interface{} {
	return ws.inbox
}

// Done returns a channel that is closed when the receive loop exits.
func (ws *WebSocketConn) Done() <-chan struct{} {
	if ws.done == nil {
		done := make(chan struct{})
		close(done)
		return done
	}
	return ws.done
}

// Err returns the error that stopped the receive loop.
//
// Err returns nil when the connection was closed through Close.
func (ws *WebSocketConn) Err() error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	return ws.err
}

// NewWebSocketConn creates a websocket handler and establishes a connection with
// Bitso's websocket servers.
func NewWebSocketConn(options ...WebSocketOption) (*WebSocketConn, error) {
	cfg := webSocketConfig{
		endpoint: wssURL,
		dialer:   websocket.DefaultDialer,
	}
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	ws := &WebSocketConn{
		inbox:   make(chan interface{}, 8),
		closing: make(chan struct{}),
		done:    make(chan struct{}),
	}

	var err error
	ws.conn, _, err = cfg.dialer.Dial(cfg.endpoint, nil)
	if err != nil {
		return nil, err
	}

	go ws.readLoop()

	return ws, nil
}

func (ws *WebSocketConn) readLoop() {
	var terminalErr error
	defer func() {
		if terminalErr != nil {
			_ = ws.Close()
		}
		ws.finish(terminalErr)
	}()

	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			if ws.isClosing() {
				return
			}
			terminalErr = fmt.Errorf("websocket read: %w", err)
			return
		}

		msg, err := decodeWebSocketMessage(data)
		if err != nil {
			terminalErr = err
			return
		}
		if msg == nil {
			continue
		}
		if !ws.dispatch(msg) {
			return
		}
	}
}

func decodeWebSocketMessage(data []byte) (interface{}, error) {
	var reply WebSocketReply
	if err := json.Unmarshal(data, &reply); err != nil {
		return nil, fmt.Errorf("websocket unmarshal message: %w", err)
	}

	switch reply.Type {
	case "diff-orders":
		if reply.Payload != nil {
			var diff WebSocketDiffOrder
			if err := json.Unmarshal(data, &diff); err != nil {
				return nil, fmt.Errorf("websocket unmarshal diff order: %w", err)
			}
			return diff, nil
		}
	case "ka":
		return nil, nil
	case "orders":
		if reply.Payload != nil {
			var order WebSocketOrder
			if err := json.Unmarshal(data, &order); err != nil {
				return nil, fmt.Errorf("websocket unmarshal order: %w", err)
			}
			return order, nil
		}
	case "trades":
		if reply.Payload != nil {
			var trade WebSocketTrade
			if err := json.Unmarshal(data, &trade); err != nil {
				return nil, fmt.Errorf("websocket unmarshal trade: %w", err)
			}
			return trade, nil
		}
	}

	return reply, nil
}

func (ws *WebSocketConn) dispatch(msg interface{}) bool {
	select {
	case ws.inbox <- msg:
		return true
	case <-ws.closing:
		return false
	}
}

func (ws *WebSocketConn) finish(err error) {
	ws.mu.Lock()
	ws.err = err
	ws.mu.Unlock()

	ws.doneOnce.Do(func() {
		close(ws.inbox)
		close(ws.done)
	})
}

func (ws *WebSocketConn) isClosing() bool {
	select {
	case <-ws.closing:
		return true
	default:
		return false
	}
}

// Close closes the active connection with Bitso's websocket servers.
func (ws *WebSocketConn) Close() error {
	var err error
	ws.closeOnce.Do(func() {
		if ws.closing != nil {
			close(ws.closing)
		}
		if ws.conn != nil {
			err = ws.conn.Close()
		}
	})
	return err
}

// Subscribe subscribes to a messages channel.
func (ws *WebSocketConn) Subscribe(book *Book, channelName string) error {
	if ws.conn == nil || ws.isClosing() {
		return ErrWebSocketClosed
	}

	m := WebSocketMessage{
		Action: "subscribe",
		Book:   book,
		Type:   channelName,
	}
	if err := ws.conn.WriteJSON(m); err != nil {
		if ws.isClosing() {
			return ErrWebSocketClosed
		}
		return err
	}
	return nil
}
