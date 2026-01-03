package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup Coolify instance data",
	Long: `Backup Coolify instance data including:
  • Coolify configuration (/data/coolify/)
  • Docker volumes for running containers
  • SSH authorized keys

This command must be run on the Coolify server itself.

Examples:
  cool-kit backup
  cool-kit backup --output=/backups/coolify-backup.tar.gz
  cool-kit backup --include-volumes=false`,
	RunE: runBackup,
}

var (
	backupOutput         string
	backupIncludeVolumes bool
	backupStopDocker     bool
)

func init() {
	timestamp := time.Now().Format("2006-01-02-150405")
	defaultOutput := fmt.Sprintf("coolify-backup-%s.tar.gz", timestamp)

	backupCmd.Flags().StringVar(&backupOutput, "output", defaultOutput, "Output file path")
	backupCmd.Flags().BoolVar(&backupIncludeVolumes, "include-volumes", true, "Include Docker volumes from running containers")
	backupCmd.Flags().BoolVar(&backupStopDocker, "stop-docker", false, "Stop Docker before backup (recommended for consistency)")

	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	coolifyDir := "/data/coolify"

	// Check if we're on a Coolify server
	if _, err := os.Stat(coolifyDir); os.IsNotExist(err) {
		return fmt.Errorf("Coolify directory %s not found. This command must run on your Coolify server", coolifyDir)
	}

	ui.Section("Coolify Backup")

	// Collect paths to backup
	var paths []string
	paths = append(paths, coolifyDir)

	// Add SSH authorized keys
	sshKeys := filepath.Join(os.Getenv("HOME"), ".ssh", "authorized_keys")
	if _, err := os.Stat(sshKeys); err == nil {
		paths = append(paths, sshKeys)
		ui.Info("Including SSH authorized keys")
	}

	// Collect Docker volumes from running containers
	if backupIncludeVolumes {
		volumePaths, err := collectDockerVolumes()
		if err != nil {
			ui.Warning(fmt.Sprintf("Could not collect Docker volumes: %v", err))
		} else if len(volumePaths) > 0 {
			paths = append(paths, volumePaths...)
			ui.Info(fmt.Sprintf("Including %d Docker volumes", len(volumePaths)))
		}
	}

	// Calculate total size
	totalSize, err := calculateTotalSize(paths)
	if err != nil {
		ui.Warning(fmt.Sprintf("Could not calculate size: %v", err))
	} else {
		ui.Info(fmt.Sprintf("Total backup size: %s", totalSize))
	}

	// Optionally stop Docker
	if backupStopDocker {
		confirmed, err := ui.Confirm("This will stop Docker. Your services will be temporarily unavailable. Continue?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Aborted")
			return nil
		}

		ui.Info("Stopping Docker...")
		if err := exec.Command("systemctl", "stop", "docker").Run(); err != nil {
			return fmt.Errorf("failed to stop Docker: %w", err)
		}
		defer func() {
			ui.Info("Starting Docker...")
			exec.Command("systemctl", "start", "docker").Run()
		}()
	}

	// Create backup
	ui.Info(fmt.Sprintf("Creating backup: %s", backupOutput))

	tarArgs := []string{
		"--exclude=*.sock",
		"-Pczf", backupOutput,
		"-C", "/",
	}
	tarArgs = append(tarArgs, paths...)

	tarCmd := exec.Command("tar", tarArgs...)
	tarCmd.Stdout = os.Stdout
	tarCmd.Stderr = os.Stderr

	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Get backup file size
	if info, err := os.Stat(backupOutput); err == nil {
		ui.Success(fmt.Sprintf("Backup created: %s (%s)", backupOutput, formatBytes(info.Size())))
	} else {
		ui.Success(fmt.Sprintf("Backup created: %s", backupOutput))
	}

	ui.Spacer()
	ui.NextSteps([]string{
		"Transfer backup to new server: scp " + backupOutput + " user@newserver:/path/",
		"On new server: tar -Pxzf " + filepath.Base(backupOutput) + " -C /",
		"Install Coolify on new server: curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash",
	})

	return nil
}

// collectDockerVolumes gets volume paths from running containers
func collectDockerVolumes() ([]string, error) {
	// Get running container names
	out, err := exec.Command("docker", "ps", "--format", "{{.Names}}").Output()
	if err != nil {
		return nil, err
	}

	containers := strings.Split(strings.TrimSpace(string(out)), "\n")
	volumeSet := make(map[string]bool)

	for _, container := range containers {
		if container == "" {
			continue
		}
		// Get volume names for container
		volOut, err := exec.Command("docker", "inspect", "--format", "{{range .Mounts}}{{printf \"%s\\n\" .Name}}{{end}}", container).Output()
		if err != nil {
			continue
		}

		volumes := strings.Split(strings.TrimSpace(string(volOut)), "\n")
		for _, vol := range volumes {
			if vol != "" && !strings.HasPrefix(vol, "/") {
				volumePath := "/var/lib/docker/volumes/" + vol
				volumeSet[volumePath] = true
			}
		}
	}

	var paths []string
	for path := range volumeSet {
		if _, err := os.Stat(path); err == nil {
			paths = append(paths, path)
		}
	}

	return paths, nil
}

// calculateTotalSize calculates total size of paths
func calculateTotalSize(paths []string) (string, error) {
	// Try using du command first (most accurate)
	duArgs := append([]string{"-csh"}, paths...)
	out, err := exec.Command("du", duArgs...).Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "total") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					return parts[0], nil
				}
			}
		}
	}

	// Fallback: calculate by walking directories
	var total int64
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			// Walk directory to calculate actual size
			filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip errors
				}
				if !info.IsDir() {
					total += info.Size()
				}
				return nil
			})
		} else {
			total += info.Size()
		}
	}
	return formatBytes(total), nil
}

// formatBytes formats bytes to human readable
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
