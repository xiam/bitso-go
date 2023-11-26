package main

import (
	"log"
	"os"

	"github.com/xiam/bitso-go/bitso"
)

func main() {
	key, secret := os.Getenv("BITSO_API_KEY"), os.Getenv("BITSO_API_SECRET")
	if key == "" || secret == "" {
		log.Fatal("Please set BITSO_API_KEY and BITSO_API_SECRET")
	}

	client := bitso.NewClient()
	client.SetLogLevel(bitso.LogLevelDebug)

	client.SetAuth(key, secret)

	fundings, err := client.Fundings(nil)
	if err != nil {
		log.Fatal("can not get fundings: ", err)
	}

	for _, funding := range fundings {
		log.Printf("%#v", funding)
	}
}
