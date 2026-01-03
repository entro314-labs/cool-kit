package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/docker"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Manage Coolify deployments using Docker Compose",
	Long: `Deploy and manage Coolify instances using Docker and Docker Compose.

This command group provides tools to deploy Coolify locally or on any server
with Docker installed, supporting different profiles (dev, staging, production).`,
}

var dockerDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify using Docker Compose",
	Long: `Deploy Coolify using Docker and Docker Compose.

This command will:
- Validate Docker installation
- Clone Coolify repository
- Generate secure credentials
- Configure environment
- Pull Docker images
- Start services
- Run health checks

Supports profiles: dev, staging, production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		useTUI, _ := cmd.Flags().GetBool("tui")
		return runDockerDeploy(profile, useTUI)
	},
}

var dockerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of Docker Coolify deployment",
	Long: `Check the status of your Docker-based Coolify deployment.

Shows information about:
- Docker daemon status
- Container health and status
- Resource usage
- Service endpoints`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return docker.CheckStatus()
	},
}

var dockerLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View logs from Docker Coolify services",
	Long: `View logs from Docker Coolify services.

If no service is specified, shows logs from all services.
Available services: coolify, coolify-db, coolify-redis, coolify-realtime`,
	RunE: func(cmd *cobra.Command, args []string) error {
		service := ""
		if len(args) > 0 {
			service = args[0]
		}
		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")
		return docker.ViewLogs(service, follow, tail)
	},
}

var dockerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Docker Coolify services",
	Long: `Stop all Docker Coolify services.

This preserves data and configuration. Use 'cool-kit docker start' to restart.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return docker.StopServices()
	},
}

var dockerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Docker Coolify services",
	Long: `Start previously stopped Docker Coolify services.

Services must have been previously deployed using 'cool-kit docker deploy'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return docker.StartServices()
	},
}

func init() {
	// Add flags for deploy command
	dockerDeployCmd.Flags().String("profile", "dev", "Deployment profile (dev, staging, production)")
	dockerDeployCmd.Flags().Bool("tui", true, "Use interactive TUI for deployment progress")

	// Add flags for logs command
	dockerLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	dockerLogsCmd.Flags().IntP("tail", "n", 100, "Number of lines to show from the end")

	// Add subcommands
	dockerCmd.AddCommand(dockerDeployCmd)
	dockerCmd.AddCommand(dockerStatusCmd)
	dockerCmd.AddCommand(dockerLogsCmd)
	dockerCmd.AddCommand(dockerStartCmd)
	dockerCmd.AddCommand(dockerStopCmd)

	// Add docker command to root
	rootCmd.AddCommand(dockerCmd)
}

func runDockerDeploy(profile string, useTUI bool) error {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Create Docker provider
	provider, err := docker.NewDockerProvider(cfg, profile)
	if err != nil {
		return fmt.Errorf("failed to create Docker provider: %w", err)
	}

	// Create deployment runner
	runner := ui.NewDeploymentRunner("Docker", provider)

	if useTUI {
		return runner.RunWithTUI()
	}
	return runner.RunSimple()
}
