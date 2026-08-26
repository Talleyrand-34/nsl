// SPDX-License-Identifier: MIT
// Package opnsenseapi adapts the github.com/t34/opnsense-api client to the
// nsl-graph Transport / Session seam.
package opnsenseapi

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/t34/opnsense-api/opnsense"
)

// Creds are the input to Open. They mirror SSH creds at the same shape and
// flow through the existing vault: APIKey / APISecret are encrypted at rest,
// decrypted once on vault unlock.
type Creds struct {
	APIKey    string
	APISecret string
	BaseURL   string // optional; defaults to https://<host>
}

// Transport implements configparser.Transport.
type Transport struct {
	InsecureTLS bool
}

// Open creates an OPNsense client and returns a Session that wraps it.
func (t Transport) Open(host string, rawCreds any) (Session, error) {
	creds, ok := rawCreds.(Creds)
	if !ok {
		return nil, fmt.Errorf("opnsenseapi: creds must be Creds, got %T", rawCreds)
	}
	baseURL := creds.BaseURL
	if baseURL == "" {
		baseURL = "https://" + host
	}
	opts := []opnsense.Option{}
	if t.InsecureTLS {
		opts = append(opts, opnsense.WithInsecureTLS())
	}
	c := opnsense.NewClient(baseURL, creds.APIKey, creds.APISecret, opts...)
	return &session{client: c, host: host}, nil
}

// Session is the seam-side interface.
type Session interface {
	Execute(command string) (string, error)
	Close() error
}

type session struct {
	client *opnsense.Client
	host   string
}

func (s *session) Close() error { return nil }

// Execute runs one OPNsense API call. command is shaped as
// "METHOD /api/{module}/{controller}/{command}[/{uuid}]\n\n<body>".
func (s *session) Execute(command string) (string, error) {
	method, path, body := splitCommand(command)
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	resp, err := s.client.Do(safeCtx(), opnsense.Request{
		Method: method,
		Path:   path,
		Body:   bodyReader,
	})
	if err != nil {
		return "", err
	}
	return string(resp.RawBody), nil
}

func splitCommand(command string) (method, path, body string) {
	parts := strings.SplitN(command, "\n\n", 2)
	if len(parts) == 2 {
		body = parts[1]
	}
	line := parts[0]
	head := strings.SplitN(line, " ", 2)
	if len(head) != 2 {
		return "GET", line, ""
	}
	return head[0], head[1], body
}

func safeCtx() context.Context { return context.Background() }

// InsecureTLSWarn is the warning rendered to stderr when the lab requires
// TLS skip. Production builds must override via env.
const InsecureTLSWarn = "OPNSENSE_INSECURE_TLS=1: skipping cert verification (lab only)"

// TransportFromCreds is a convenience for tests.
func TransportFromCreds(c *opnsense.Client, creds Creds) Session {
	return &session{client: c, host: ""}
}

var _ = io.EOF // pin the import marker; the package uses io.Reader for the request body
