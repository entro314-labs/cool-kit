package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/digitalocean"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var digitaloceanCmd = &cobra.Command{
	Use:     "digitalocean",
	Aliases: []string{"do"},
	Short:   "Manage Coolify deployments on DigitalOcean",
	Long: `Deploy and manage Coolify instances on DigitalOcean.

This command group provides tools to provision DigitalOcean Droplets, deploy Coolify,
manage instances, and configure deployments using the official DigitalOcean SDK.

Authentication:
  Set DIGITALOCEAN_TOKEN environment variable or configure in cool-kit config.`,
}

var digitaloceanDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify to DigitalOcean",
	Long: `Deploy a new Coolify instance to DigitalOcean.

This command will:
- Validate DigitalOcean API credentials
- Configure SSH key for droplet access
- Create a DigitalOcean Droplet
- Install Docker via cloud-init
- Deploy Coolify containers
- Run health checks

Requires: DIGITALOCEAN_TOKEN environment variable`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		return runDigitalOceanDeploy(useTUI)
	},
}

var digitaloceanStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of DigitalOcean Coolify instance",
	Long: `Check the status of your DigitalOcean-hosted Coolify instance.

Shows information about:
- Droplet status and health
- Container status
- Resource usage
- Network connectivity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return digitalocean.CheckStatus()
	},
}

var digitaloceanSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into DigitalOcean Coolify instance",
	Long: `Open an SSH connection to your DigitalOcean Coolify instance.

This provides direct terminal access to the droplet for debugging and
manual operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return digitalocean.SSHIntoInstance()
	},
}

var digitaloceanDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy DigitalOcean Coolify instance",
	Long: `Destroy your DigitalOcean Coolify droplet.

WARNING: This will permanently delete the droplet and all data.
This action cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return digitalocean.DestroyInstance()
	},
}

func init() {
	// Add flags
	digitaloceanDeployCmd.Flags().Bool("tui", true, "Use interactive TUI for deployment progress")
	digitaloceanDeployCmd.Flags().String("size", "s-2vcpu-4gb", "Droplet size slug")
	digitaloceanDeployCmd.Flags().String("region", "nyc1", "Datacenter region")
	digitaloceanDeployCmd.Flags().String("image", "ubuntu-22-04-x64", "Droplet image slug")

	// Add subcommands
	digitaloceanCmd.AddCommand(digitaloceanDeployCmd)
	digitaloceanCmd.AddCommand(digitaloceanStatusCmd)
	digitaloceanCmd.AddCommand(digitaloceanSSHCmd)
	digitaloceanCmd.AddCommand(digitaloceanDestroyCmd)

	// Add digitalocean command to root
	rootCmd.AddCommand(digitaloceanCmd)
}

func runDigitalOceanDeploy(useTUI bool) error {
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	provider, err := digitalocean.NewDigitalOceanProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create DigitalOcean provider: %w", err)
	}

	runner := ui.NewDeploymentRunner("DigitalOcean", provider)

	if useTUI {
		return runner.RunWithTUI()
	}
	return runner.RunSimple()
}
