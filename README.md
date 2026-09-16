# bitso-go

`bitso-go` is a Go wrapper around the [Bitso API][1] for the Bitso
Cryptocurrency Exchange.

```
go get -u github.com/xiam/bitso-go/bitso
```

## Examples

The example below prints fundings in your account:

```go
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
```

See a few more examples: https://github.com/xiam/bitso-go/tree/master/_examples

## Supported API surface

The client covers the core Trading API endpoints documented by Bitso:

- Public market data: books, tickers, trades, order books, and websocket
  channels.
- Private account data: balances, account status, fees, ledger entries,
  fundings, withdrawals, user trades, order trades, open orders, and order
  lookups.
- Order management: order placement with current optional fields, cancellation,
  and v4 order modification by order id, query `oid`, or `origin_id`.
- Conversions: v4 quote request, quote execution, and conversion status lookup.
- RFQ: pairs, quote request, quote lookup, quote conversion, and conversion
  lookup.
- Margin trading: alpha margin account creation and summary, available
  currencies, movements, movement lookup, movement processing, and loans.

Signed private requests use Bitso Nonce v2 by default. The legacy v3 and current
v4 REST endpoints use the configured API base URL. Product-root APIs such as RFQ
and margin trading default to `https://api.bitso.com` and derive the same host
and parent path when `SetAPIBaseURL` is customized for tests or sandbox use.

### HTTP client configuration

`bitso.NewClient()` uses an internal `http.Client` with
`bitso.DefaultHTTPClientTimeout`. Inject a custom client when you need a
different timeout, proxy, transport, or test double:

```go
client := bitso.NewClient()
client.SetHTTPClient(&http.Client{
    Timeout: 10 * time.Second,
})
```

Passing `nil` to `SetHTTPClient` restores the default timeout-backed client.

### Monetary values

Bitso amounts are decimal strings. Prefer `bitso.NewMonetary` for order
placement so invalid input is rejected and precision-sensitive values are not
rounded through `float64`:

```go
major, err := bitso.NewMonetary("0.10000000")
if err != nil {
    log.Fatal("invalid major amount: ", err)
}
price, err := bitso.NewMonetary("500000.00")
if err != nil {
    log.Fatal("invalid price: ", err)
}

oid, err := client.PlaceOrder(&bitso.OrderPlacement{
    Book:  *bitso.NewBook(bitso.BTC, bitso.MXN),
    Side:  bitso.OrderSideBuy,
    Type:  bitso.OrderTypeLimit,
    Major: major,
    Price: price,
})
if err != nil {
    log.Fatal("can not place order: ", err)
}
log.Println("new order:", oid)
```

`bitso.ToMonetary(float64)` and `Monetary.Float64()` remain available for
backward compatibility, but they are deprecated for monetary calculations. Use
`NewMonetary` with string input, `NewMonetaryFromDecimal` with
`decimal.Decimal`, or `Monetary.Decimal()` for exact decimal handling. If a
`float64` integration is unavoidable, use `Monetary.Float64E()` and handle the
returned error.

## Development

This module supports Go 1.25.12 and newer stable Go releases. The `go.mod`
directive stays at Go 1.25.0 for language compatibility, while CI runs on Go
1.25.12 to keep the minimum supported line covered with current security
patches.

The lint baseline uses golangci-lint v2 with the version pinned in
the CI configuration. Run the same checks locally with:

```sh
go test ./...
go test -race ./...
golangci-lint run ./...
GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
```

Live Bitso sandbox checks are skipped by default so normal test runs stay
offline and deterministic. Run them explicitly with:

```sh
BITSO_INTEGRATION=1 go test -count=1 -v ./testing -run TestPublicAPIIntegration
```

Private auth checks are also opt-in and read-only. Use restricted credentials;
the test only calls the signed balance endpoint and does not place orders, move
funds, create margin accounts, or execute conversions:

```sh
BITSO_AUTH_INTEGRATION=1 BITSO_API_KEY=... BITSO_API_SECRET=... go test -count=1 -v ./testing -run TestPrivateAuthIntegration
```

CI keeps the standard test jobs offline. Start a pipeline with
`BITSO_CI_INTEGRATION=1` to add the opt-in live sandbox job.

The CI vulnerability scan is blocking. If govulncheck reports a reachable
vulnerability, update the affected dependency or document a deliberate
exception before merging.

## License

MIT

> Copyright 2017-today, José Nieto.
>
> Permission is hereby granted, free of charge, to any person obtaining a copy of
> this software and associated documentation files (the "Software"), to deal in
> the Software without restriction, including without limitation the rights to
> use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
> of the Software, and to permit persons to whom the Software is furnished to do
> so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in all
> copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
> IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
> FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
> AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
> LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
> OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
> SOFTWARE.

[1]: https://docs.bitso.com/bitso-api/docs/api-overview
