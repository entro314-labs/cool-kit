package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Coolify to the latest version",
	Long: `Update Coolify to the latest version on a deployed server.

This command connects to your server via SSH and runs the official
Coolify update script.

Examples:
  cool-kit update --host 192.168.1.100 --user root
  cool-kit update  # Uses last deployment settings`,
	RunE: runUpdate,
}

var (
	updateHost    string
	updateUser    string
	updateKeyPath string
	updateDryRun  bool
	updateForce   bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateHost, "host", "H", "", "Server IP or hostname")
	updateCmd.Flags().StringVarP(&updateUser, "user", "u", "root", "SSH username")
	updateCmd.Flags().StringVarP(&updateKeyPath, "key", "k", "", "SSH private key path")
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Show what would be done without executing")
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Skip confirmation prompt")

	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ui.Section("Update Coolify")

	// Get host from flags or config
	host := updateHost
	user := updateUser
	keyPath := updateKeyPath

	if host == "" {
		// Try to get from last deployment
		cfg := config.Get()
		if cfg.Settings != nil {
			if h, ok := cfg.Settings["public_ip"].(string); ok && h != "" {
				host = h
				ui.Info(fmt.Sprintf("Using host from last deployment: %s", host))
			}
		}
	}

	if host == "" {
		return fmt.Errorf("no host specified. Use --host flag or deploy first")
	}

	// Get user from config if not specified
	if user == "root" {
		cfg := config.Get()
		if cfg.Azure.AdminUsername != "" {
			user = cfg.Azure.AdminUsername
		}
	}

	// Show what will be done
	ui.Info(fmt.Sprintf("Target: %s@%s", user, host))

	if updateDryRun {
		ui.Warning("[DRY RUN] Would execute:")
		fmt.Println("  curl -fsSL https://cdn.coollabs.io/coolify/install.sh | sudo bash")
		return nil
	}

	// Confirm unless forced
	if !updateForce {
		confirm, err := ui.Confirm(fmt.Sprintf("Update Coolify on %s?", host))
		if err != nil {
			return err
		}
		if !confirm {
			return nil
		}
	}

	// Build SSH command
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=30",
	}

	if keyPath != "" {
		sshArgs = append(sshArgs, "-i", keyPath)
	}

	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", user, host))

	// Step 1: Check current version
	ui.Info("Checking current Coolify version...")

	checkCmd := exec.Command("ssh", append(sshArgs, "docker inspect coolify --format '{{.Config.Image}}' 2>/dev/null || echo 'unknown'")...)
	output, err := checkCmd.CombinedOutput()
	if err != nil && IsVerbose() {
		ui.Dim(fmt.Sprintf("Version check warning: %v", err))
	}
	currentVersion := strings.TrimSpace(string(output))
	ui.Dim(fmt.Sprintf("Current: %s", currentVersion))

	// Step 2: Run update script
	ui.Info("Running Coolify update script...")
	startTime := time.Now()

	updateScript := "curl -fsSL https://cdn.coollabs.io/coolify/install.sh | sudo bash"
	updateExecCmd := exec.Command("ssh", append(sshArgs, updateScript)...)

	// Stream output
	updateExecCmd.Stdout = nil // Suppress verbose output
	updateExecCmd.Stderr = nil

	if err := updateExecCmd.Run(); err != nil {
		return ui.NewDeploymentError("update", "Run update script", err)
	}

	elapsed := time.Since(startTime).Round(time.Second)

	// Step 3: Verify update
	ui.Info("Verifying update...")
	time.Sleep(5 * time.Second) // Wait for containers to restart

	verifyCmd := exec.Command("ssh", append(sshArgs, "docker inspect coolify --format '{{.Config.Image}}' 2>/dev/null || echo 'unknown'")...)
	output, err = verifyCmd.CombinedOutput()
	if err != nil {
		ui.Dim(fmt.Sprintf("Verification warning: %v", err))
	}
	newVersion := strings.TrimSpace(string(output))

	// Step 4: Health check
	healthCmd := exec.Command("ssh", append(sshArgs, "sudo docker ps --filter name=coolify --format '{{.Status}}' | head -1")...)
	output, err = healthCmd.CombinedOutput()
	if err != nil {
		ui.Dim(fmt.Sprintf("Health check warning: %v", err))
	}
	status := strings.TrimSpace(string(output))

	fmt.Println()
	ui.Success("Update complete!")
	fmt.Println()

	// Summary
	headers := []string{"Property", "Value"}
	rows := [][]string{
		{"Host", host},
		{"Previous Version", currentVersion},
		{"New Version", newVersion},
		{"Container Status", status},
		{"Duration", elapsed.String()},
	}
	ui.Table(headers, rows)

	if strings.Contains(status, "healthy") || strings.Contains(status, "Up") {
		ui.Success("Coolify is running and healthy")
	} else {
		ui.Warning("Container may still be starting. Check status with: cool-kit health")
	}

	ui.NextSteps([]string{
		fmt.Sprintf("Open your Coolify dashboard: http://%s:8000", host),
		"Check logs: ssh " + user + "@" + host + " 'docker logs coolify'",
	})

	return nil
}
