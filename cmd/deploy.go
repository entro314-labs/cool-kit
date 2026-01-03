package cmd

import (
	"fmt"
	"os"

	"github.com/entro314-labs/cool-kit/internal/api"
	"github.com/entro314-labs/cool-kit/internal/appdeploy"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/smart"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	deployProdFlag    bool
	deployPreviewFlag bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current directory to Coolify",
	Long: `Deploy the current project to Coolify.

Use --prod to explicitly deploy to production (default).
Use --preview to create a preview deployment for testing.

Examples:
  cool-kit deploy              # Deploy to production (default)
  cool-kit deploy --prod       # Explicitly deploy to production
  cool-kit deploy --preview    # Create preview deployment`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeploy()
	},
}

func init() {
	deployCmd.Flags().BoolVar(&deployProdFlag, "prod", false, "Deploy to production (default)")
	deployCmd.Flags().BoolVar(&deployPreviewFlag, "preview", false, "Create preview deployment")
}

func runDeploy() error {
	if err := checkLogin(); err != nil {
		return err
	}

	globalCfg, err := config.LoadGlobal()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	projectCfg, err := config.LoadProject()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	client := api.NewClient(globalCfg.CoolifyURL, globalCfg.CoolifyToken)

	isFirstDeploy := false
	var deploymentConfig *smart.DeploymentConfig

	// First-time setup if no project config exists
	if projectCfg == nil {
		ui.Section("New Project Setup")
		ui.Dim("Let's configure your project for deployment")

		setupResult, err := appdeploy.FirstTimeSetup(client, globalCfg)
		if err != nil {
			return err
		}
		projectCfg = setupResult.ProjectConfig
		deploymentConfig = setupResult.DeploymentConfig
		isFirstDeploy = true
	}

	// Determine deployment type based on flags
	var prNumber int
	var deploymentType string

	if deployPreviewFlag && deployProdFlag {
		return fmt.Errorf("cannot use both --prod and --preview flags")
	}

	if deployPreviewFlag {
		// Preview deployment - use PR number 1 for manual previews
		prNumber = 1
		deploymentType = "preview"
	} else {
		// Production deployment (default)
		prNumber = 0
		deploymentType = "production"
	}

	ui.Spacer()

	// Confirm deployments (except first deploy)
	if !isFirstDeploy {
		confirmMsg := fmt.Sprintf("Deploy to %s?", deploymentType)
		confirmed, err := ui.Confirm(confirmMsg)
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Dim("Deployment cancelled")
			return nil
		}
		// Confirmation already leaves a blank line, so just show the title
		ui.Bold("Deploy")
		ui.Spacer()
	} else {
		ui.Section("Deploy")
	}
	ui.KeyValue("Project", projectCfg.Name)
	ui.KeyValue("Type", deploymentType)
	ui.KeyValue("Method", projectCfg.DeployMethod)

	// Check verbose mode
	verbose := IsVerbose()

	// Deploy based on method
	if projectCfg.DeployMethod == config.DeployMethodDocker {
		return appdeploy.DeployDocker(client, globalCfg, projectCfg, deploymentConfig, prNumber, verbose)
	}
	return appdeploy.DeployGit(client, globalCfg, projectCfg, deploymentConfig, prNumber, verbose)
}
