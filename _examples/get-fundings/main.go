package main

import (
	"log"
	"os"

	"github.com/xiam/bitso-go/bitso"
)

func main() {
	client := bitso.NewClient(nil)

	client.SetAPIKey(os.Getenv("BITSO_API_KEY"))
	client.SetAPISecret(os.Getenv("BITSO_API_SECRET"))

	fundings, err := client.Fundings(nil)
	if err != nil {
		log.Fatal("client.Fundings: ", err)
	}

	for _, funding := range fundings {
		log.Print(funding)
	}
}
