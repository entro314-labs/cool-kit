package cmd

import (
	"github.com/entro314-labs/cool-kit/internal/local"
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Manage local Coolify development instance",
	Long: `Manage a local Coolify development instance using Docker Compose.

This command group provides tools to setup, start, stop, update, and manage
a local Coolify instance for development and testing purposes.`,
}

var localSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup local Coolify development environment",
	Long: `Setup a complete local Coolify development environment.

This command will:
- Check Docker prerequisites
- Collect configuration information
- Generate secure credentials
- Create directory structure
- Setup Docker Compose configuration
- Start services
- Configure the application
- Perform health checks

The setup process is interactive and will guide you through all necessary steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return local.Setup()
	},
}

var localStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local Coolify instance",
	Long: `Start an existing local Coolify instance.

This command starts all Coolify services using Docker Compose.
The instance must have been previously configured using 'cool-kit local setup'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return local.Start()
	},
}

var localStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local Coolify instance",
	Long: `Stop the running local Coolify instance.

This command stops all Coolify services but preserves data and configuration.
Use 'cool-kit local start' to restart the instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return local.Stop()
	},
}

var localLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View logs from local Coolify instance",
	Long: `View logs from local Coolify services.

If no service is specified, shows logs from all services.
Available services: coolify, postgres, redis, soketi`,
	RunE: func(cmd *cobra.Command, args []string) error {
		service := ""
		if len(args) > 0 {
			service = args[0]
		}
		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")
		return local.Logs(service, follow, tail)
	},
}

var localResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset local Coolify instance to clean state",
	Long: `Reset local Coolify instance by removing all data and configuration.

WARNING: This command will delete all application data, databases, and configuration.
Use with caution - this action cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		return local.Reset(force)
	},
}

var localUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local Coolify instance to latest version",
	Long: `Update the local Coolify instance to the latest version.

This command will:
- Pull latest Docker images
- Recreate containers with new images
- Run database migrations
- Restart services`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return local.Update()
	},
}

func init() {
	// Add subcommands
	localCmd.AddCommand(localSetupCmd)
	localCmd.AddCommand(localStartCmd)
	localCmd.AddCommand(localStopCmd)
	localCmd.AddCommand(localLogsCmd)
	localCmd.AddCommand(localResetCmd)
	localCmd.AddCommand(localUpdateCmd)

	// Add flags for logs command
	localLogsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	localLogsCmd.Flags().IntP("tail", "n", 100, "Number of lines to show from the end of the logs")

	// Add flags for reset command
	localResetCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	// Add local command to root
	rootCmd.AddCommand(localCmd)
}
