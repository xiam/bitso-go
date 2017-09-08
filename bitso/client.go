package bitso

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	apiPrefix = "https://api.bitso.com/"
)

var DefaultClient = Client{
	client:  http.DefaultClient,
	Version: "v3",
}

type Client struct {
	client *http.Client

	Key     string
	Secret  string
	Version string
}

func (c *Client) EndpointURL(endpoint string) (*url.URL, error) {
	return url.Parse(apiPrefix + c.Version + "/" + endpoint)
}

func (c *Client) Get(uri string) (*http.Response, error) {
	return c.client.Get(uri)
}

func (c *Client) getResponse(endpoint string, params url.Values, dest interface{}) error {
	u, err := c.EndpointURL(endpoint)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	res, err := c.client.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Printf("buf: %v", string(buf))

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
