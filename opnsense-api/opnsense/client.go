// SPDX-License-Identifier: AGPL-3.0-or-later
// Package opnsense is a hand-rolled Go client for the OPNsense REST API.
package opnsense

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultTimeout = 30 * time.Second

// GenericResponse is the common wrapper for all OPNsense API responses.
// UUID is populated on POST to add_item endpoints; RawBody carries the raw
// response bytes for caller-side decoding.
type GenericResponse struct {
	StatusCode int    `json:"-"`
	Status     string `json:"status"`
	Msg        string `json:"msg"`
	UUID       string `json:"uuid,omitempty"`
	RawBody    []byte `json:"-"`
}

// Client is the OPNsense REST API client.
type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// NewClient creates a new OPNsense API client.
func NewClient(baseURL, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Client) BaseURL() string             { return c.baseURL }
func (c *Client) HTTPClient() *http.Client   { return c.httpClient }
func (c *Client) SetHTTPClient(hc *http.Client) { c.httpClient = hc }
func (c *Client) SetBaseURL(u string)         { c.baseURL = strings.TrimRight(u, "/") }

type Option func(*Client)

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

func WithInsecureTLS() Option {
	return func(c *Client) {
		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		c.httpClient.Transport = transport
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

func WithRetry(maxRetries int, initialDelay time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Transport = &retryTransport{
			maxRetries:   maxRetries,
			initialDelay: initialDelay,
			base:         c.httpClient.Transport,
		}
	}
}

type retryTransport struct {
	maxRetries   int
	initialDelay time.Duration
	base         http.RoundTripper
}

func (t *retryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.base.RoundTrip(r) // retry omitted for v1
}

func BasicAuthHeader(key, secret string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(key+":"+secret))
}

type Request struct {
	Method      string
	Path        string
	Body        io.Reader
	ContentType string
}

// Do sends an authenticated request to the OPNsense API and returns the parsed
// GenericResponse. The caller decodes RawBody for typed results.
func (c *Client) Do(ctx context.Context, req Request) (*GenericResponse, error) {
	url := c.baseURL + req.Path
	var body io.Reader = http.NoBody
	if req.Body != nil {
		body = req.Body
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", BasicAuthHeader(c.apiKey, c.apiSecret))
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gr GenericResponse
	if err := json.Unmarshal(raw, &gr); err != nil && len(raw) > 0 {
		gr.RawBody = raw
		gr.StatusCode = resp.StatusCode
		return &gr, fmt.Errorf("opnsense: unmarshal error: %w", err)
	}
	gr.RawBody = raw
	gr.StatusCode = resp.StatusCode
	return &gr, nil
}

var _ = io.EOF
