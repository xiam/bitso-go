package main

import (
	"log"

	"github.com/xiam/bitso-go/bitso"
)

func main() {
	ws, err := bitso.NewWebSocketConn()
	if err != nil {
		log.Fatal("bitso.NewWebSocketConn: ", err)
	}

	err = ws.Subscribe(bitso.NewBook(bitso.ETH, bitso.MXN), "orders")
	if err != nil {
		log.Fatal("ws.Subscribe: ", err)
	}

	for {
		m := <-ws.Receive()
		log.Printf("message: %#v\n\n", m)
	}
}
