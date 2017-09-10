package bitso

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	debug = false
)

const (
	apiPrefix      = "https://api.bitso.com/"
	defaultTickets = 1
)

var (
	DefaultClient    = NewClient(http.DefaultClient)
	defaultBurstRate = time.Millisecond * 1500
)

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{
		client:  httpClient,
		tickets: make(chan struct{}, defaultTickets),

		version:   "v3",
		BurstRate: defaultBurstRate,
	}
	for i := 0; i < defaultTickets; i++ {
		c.tickets <- struct{}{}
	}
	return c
}

type Client struct {
	client *http.Client

	key     string
	secret  string
	version string

	tickets chan struct{}

	BurstRate time.Duration
}

func (c *Client) SetKey(key string) {
	c.key = key
}

func (c *Client) SetSecret(secret string) {
	c.secret = secret
}

func (c *Client) EndpointURL(endpoint string) (*url.URL, error) {
	return url.Parse(apiPrefix + c.version + "/" + endpoint)
}

func (c *Client) lock() {
	tk := time.NewTicker(c.BurstRate)
	<-tk.C
}

func (c *Client) Get(uri string) (*http.Response, error) {
	c.debugf("Waiting...")

	<-c.tickets
	c.debugf("Get")
	defer func() {
		c.debugf("Done!")
	}()

	ticker := time.NewTicker(c.BurstRate)
	go func() {
		<-ticker.C
		c.debugf("Next!")
		c.tickets <- struct{}{}
	}()

	return c.client.Get(uri)
}

func (c *Client) debugf(f string, a ...interface{}) {
	if !debug {
		return
	}
	log.Printf(f, a...)
}

func (c *Client) newRequest(method string, uri string, body io.Reader) (*http.Request, error) {
	nonce := time.Now().UnixNano()

	var buf []byte

	if body != nil {
		var err error
		buf, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	if c.key == "" && c.secret == "" {
		// Return unsigned request
		return req, nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("%d%s%s%s", nonce, method, u.Path, string(buf))

	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(message))
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	authHeader := fmt.Sprintf("Bitso %s:%d:%s", c.key, nonce, signature)
	req.Header.Set("Authorization", authHeader)

	return req, err
}

func (c *Client) getResponse(endpoint string, params url.Values, dest interface{}) error {
	u, err := c.EndpointURL(endpoint)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	req, err := c.newRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	//c.debugf("buf: %v", string(buf))

	if err := json.Unmarshal(buf, dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) Trades(params url.Values) (*TradesResponse, error) {
	var res TradesResponse
	if err := c.getResponse("/trades", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Fundings(params url.Values) (*FundingsResponse, error) {
	var res FundingsResponse
	if err := c.getResponse("/fundings", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
