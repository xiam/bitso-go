package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/xiam/bitso-go/bitso"
)

var client = bitso.NewClient(nil)

func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 4, 4, 3, ' ', 0)
}

func init() {
	key, secret := os.Getenv("BITSO_API_KEY"), os.Getenv("BITSO_API_SECRET")

	client.SetAPIKey(key)
	client.SetAPISecret(secret)
}

func printBalance() {
	balances, err := client.Balances(nil)
	if err != nil {
		log.Fatal("client.Balances: ", err)
	}

	w := newTabWriter()
	fmt.Fprintf(w, "CURRENCY\tTOTAL\tLOCKED\tAVAILABLE\n")
	for _, b := range balances {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			strings.ToUpper(b.Currency.String()),
			b.Total,
			b.Locked,
			b.Available,
		)
	}

	w.Flush()
}

func main() {
	printBalance()
}
