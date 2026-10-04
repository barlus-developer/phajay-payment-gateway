package phajay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
)

type requestAuth int

const (
	authBasic requestAuth = iota
	authSecretKey
)

// APIError reports a non-2xx HTTP response from the gateway. Body contains at
// most 1 MiB of the response body. Use errors.As to inspect it.
type APIError struct {
	StatusCode int
	Body       string
	Operation  string
}

// Error returns the gateway status and response body.
func (e *APIError) Error() string {
	return fmt.Sprintf("phajay: %s request failed with status %d: %s", e.Operation, e.StatusCode, e.Body)
}

func (p *Phajay) do(ctx context.Context, method, path string, body any, auth requestAuth, operation string, response any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("phajay: marshal %s request: %w", operation, err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(p.baseURL, "/")+path, reader)
	if err != nil {
		return fmt.Errorf("phajay: build %s request: %w", operation, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if auth == authBasic {
		// Phajay encodes the raw key, without a username/password separator.
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(p.key)))
	} else {
		req.Header.Set("secretKey", p.key)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("phajay: send %s request: %w", operation, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return fmt.Errorf("phajay: read %s error response: %w", operation, err)
		}
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(data)), Operation: operation}
	}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("phajay: decode %s response: %w", operation, err)
	}
	return nil
}

func validatePayment(amount float64, description string) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return fmt.Errorf("phajay: amount must be finite and greater than zero")
	}
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("phajay: description is required")
	}
	return nil
}

func pathID(id, name string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("phajay: %s is required", name)
	}
	if id == "." || id == ".." {
		return "", fmt.Errorf("phajay: invalid %s", name)
	}
	return url.PathEscape(id), nil
}

func (p *Phajay) requireProduction(operation string) error {
	if p.sandbox {
		return fmt.Errorf("phajay: %s has no documented sandbox endpoint", operation)
	}
	return nil
}
