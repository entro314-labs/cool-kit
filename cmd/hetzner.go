package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/hetzner"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var hetznerCmd = &cobra.Command{
	Use:   "hetzner",
	Short: "Manage Coolify deployments on Hetzner Cloud",
	Long: `Deploy and manage Coolify instances on Hetzner Cloud.

This command group provides tools to provision Hetzner Cloud servers, deploy Coolify,
manage instances, and configure deployments using the official Hetzner Cloud SDK.

Authentication:
  Set HCLOUD_TOKEN environment variable or configure in cool-kit config.
  
  Don't have a Hetzner account? Sign up with their referral link to support the Coolify project:
  https://coolify.io/hetzner (or https://console.hetzner.com/refer?pk_campaign=referral-invite&pk_medium=referral-program&pk_source=reflink&pk_content=VBVO47VycYLt)`,
}

var hetznerDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify to Hetzner Cloud",
	Long: `Deploy a new Coolify instance to Hetzner Cloud.

This command will:
- Validate Hetzner Cloud API credentials
- Configure SSH key for server access
- Create a Hetzner Cloud server
- Install Docker via cloud-init
- Deploy Coolify containers
- Run health checks

Requires: HCLOUD_TOKEN environment variable`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		return runHetznerDeploy(useTUI)
	},
}

var hetznerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of Hetzner Cloud Coolify instance",
	Long: `Check the status of your Hetzner Cloud-hosted Coolify instance.

Shows information about:
- Server status and health
- Container status
- Resource usage
- Network connectivity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hetzner.CheckStatus()
	},
}

var hetznerSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into Hetzner Cloud Coolify instance",
	Long: `Open an SSH connection to your Hetzner Cloud Coolify instance.

This provides direct terminal access to the server for debugging and
manual operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hetzner.SSHIntoInstance()
	},
}

var hetznerDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy Hetzner Cloud Coolify instance",
	Long: `Destroy your Hetzner Cloud Coolify instance.

WARNING: This will permanently delete the server and all data.
This action cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hetzner.DestroyInstance()
	},
}

func init() {
	// Add flags
	hetznerDeployCmd.Flags().Bool("tui", true, "Use interactive TUI for deployment progress")
	hetznerDeployCmd.Flags().String("server-type", "cx21", "Hetzner server type (e.g., cx21, cx31)")
	hetznerDeployCmd.Flags().String("location", "nbg1", "Datacenter location (nbg1, fsn1, hel1)")
	hetznerDeployCmd.Flags().String("image", "ubuntu-22.04", "Server image")

	// Add subcommands
	hetznerCmd.AddCommand(hetznerDeployCmd)
	hetznerCmd.AddCommand(hetznerStatusCmd)
	hetznerCmd.AddCommand(hetznerSSHCmd)
	hetznerCmd.AddCommand(hetznerDestroyCmd)

	// Add hetzner command to root
	rootCmd.AddCommand(hetznerCmd)
}

func runHetznerDeploy(useTUI bool) error {
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	provider, err := hetzner.NewHetznerProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Hetzner provider: %w", err)
	}

	runner := ui.NewDeploymentRunner("Hetzner Cloud", provider)

	if useTUI {
		return runner.RunWithTUI()
	}
	return runner.RunSimple()
}
