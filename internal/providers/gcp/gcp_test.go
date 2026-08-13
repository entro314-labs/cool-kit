package gcp

import (
	"testing"

	"github.com/entro314-labs/cool-kit/internal/config"
)

func TestGetNetworkUsesStructuredConfig(t *testing.T) {
	cfg := &config.Config{
		GCP: config.GCPConfig{Network: "configured-network"},
	}
	provider, err := NewGCPProvider(cfg)
	if err != nil {
		t.Fatalf("failed to construct GCP provider: %v", err)
	}

	if got := provider.getNetwork(); got != "configured-network" {
		t.Fatalf("expected structured network value, got %q", got)
	}
}

func TestGetInstanceNameFallback(t *testing.T) {
	provider, err := NewGCPProvider(&config.Config{Settings: map[string]any{}})
	if err != nil {
		t.Fatalf("failed to construct GCP provider: %v", err)
	}

	if got := provider.getInstanceName(); got != "coolify-instance" {
		t.Fatalf("expected default instance name, got %q", got)
	}
}

func TestGetRegionDerivesFromZone(t *testing.T) {
	provider, err := NewGCPProvider(&config.Config{
		GCP: config.GCPConfig{Zone: "europe-west1-b"},
	})
	if err != nil {
		t.Fatalf("failed to construct GCP provider: %v", err)
	}

	if got := provider.getRegion(); got != "europe-west1" {
		t.Fatalf("expected region derived from zone, got %q", got)
	}
}
