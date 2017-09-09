package bitso

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	debug = true
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
	c := &Client{
		client:  httpClient,
		tickets: make(chan struct{}, defaultTickets),

		Version:   "v3",
		BurstRate: defaultBurstRate,
	}
	for i := 0; i < defaultTickets; i++ {
		c.tickets <- struct{}{}
	}
	return c
}

type Client struct {
	client *http.Client

	Key     string
	Secret  string
	Version string

	tickets chan struct{}

	BurstRate time.Duration
}

func (c *Client) EndpointURL(endpoint string) (*url.URL, error) {
	return url.Parse(apiPrefix + c.Version + "/" + endpoint)
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

func (c *Client) getResponse(endpoint string, params url.Values, dest interface{}) error {
	u, err := c.EndpointURL(endpoint)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	res, err := c.Get(u.String())
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
