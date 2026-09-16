package testing

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xiam/bitso-go/bitso"
)

const (
	bitsoIntegrationEnv     = "BITSO_INTEGRATION"
	bitsoAuthIntegrationEnv = "BITSO_AUTH_INTEGRATION"
	bitsoAPIKeyEnv          = "BITSO_API_KEY"
	bitsoAPISecretEnv       = "BITSO_API_SECRET"
)

func bitsoIntegrationEnabled() bool {
	return os.Getenv(bitsoIntegrationEnv) == "1"
}

func bitsoAuthIntegrationEnabled() bool {
	return os.Getenv(bitsoAuthIntegrationEnv) == "1"
}

func requireBitsoIntegration(t *testing.T) {
	t.Helper()

	if !bitsoIntegrationEnabled() {
		t.Skipf("set %s=1 to run live Bitso sandbox integration tests", bitsoIntegrationEnv)
	}
}

func requireBitsoAuthIntegration(t *testing.T) (string, string) {
	t.Helper()

	if !bitsoAuthIntegrationEnabled() {
		t.Skipf("set %s=1 to run live Bitso private auth integration tests", bitsoAuthIntegrationEnv)
	}

	key, secret := os.Getenv(bitsoAPIKeyEnv), os.Getenv(bitsoAPISecretEnv)
	if key == "" || secret == "" {
		t.Skipf("set %s and %s to run live Bitso private auth integration tests", bitsoAPIKeyEnv, bitsoAPISecretEnv)
	}
	return key, secret
}

func TestBitsoIntegrationGate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantRun bool
	}{
		{name: "empty"},
		{name: "disabled", value: "0"},
		{name: "enabled", value: "1", wantRun: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(bitsoIntegrationEnv, tc.value)
			assert.Equal(t, tc.wantRun, bitsoIntegrationEnabled())
		})
	}
}

func TestBitsoAuthIntegrationGate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantRun bool
	}{
		{name: "empty"},
		{name: "disabled", value: "0"},
		{name: "enabled", value: "1", wantRun: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(bitsoAuthIntegrationEnv, tc.value)
			assert.Equal(t, tc.wantRun, bitsoAuthIntegrationEnabled())
		})
	}
}

func TestPublicAPIIntegration(t *testing.T) {
	requireBitsoIntegration(t)

	c := bitso.NewClient()
	c.SetAPIBaseURL("https://sandbox.bitso.com/api")

	t.Run("AvailableBooks", func(t *testing.T) {
		books, err := c.AvailableBooks()
		require.NoError(t, err)
		assert.NotNil(t, books)
	})

	t.Run("Ticker", func(t *testing.T) {
		tickers, err := c.Tickers()
		require.NoError(t, err)
		assert.NotNil(t, tickers)

		ticker, err := c.Ticker(bitso.NewBook(bitso.ETH, bitso.MXN))
		require.NoError(t, err)
		assert.NotNil(t, ticker)
	})

	t.Run("OrderBook", func(t *testing.T) {
		orderBook, err := c.OrderBook(url.Values{
			"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
		})
		require.NoError(t, err)
		assert.NotNil(t, orderBook)
	})

	t.Run("Trades", func(t *testing.T) {
		_, err := c.Trades(nil)
		assert.Error(t, err)

		ticker, err := c.Trades(url.Values{
			"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
		})
		require.NoError(t, err)
		assert.NotNil(t, ticker)
	})
}

func TestPrivateAuthIntegration(t *testing.T) {
	key, secret := requireBitsoAuthIntegration(t)

	c := bitso.NewClient()
	c.SetAuth(key, secret)

	balances, err := c.Balances(nil)

	require.NoError(t, err)
	assert.NotNil(t, balances)
}
