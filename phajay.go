package phajay

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://payment-gateway.phajay.co"
	defaultTimeout = 30 * time.Second
)

type Phajay struct {
	key     string
	baseURL string
	client  *http.Client
}

type Option func(*Phajay)

func WithBaseURL(baseURL string) Option {
	return func(p *Phajay) { p.baseURL = baseURL }
}

func WithHTTPClient(client *http.Client) Option {
	return func(p *Phajay) { p.client = client }
}

func WithTimeout(timeout time.Duration) Option {
	return func(p *Phajay) { p.client.Timeout = timeout }
}

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
