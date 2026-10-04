package phajay

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://payment-gateway.phajay.co"
	defaultTimeout = 30 * time.Second
)

// Phajay is a client for the Phajay payment gateway. It may be used concurrently
// after construction, provided its HTTP client is not modified.
type Phajay struct {
	key     string
	baseURL string
	client  *http.Client
	sandbox bool
}

// Option configures a Phajay client during construction.
type Option func(*Phajay)

// WithBaseURL overrides the gateway URL, for example for a local test server.
func WithBaseURL(baseURL string) Option {
	return func(p *Phajay) { p.baseURL = baseURL }
}

// WithHTTPClient supplies an HTTP client. A nil client keeps the default.
// Options are applied in order; a later WithHTTPClient replaces earlier timeout
// settings. The supplied client is copied so options do not modify it.
func WithHTTPClient(client *http.Client) Option {
	return func(p *Phajay) {
		if client != nil {
			copy := *client
			p.client = &copy
		}
	}
}

// WithTimeout sets the HTTP timeout. Zero disables the timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(p *Phajay) { p.client.Timeout = timeout }
}

// WithSandbox selects the documented test routes for payment links and bank QR
// codes. Use a portal test key. APIs without documented test routes return an
// error in sandbox mode before making an HTTP request.
func WithSandbox() Option {
	return func(p *Phajay) { p.sandbox = true }
}

// New constructs a client using a portal secret key (or test key for sandbox).
func New(key string, opts ...Option) *Phajay {
	p := &Phajay{
		key:     key,
		baseURL: defaultBaseURL,
		client:  &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
