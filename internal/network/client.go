package network

import (
	"errors"
	"math/rand"
	"net/url"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	fhttpcookiejar "github.com/bogdanfinn/fhttp/cookiejar"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

var ErrRequestFailed = errors.New("request failed")

type Client struct {
	http       tls_client.HttpClient
	rotator    *Rotator
	userAgents []string
	rand        *rand.Rand
}

func NewClient(rotator *Rotator) (*Client, error) {
	jar, _ := fhttpcookiejar.New(nil)

	client, err := tls_client.NewHttpClient(
		tls_client.NewNoopLogger(),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithCookieJar(jar),
	)
	if err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Client{
		http:       client,
		rotator:    rotator,
		userAgents: append([]string{}, userAgents...),
		rand:        rng,
	}, nil
}

func (c *Client) Do(req *fhttp.Request) (*fhttp.Response, error) {
	proxy, _ := c.rotateProxy()
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.randomUA())
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if proxy != nil {
		c.rotator.Report(proxy, resp.StatusCode)
	}
	return resp, nil
}

func (c *Client) rotateProxy() (*url.URL, error) {
	if c.rotator == nil {
		return nil, nil
	}
	proxy, err := c.rotator.Next()
	if err != nil {
		return nil, err
	}

	if proxy != nil {
		_ = c.http.SetProxy(proxy.String())
	}
	return proxy, nil
}

func (c *Client) randomUA() string {
	if len(c.userAgents) == 0 {
		return ""
	}
	return c.userAgents[c.rand.Intn(len(c.userAgents))]
}
