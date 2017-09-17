# bitso-go

`bitso-go` is a Go wrapper around the [Bitso API](https://bitso.com/api_info)
for the Bitso Bitcoin Exchange.

```
go get -u github.com/mazingstudio/bitso-go/bitso
```

`bitso-go` supports the
[public](https://bitso.com/api_info?l=es#public-rest-api) and
[private](https://bitso.com/api_info?l=es#private-rest-api) REST APIs and also
provides a websocket interface for the
[Websocket](https://bitso.com/api_info?l=es#websocket-api) API.


## Get API key and secret

You can get you API keys and secrets from here: https://bitso.com/api_setup

## Examples

The example below prints fundings in your account:

```go
package main

import (
	"log"
	"os"

	"github.com/mazingstudio/bitso-go/bitso"
)

func main() {
	client := bitso.NewClient(nil)

	client.SetAPIKey(os.Getenv("API_KEY"))
	client.SetAPISecret(os.Getenv("API_SECRET"))

	fundings, err := client.Fundings(nil)
	if err != nil {
		log.Fatal("client.Fundings: ", err)
	}

	for _, funding := range fundings {
		log.Print(funding)
	}
}
```

You can compile and run it with:

```
API_KEY=foo API_SECRET=bar go run main.go
```

See also
[print-balance](https://github.com/mazingstudio/bitso-go/blob/master/_examples/print-balance/main.go)
and
[websocket](https://github.com/mazingstudio/bitso-go/blob/master/_examples/websocket/main.go)
examples.

## License

MIT

> Copyright 2017, Mazing Studio SA de CV
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
