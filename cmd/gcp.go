package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/gcp"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Manage Coolify deployments on Google Cloud Platform",
	Long: `Deploy and manage Coolify instances on Google Cloud Platform.

This command group provides tools to provision Compute Engine instances, deploy Coolify,
manage updates, and configure GCP-hosted instances.`,
}

var gcpDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify to GCP",
	Long: `Deploy a new Coolify instance to Google Cloud Platform.

This command will:
- Validate GCP credentials
- Clone Coolify repository
- Create VPC network
- Configure firewall rules
- Launch Compute Engine instance
- Assign static IP
- Install Docker
- Deploy Coolify containers
- Configure Cloud DNS (optional)
- Run health checks

The deployment process is interactive and will guide you through all steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		return runGCPDeploy(useTUI)
	},
}

var gcpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of GCP Coolify instance",
	Long: `Check the status of your GCP-hosted Coolify instance.

Shows information about:
- Compute Engine instance status
- Container status
- Resource usage
- Network connectivity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return gcp.CheckStatus()
	},
}

var gcpSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into GCP Coolify instance",
	Long: `Open an SSH connection to your GCP Coolify instance.

This provides direct terminal access to the Compute Engine instance for debugging and
manual operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return gcp.SSHIntoInstance()
	},
}

func init() {
	// Add subcommands
	gcpCmd.AddCommand(gcpDeployCmd)
	gcpCmd.AddCommand(gcpStatusCmd)
	gcpCmd.AddCommand(gcpSSHCmd)

	// Add flags for deploy command
	gcpDeployCmd.Flags().Bool("tui", true, "Use interactive TUI for deployment progress")

	// Add gcp command to root
	rootCmd.AddCommand(gcpCmd)
}

func runGCPDeploy(useTUI bool) error {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Create GCP provider
	provider, err := gcp.NewGCPProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create GCP provider: %w", err)
	}

	// Create deployment runner
	runner := ui.NewDeploymentRunner("GCP", provider)

	if useTUI {
		return runner.RunWithTUI()
	}
	return runner.RunSimple()
}
