package cmd

import (
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/providers/aws"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Manage Coolify deployments on AWS",
	Long: `Deploy and manage Coolify instances on Amazon Web Services.

This command group provides tools to provision EC2 instances, deploy Coolify,
manage updates, and configure AWS-hosted instances.`,
}

var awsDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Coolify to AWS",
	Long: `Deploy a new Coolify instance to Amazon Web Services.

This command will:
- Validate AWS credentials
- Clone Coolify repository
- Create VPC and subnets
- Configure security groups
- Launch EC2 instance
- Assign Elastic IP
- Install Docker
- Deploy Coolify containers
- Configure SSL
- Run health checks

The deployment process is interactive and will guide you through all steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useTUI, _ := cmd.Flags().GetBool("tui")
		return runAWSDeploy(useTUI)
	},
}

var awsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of AWS Coolify instance",
	Long: `Check the status of your AWS-hosted Coolify instance.

Shows information about:
- EC2 instance status
- Container status
- Resource usage
- Network connectivity`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return aws.CheckStatus()
	},
}

var awsSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into AWS Coolify instance",
	Long: `Open an SSH connection to your AWS Coolify instance.

This provides direct terminal access to the EC2 instance for debugging and
manual operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return aws.SSHIntoInstance()
	},
}

func init() {
	// Add subcommands
	awsCmd.AddCommand(awsDeployCmd)
	awsCmd.AddCommand(awsStatusCmd)
	awsCmd.AddCommand(awsSSHCmd)

	// Add flags for deploy command
	awsDeployCmd.Flags().Bool("tui", true, "Use interactive TUI for deployment progress")

	// Add aws command to root
	rootCmd.AddCommand(awsCmd)
}

func runAWSDeploy(useTUI bool) error {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not initialized")
	}

	// Create AWS provider
	provider := aws.NewAWSProvider(cfg)

	// Create deployment runner
	runner := ui.NewDeploymentRunner("AWS", provider)

	if useTUI {
		return runner.RunWithTUI()
	}
	return runner.RunSimple()
}
