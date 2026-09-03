// Package wi provides a Go client for the wi-api WhatsApp platform.
//
// # Quick start
//
//	client := wi.New("YOUR_API_KEY")
//	msg, err := client.Session("my-instance").SendText(ctx, wi.SendTextParams{
//	    To:   "5511999999999",
//	    Text: "Hello from wi-api",
//	})
//
// # Base URL
//
// Defaults to https://endpoint.wi.api.br. Override with [WithBaseURL].
package wi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://endpoint.wi.api.br"
const defaultTimeout = 30 * time.Second

// Client is the wi-api HTTP client. Create one with [New] or [NewWithOptions].
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Option configures a [Client].
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithHTTPClient replaces the default http.Client.
// Useful for custom transports, proxies, or mTLS.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// WithTimeout sets the per-request timeout. Default: 30s.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.http.Timeout = d }
}

// New creates a client with the given API key and optional [Option]s.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Session returns a [SessionClient] scoped to the given session ID.
func (c *Client) Session(id string) *SessionClient {
	return &SessionClient{id: id, client: c}
}

// ─── Internal HTTP ────────────────────────────────────────────────────────────

func (c *Client) post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("wi: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("wi: build request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("wi: request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return parseErrorResponse(res)
	}

	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			return fmt.Errorf("wi: decode response: %w", err)
		}
	}
	return nil
}
