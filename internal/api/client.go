// Package api provides a client for the Coolify API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client is the Coolify API client with context-based operations
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
	token      string
	debug      bool
	retries    int
	timeout    time.Duration
}

// APIError represents an error from the Coolify API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// IsConflict returns true if the error is a 409 Conflict
func IsConflict(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 409
	}
	return false
}

// IsNotFound returns true if the error is a 404 Not Found
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// ClientOption is a functional option for configuring the client
type ClientOption func(*Client)

// WithDebug enables debug logging for API requests
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// WithRetries sets the maximum number of retries for failed requests
func WithRetries(retries int) ClientOption {
	return func(c *Client) {
		c.retries = retries
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Coolify API client with optional configuration
func NewClient(baseURL, token string, opts ...ClientOption) *Client {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	// Ensure it has /api/v1 suffix
	if !strings.HasSuffix(baseURL, "/api/v1") {
		baseURL = baseURL + "/api/v1"
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// Fall back to simple URL if parsing fails
		parsedURL = &url.URL{
			Scheme: "https",
			Host:   baseURL,
		}
	}

	client := &Client{
		BaseURL: parsedURL,
		token:   token,
		debug:   os.Getenv("CDP_DEBUG") != "",
		retries: DefaultRetries,
		timeout: DefaultTimeout,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Default configuration values
const (
	DefaultRetries = 3
	DefaultTimeout = 30 * time.Second
	MinRetryDelay  = 1 * time.Second
	MaxRetryDelay  = 10 * time.Second
)

// doRequest performs an HTTP request with context support (CAGC pattern)
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, v interface{}) error {
	u, err := c.BaseURL.Parse(path)
	if err != nil {
		return err
	}

	var reqBody []byte
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	return c.doWithRetry(ctx, method, u.String(), reqBody, v)
}

// doWithRetry executes request with exponential backoff
func (c *Client) doWithRetry(ctx context.Context, method, urlStr string, body []byte, v interface{}) error {
	var lastErr error
	delay := MinRetryDelay

	for i := 0; i <= c.retries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay *= 2
				if delay > MaxRetryDelay {
					delay = MaxRetryDelay
				}
			}
		}

		var req *http.Request
		var err error

		if body != nil {
			req, err = http.NewRequestWithContext(ctx, method, urlStr, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequestWithContext(ctx, method, urlStr, nil)
		}

		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

		if c.debug {
			fmt.Printf("[API] %s %s (Attempt %d/%d)\n", method, urlStr, i+1, c.retries+1)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue // Network error, retry
		}

		// Don't retry on client errors (4xx) except maybe 429?
		// For now simple logic: if 5xx retry, else return
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading error response: %v", err)
			}
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    string(bodyBytes),
			}
		}

		if v != nil {
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("request failed after %d retries: %w", c.retries, lastErr)
}

// Convenience methods that use context.Background() internally.
// For cancellation support and timeouts, use the context-based methods below directly.

// Get performs a GET request
func (c *Client) Get(path string, result interface{}) error {
	return c.doRequest(context.Background(), http.MethodGet, path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.doRequest(context.Background(), http.MethodPost, path, body, result)
}

// Patch performs a PATCH request
func (c *Client) Patch(path string, body interface{}, result interface{}) error {
	return c.doRequest(context.Background(), http.MethodPatch, path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) error {
	return c.doRequest(context.Background(), http.MethodDelete, path, nil, nil)
}

// Context-based methods for explicit cancellation and timeout control

// GetWithContext performs a GET request with context
func (c *Client) GetWithContext(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// PostWithContext performs a POST request with context
func (c *Client) PostWithContext(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// PatchWithContext performs a PATCH request with context
func (c *Client) PatchWithContext(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, result)
}

// DeleteWithContext performs a DELETE request with context
func (c *Client) DeleteWithContext(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// GetWithParams performs a GET request with query parameters
func (c *Client) GetWithParams(path string, params map[string]string, result interface{}) error {
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		path = path + "?" + values.Encode()
	}
	return c.Get(path, result)
}
