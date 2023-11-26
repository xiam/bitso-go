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

[1]: https://docs.bitso.com/bitso-api/docs
