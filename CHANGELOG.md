# Changelog

Public release notes for `github.com/xiam/bitso-go`. Each `## vX.Y.Z - YYYY-MM-DD`
section is the release's public note; an annotated release tag carries it as its
annotation body. Entries follow the public changelog rubric.

## Unreleased

## v0.10.0 - 2026-09-16

### Added

- Add `Client.ModifyOrder()`, `Client.ModifyOrderByOriginID()`, and
  `Client.ModifyOrderByQueryOID()` with `OrderModification` to change an open
  order through the v4 API.
- Add `OrderTimeInForce` to set the time-in-force of a placed order, with the
  values `OrderTimeInForceGoodTillCancelled`, `OrderTimeInForceFillOrKill`,
  `OrderTimeInForceImmediateOrCancel`, and `OrderTimeInForcePostOnly`.
- Add margin trading: `Client.CreateMarginAccount()`,
  `Client.MarginAccountSummary()`, `Client.MarginCurrencies()`,
  `Client.MarginLoans()`, `Client.MarginMovements()`, `Client.MarginMovement()`,
  and `Client.ProcessMarginMovement()`, with the `Margin*` request and response
  types and `MarginErrors`.
- Add RFQ (request for quote): `Client.RFQPairs()`, `Client.RequestRFQQuote()`,
  `Client.RFQQuote()`, `Client.ConvertRFQQuote()`, and `Client.RFQConversion()`,
  with the `RFQ*` types and `RFQErrors`.
- Add currency conversion: `Client.RequestCurrencyConversionQuote()`,
  `Client.ExecuteCurrencyConversion()`, and
  `Client.CurrencyConversionStatus()`, with the `CurrencyConversion*` types.
- Add `Client.AccountStatus()` and the `AccountStatus` type.
- Add `Client.SetHTTPClient()` and `Client.HTTPClient()` to supply a custom
  `*http.Client` for all requests.
- Add `NewMonetary()` and `NewMonetaryFromDecimal()` to build a `Monetary`
  value without going through a float.
- Add `Monetary.Float64E()`. It returns the float64 value and any parse error.
- Add `NewWebSocketConn()` options of type `WebSocketOption`.
  `WithWebSocketEndpoint()` selects the WebSocket endpoint.
  `WithWebSocketDialer()` injects the dialer.
- Add `WebSocketConn.Done()`. The channel it returns closes when the
  connection ends.
- Add `WebSocketConn.Err()`. It reports why the connection ended.
- Add `ErrWebSocketClosed`, returned by operations that need an open
  connection.
- Add `DefaultHTTPClientTimeout` (30 seconds), the request timeout
  `NewClient()` applies.
- Add `MarginOrderType` with `MarginOrderTypeCrossMargin`, and the
  `OrderPlacement` fields `MarginOrderType`, `OriginID`, `SlippageTolerance`,
  `Stop`, and `TimeInForce`.
- Add the `OriginID`, `MarginOrderType`, and `TimeInForce` fields to
  `UserOrder`.
- Add the `MajorCurrency`, `MinorCurrency`, `MakerSide`, and `OriginID` fields
  to `UserTrade` and `UserOrderTrade`, and `MarginOrderType` to `UserTrade`.
- Add the `Fee` fields `MakerFeeDecimal`, `MakerFeePercent`, `TakerFeeDecimal`,
  `TakerFeePercent`, `NextMakerFeePercent`, `NextTakerFeePercent`, `NextFee`,
  `NextTakerFee`, `CurrentVolume`, `NextVolume`, and `VolumeCurrency`.
- Add `CustomerFees.DepositFees` and the `DepositFee` type.

### Changed

- Sign private requests with Bitso nonce v2.
- Change the JSON key of `UserOrderTrade.FeesCurrency` to `fees_currency`.
  Decoding still accepts the legacy `currency` key.
- Deprecate `Monetary.Float64()`; use `Monetary.Decimal()` for exact decimal
  math or `Monetary.Float64E()` when a float64 is required.
- Deprecate `ToMonetary()`; use `NewMonetary()` or `NewMonetaryFromDecimal()`
  to avoid float64 rounding and formatting loss.
- Deprecate `Fee.FeeDecimal` and `Fee.FeePercent`; use `Fee.MakerFeeDecimal`
  and `Fee.MakerFeePercent`.
- Deprecate `Fee.NextFee` and `Fee.NextTakerFee`; use `Fee.NextMakerFeePercent`
  and `Fee.NextTakerFeePercent`.
- Deprecate `Balance.PendingDeposit` and `Balance.PendingWithdrawal`. Current
  Bitso balance responses no longer document them.
- Raise the minimum Go version to 1.25.
- Document the supported API surface, HTTP client configuration, and monetary
  values in the README.

### Fixed

- Fix REST requests that never timed out. `NewClient()` now applies
  `DefaultHTTPClientTimeout`.
- Fix the data race between `Client.SetAuth()` and an in-flight request.
  `Client` now reads its configuration under a lock.
- Fix the WebSocket connection lifecycle. `Close()` ends the reader, the
  `Receive()` channel closes when the connection ends, and `Err()` reports the
  final error.
- Fix SQL scanning of `Book`, `Currency`, `Operation`, `OrderSide`, and
  `OrderStatus` to accept `[]byte` and `nil` database values, and to return an
  error for other types instead of panicking.
- Fix `Time` JSON decoding to accept `null` and RFC 3339 timestamps with a `Z`
  zone.
- Redact authentication material from trace logs.
