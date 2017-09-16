package main

import (
	"log"
	"os"

	"github.com/mazingstudio/bitso-go/bitso"
)

func main() {
	client := bitso.NewClient(nil)

	client.SetKey(os.Getenv("API_KEY"))
	client.SetSecret(os.Getenv("API_SECRET"))

	fundings, err := client.Fundings(nil)
	if err != nil {
		log.Fatalf("ERR: %#v", err)
	}

	log.Printf("OK: %#v", fundings)
}
