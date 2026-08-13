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
	parsedURL := normalizeBaseURL(baseURL)

	client := &Client{
		BaseURL: parsedURL,
		token:   token,
		debug:   os.Getenv("COOLKIT_DEBUG") != "" || os.Getenv("CDP_DEBUG") != "",
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

func normalizeBaseURL(raw string) *url.URL {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "https://localhost"
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	parsedURL, err := url.Parse(raw)
	if err != nil || parsedURL.Host == "" {
		host := strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://")
		host = strings.TrimSuffix(host, "/")
		parsedURL = &url.URL{
			Scheme: "https",
			Host:   host,
		}
	}

	basePath := strings.TrimSuffix(parsedURL.Path, "/")
	if basePath == "" {
		basePath = "/api/v1"
	} else if !strings.HasSuffix(basePath, "/api/v1") {
		basePath = basePath + "/api/v1"
	}

	parsedURL.Path = basePath
	parsedURL.RawPath = ""
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""

	return parsedURL
}

func buildRequestURL(base *url.URL, requestPath string) (string, error) {
	if base == nil {
		return "", fmt.Errorf("base URL is not configured")
	}
	if requestPath == "" {
		return base.String(), nil
	}

	parsedPath, err := url.Parse(requestPath)
	if err != nil {
		return "", fmt.Errorf("invalid request path %q: %w", requestPath, err)
	}
	if parsedPath.IsAbs() {
		return parsedPath.String(), nil
	}

	u := *base
	basePath := strings.TrimSuffix(u.Path, "/")
	requestSegment := strings.TrimPrefix(parsedPath.Path, "/")

	switch {
	case requestSegment == "":
		u.Path = basePath
	case basePath == "":
		u.Path = "/" + requestSegment
	default:
		u.Path = basePath + "/" + requestSegment
	}

	u.RawQuery = parsedPath.RawQuery
	u.Fragment = parsedPath.Fragment

	return u.String(), nil
}

// Default configuration values
const (
	DefaultRetries = 3
	DefaultTimeout = 30 * time.Second
	MinRetryDelay  = 1 * time.Second
	MaxRetryDelay  = 10 * time.Second
)

// doRequest performs an HTTP request with context support (CAGC pattern)
func (c *Client) doRequest(ctx context.Context, method, path string, body any, v any) error {
	u, err := buildRequestURL(c.BaseURL, path)
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

	return c.doWithRetry(ctx, method, u, reqBody, v)
}

// doWithRetry executes request with exponential backoff
func (c *Client) doWithRetry(ctx context.Context, method, urlStr string, body []byte, v any) error {
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

		// Retry on specific server errors or rate limiting
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusInternalServerError ||
			resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {

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

// Get performs a GET request with context
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request with context
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Patch performs a PATCH request with context
func (c *Client) Patch(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPatch, path, body, result)
}

// Delete performs a DELETE request with context
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// GetWithParams performs a GET request with query parameters and context
func (c *Client) GetWithParams(ctx context.Context, path string, params map[string]string, result any) error {
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		path = path + "?" + values.Encode()
	}
	return c.Get(ctx, path, result)
}
