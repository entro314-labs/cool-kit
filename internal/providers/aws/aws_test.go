package aws

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/entro314-labs/cool-kit/internal/config"
)

func TestGetKeyNamePrecedence(t *testing.T) {
	t.Run("uses structured AWS key name", func(t *testing.T) {
		cfg := &config.Config{
			AWS: config.AWSConfig{KeyName: "structured-key"},
		}
		provider := NewAWSProvider(cfg)
		if got := provider.getKeyName(); got != "structured-key" {
			t.Fatalf("expected structured key name, got %q", got)
		}
	})

	t.Run("returns empty when key name is not configured", func(t *testing.T) {
		cfg := &config.Config{
			Settings: map[string]any{},
		}
		provider := NewAWSProvider(cfg)
		if got := provider.getKeyName(); got != "" {
			t.Fatalf("expected empty key name without structured config, got %q", got)
		}
	})
}

func TestGetSSHPrivateKeyPath(t *testing.T) {
	t.Run("strips .pub suffix and resolves existing private key path", func(t *testing.T) {
		dir := t.TempDir()
		privateKeyPath := filepath.Join(dir, "id_rsa")
		if err := os.WriteFile(privateKeyPath, []byte("dummy-private-key"), 0600); err != nil {
			t.Fatalf("failed to create private key fixture: %v", err)
		}

		cfg := &config.Config{
			AWS:      config.AWSConfig{SSHKeyPath: privateKeyPath + ".pub"},
			Settings: map[string]any{},
		}
		provider := NewAWSProvider(cfg)

		got, err := provider.getSSHPrivateKeyPath()
		if err != nil {
			t.Fatalf("expected key path resolution to succeed: %v", err)
		}
		if got != privateKeyPath {
			t.Fatalf("expected private key path %q, got %q", privateKeyPath, got)
		}
	})

	t.Run("returns error when key path is missing", func(t *testing.T) {
		cfg := &config.Config{
			AWS:      config.AWSConfig{},
			Settings: map[string]any{},
		}
		provider := NewAWSProvider(cfg)

		if _, err := provider.getSSHPrivateKeyPath(); err == nil {
			t.Fatal("expected missing SSH key path to return error")
		}
	})
}
