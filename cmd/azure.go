package cmd

import (
	"github.com/entro314-labs/cool-kit/internal/providers/azure"
	"github.com/spf13/cobra"
)

var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Manage Coolify deployments on Azure",
	Long: `Deploy and manage Coolify instances on Microsoft Azure.

This command group provides tools to provision Azure VMs, deploy Coolify,
manage updates, create backups, and handle rollbacks for Azure-hosted instances.`,
}

var azureDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify to Azure",
	Long: `Deploy a new Coolify instance to Microsoft Azure.

This command will:
- Provision an Azure VM with optimal settings
- Install Docker and required dependencies
- Deploy Coolify containers
- Configure networking and security
- Setup initial admin credentials
- Provide access URLs

The deployment process is interactive and will guide you through all steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.Deploy()
	},
}

var azureUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Coolify instance on Azure",
	Long: `Update an existing Coolify instance on Azure to the latest version.

This command will:
- Create a backup of the current instance
- Pull latest Coolify images
- Update containers
- Run database migrations
- Perform health checks
- Provide rollback option if update fails`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.Update()
	},
}

var azureStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of Azure Coolify instance",
	Long: `Check the status of your Azure-hosted Coolify instance.

Shows information about:
- VM status and health
- Container status
- Resource usage
- Network connectivity
- Recent logs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.Status()
	},
}

var azureBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create backup of Azure Coolify instance",
	Long: `Create a backup of your Azure Coolify instance.

Creates backups of:
- Application data
- Database
- Configuration files
- Docker volumes

Backups are stored on the Azure VM and can be used for rollback.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.Backup()
	},
}

var azureRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to previous version",
	Long: `Rollback Coolify instance to a previous version.

This command will:
- List available backups
- Allow you to select a backup point
- Restore from the selected backup
- Verify restoration success`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.Rollback()
	},
}

var azureConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Azure deployment configuration",
	Long: `Manage Azure deployment configuration.

Commands to view, edit, and validate Azure deployment configuration including:
- VM settings
- Network configuration
- Coolify settings
- Credentials`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.ManageConfig()
	},
}

var azureSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into Azure Coolify instance",
	Long: `Open an SSH connection to your Azure Coolify instance.

This provides direct terminal access to the Azure VM for debugging and
manual operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return azure.SSH()
	},
}

func init() {
	// Add subcommands
	azureCmd.AddCommand(azureDeployCmd)
	azureCmd.AddCommand(azureUpdateCmd)
	azureCmd.AddCommand(azureStatusCmd)
	azureCmd.AddCommand(azureBackupCmd)
	azureCmd.AddCommand(azureRollbackCmd)
	azureCmd.AddCommand(azureConfigCmd)
	azureCmd.AddCommand(azureSSHCmd)

	// Add azure command to root
	rootCmd.AddCommand(azureCmd)
}
