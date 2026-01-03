package hetzner

import (
	"fmt"
	"os"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// CheckStatus checks the status of a Hetzner Coolify deployment
func CheckStatus() error {
	ui.Section("Hetzner Cloud Coolify Status")

	// Get token
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		if err := config.Initialize(); err == nil {
			cfg := config.Get()
			if cfg != nil {
				if t, ok := cfg.Settings["hetzner_token"].(string); ok {
					token = t
				}
			}
		}
	}

	if token == "" {
		ui.Error("No Hetzner Cloud token found")
		ui.Info("Set HCLOUD_TOKEN environment variable or configure in cool-kit config")
		return nil
	}

	client, err := NewClient(token)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create client: %v", err))
		return nil
	}

	ui.Success("API credentials valid")

	// Find Coolify server
	server, err := client.GetServerByLabel("application", "coolify")
	if err != nil {
		ui.Warning(fmt.Sprintf("Failed to find server: %v", err))
		return nil
	}

	if server == nil {
		ui.Warning("No Coolify server found")
		ui.Info("Deploy with: cool-kit hetzner deploy")
		return nil
	}

	displayServerStatus(server)

	// If running, try SSH checks
	if server.Status == "running" && server.PublicIPv4 != "" {
		checkSSHServices(server.PublicIPv4)
	}

	return nil
}

func displayServerStatus(server *ServerInfo) {
	ui.Info("Server Information")

	switch server.Status {
	case "running":
		ui.Success(fmt.Sprintf("  Status: %s", server.Status))
	case "off", "stopping":
		ui.Error(fmt.Sprintf("  Status: %s", server.Status))
	default:
		ui.Warning(fmt.Sprintf("  Status: %s", server.Status))
	}

	ui.Dim(fmt.Sprintf("  Name: %s", server.Name))
	ui.Dim(fmt.Sprintf("  ID: %d", server.ID))
	ui.Dim(fmt.Sprintf("  Type: %s", server.ServerType))
	ui.Dim(fmt.Sprintf("  Location: %s", server.Location))
	ui.Dim(fmt.Sprintf("  IPv4: %s", server.PublicIPv4))
	ui.Dim(fmt.Sprintf("  Created: %s", server.Created.Format("2006-01-02 15:04:05")))
}

func checkSSHServices(ip string) {
	ui.Info("Services (via SSH)")

	client := NewSSHClient(ip, "root")
	if err := client.TestConnection(); err != nil {
		ui.Warning(fmt.Sprintf("  SSH unavailable: %v", err))
		return
	}

	ui.Success("  SSH connection available")

	// Check Docker containers
	output, err := client.Execute(`for c in coolify coolify-db coolify-redis coolify-realtime; do
    status=$(docker inspect --format='{{.State.Running}}' $c 2>/dev/null || echo "notfound")
    echo "$c:$status"
done`)
	if err != nil {
		ui.Warning("  Could not check containers")
		return
	}

	for _, line := range splitLines(output) {
		parts := splitColon(line)
		if len(parts) < 2 {
			continue
		}
		name, status := parts[0], parts[1]
		if status == "notfound" {
			continue
		}
		if status == "true" {
			ui.Success(fmt.Sprintf("  %s: running", name))
		} else {
			ui.Error(fmt.Sprintf("  %s: stopped", name))
		}
	}
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	var current string
	for _, c := range s {
		if c == '\n' {
			if current != "" {
				lines = append(lines, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitColon(s string) []string {
	var parts []string
	var current string
	for _, c := range s {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}
