package digitalocean

import (
	"fmt"
	"os"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

// CheckStatus checks the status of a DigitalOcean Coolify deployment
func CheckStatus() error {
	ui.Section("DigitalOcean Coolify Status")

	token := getToken()
	if token == "" {
		ui.Error("No DigitalOcean token found")
		ui.Info("Set DIGITALOCEAN_TOKEN environment variable or configure in cool-kit config")
		return nil
	}

	client, err := NewClient(token)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create client: %v", err))
		return nil
	}

	// Validate credentials
	account, err := client.GetAccount()
	if err != nil {
		ui.Error(fmt.Sprintf("API authentication failed: %v", err))
		return nil
	}
	ui.Success(fmt.Sprintf("Authenticated as: %s", account.Email))

	// Find Coolify droplet
	droplet, err := client.GetDropletByTag("coolify")
	if err != nil {
		ui.Warning(fmt.Sprintf("Failed to find droplet: %v", err))
		return nil
	}

	if droplet == nil {
		ui.Warning("No Coolify droplet found")
		ui.Info("Deploy with: cool-kit digitalocean deploy")
		return nil
	}

	displayDropletStatus(droplet)

	// If running, check SSH and services
	if droplet.Status == "active" && droplet.PublicIP != "" {
		checkSSHServices(droplet.PublicIP)
	}

	return nil
}

func getToken() string {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		if err := config.Initialize(); err == nil {
			cfg := config.Get()
			if cfg != nil {
				if t, ok := cfg.Settings["digitalocean_token"].(string); ok {
					token = t
				}
			}
		}
	}
	return token
}

func displayDropletStatus(droplet *DropletInfo) {
	ui.Info("Droplet Information")

	switch droplet.Status {
	case "active":
		ui.Success(fmt.Sprintf("  Status: %s", droplet.Status))
	case "off", "archive":
		ui.Error(fmt.Sprintf("  Status: %s", droplet.Status))
	default:
		ui.Warning(fmt.Sprintf("  Status: %s", droplet.Status))
	}

	ui.Dim(fmt.Sprintf("  Name: %s", droplet.Name))
	ui.Dim(fmt.Sprintf("  ID: %d", droplet.ID))
	ui.Dim(fmt.Sprintf("  Size: %s", droplet.Size))
	ui.Dim(fmt.Sprintf("  Region: %s", droplet.Region))
	ui.Dim(fmt.Sprintf("  IP: %s", droplet.PublicIP))
	ui.Dim(fmt.Sprintf("  Created: %s", droplet.Created.Format("2006-01-02 15:04:05")))
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
