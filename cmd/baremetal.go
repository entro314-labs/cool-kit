package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/baremetal"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var baremetalCmd = &cobra.Command{
	Use:   "baremetal",
	Short: "Manage Coolify deployments on bare metal/VMs via SSH",
	Long: `Deploy and manage Coolify instances on bare metal servers or virtual machines via SSH.

This command group provides tools to deploy Coolify to any server with SSH access,
install dependencies, and manage the installation.`,
}

var baremetalDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify to a bare metal server or VM",
	Long: `Deploy Coolify to a bare metal server or virtual machine via SSH.

This command will:
- Validate SSH connectivity
- Check system requirements
- Clone Coolify repository
- Install Docker
- Install Docker Compose
- Configure environment
- Deploy Coolify containers
- Run health checks

Supports Ubuntu, Debian, CentOS, and RHEL distributions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		useTUI, _ := cmd.Flags().GetBool("tui")
		return runBareMetalDeploy(host, user, useTUI)
	},
}

var baremetalStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of bare metal Coolify instance",
	Long: `Check the status of your bare metal Coolify instance.

Shows information about:
- SSH connectivity
- Container status
- Resource usage
- System health`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return baremetal.CheckStatus()
	},
}

var baremetalSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into bare metal Coolify instance",
	Long: `Open an SSH connection to your bare metal Coolify instance.

This provides direct terminal access for debugging and manual operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return baremetal.SSHIntoInstance()
	},
}

func init() {
	// Add flags for deploy
	baremetalDeployCmd.Flags().String("host", "", "Target host IP or hostname (required)")
	baremetalDeployCmd.Flags().String("user", "root", "SSH username")
	baremetalDeployCmd.Flags().Bool("tui", true, "Use interactive TUI")

	// Add subcommands
	baremetalCmd.AddCommand(baremetalDeployCmd)
	baremetalCmd.AddCommand(baremetalStatusCmd)
	baremetalCmd.AddCommand(baremetalSSHCmd)

	// Add baremetal command to root
	rootCmd.AddCommand(baremetalCmd)
}

func runBareMetalDeploy(host, user string, useTUI bool) error {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Override config with flags if provided
	if host != "" {
		cfg.BareMetal.Host = host
	}
	if user != "" {
		cfg.BareMetal.User = user
	}

	// Validate host is provided
	if cfg.BareMetal.Host == "" {
		return fmt.Errorf("host is required. Use --host flag or set in configuration")
	}

	// Create bare metal provider
	provider, err := baremetal.NewBareMetalProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create bare metal provider: %w", err)
	}

	// Create deployment runner
	runner := ui.NewDeploymentRunner("Bare Metal", provider)

	if useTUI {
		return runner.RunWithTUI()
	}
	return runner.RunSimple()
}
