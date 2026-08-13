package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "host only adds api prefix and https",
			input:    "coolify.example.com",
			expected: "https://coolify.example.com/api/v1",
		},
		{
			name:     "plain base appends api prefix",
			input:    "https://coolify.example.com",
			expected: "https://coolify.example.com/api/v1",
		},
		{
			name:     "existing api prefix is preserved",
			input:    "https://coolify.example.com/api/v1",
			expected: "https://coolify.example.com/api/v1",
		},
		{
			name:     "existing path keeps prefix and appends api",
			input:    "https://coolify.example.com/coolify",
			expected: "https://coolify.example.com/coolify/api/v1",
		},
		{
			name:     "query and fragment are stripped from base",
			input:    "https://coolify.example.com/api/v1?foo=bar#section",
			expected: "https://coolify.example.com/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeBaseURL(tt.input).String()
			if got != tt.expected {
				t.Fatalf("normalizeBaseURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildRequestURL(t *testing.T) {
	base := normalizeBaseURL("https://coolify.example.com")

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "absolute API path is appended under api prefix",
			path:     "/applications",
			expected: "https://coolify.example.com/api/v1/applications",
		},
		{
			name:     "relative API path is appended under api prefix",
			path:     "applications",
			expected: "https://coolify.example.com/api/v1/applications",
		},
		{
			name:     "query only keeps base path",
			path:     "?team=main",
			expected: "https://coolify.example.com/api/v1?team=main",
		},
		{
			name:     "path query is preserved",
			path:     "/deploy?uuid=abc",
			expected: "https://coolify.example.com/api/v1/deploy?uuid=abc",
		},
		{
			name:     "absolute URL bypasses base",
			path:     "https://other.example.com/status",
			expected: "https://other.example.com/status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildRequestURL(base, tt.path)
			if err != nil {
				t.Fatalf("buildRequestURL returned error: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("buildRequestURL(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestClientRequestTargetsAPIV1(t *testing.T) {
	type payload struct {
		OK bool `json:"ok"`
	}

	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload{OK: true})
	}))
	defer server.Close()

	client := NewClient(server.URL, "token")

	var out payload
	if err := client.Get(context.Background(), "/applications", &out); err != nil {
		t.Fatalf("client.Get returned error: %v", err)
	}
	if gotPath != "/api/v1/applications" {
		t.Fatalf("request path = %q, want %q", gotPath, "/api/v1/applications")
	}
	if !out.OK {
		t.Fatalf("decoded response mismatch: %#v", out)
	}
}
