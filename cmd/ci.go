package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Generate CI/CD workflow files",
	Long:  `Generate CI/CD workflow files for various platforms.`,
}

var ciGithubCmd = &cobra.Command{
	Use:   "github",
	Short: "Generate GitHub Actions workflow for Coolify deployment",
	Long: `Generate a GitHub Actions workflow file that deploys to Coolify.

The workflow triggers on push to main branch and deploys your application
using the Coolify API.

Examples:
  cool-kit ci github
  cool-kit ci github --app-uuid=abc123
  cool-kit ci github --branch=main`,
	RunE: runCIGithub,
}

var (
	ciAppUUID string
	ciBranch  string
)

func init() {
	ciGithubCmd.Flags().StringVar(&ciAppUUID, "app-uuid", "", "Application UUID (or use COOLIFY_APP_UUID secret)")
	ciGithubCmd.Flags().StringVar(&ciBranch, "branch", "main", "Branch to trigger deployment on")

	ciCmd.AddCommand(ciGithubCmd)
	rootCmd.AddCommand(ciCmd)
}

func runCIGithub(cmd *cobra.Command, args []string) error {
	workflowDir := ".github/workflows"
	workflowFile := filepath.Join(workflowDir, "coolify-deploy.yml")

	// Check if file already exists
	if _, err := os.Stat(workflowFile); err == nil {
		confirmed, err := ui.Confirm(fmt.Sprintf("%s already exists. Overwrite?", workflowFile))
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Info("Aborted")
			return nil
		}
	}

	// Create directory
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Determine app UUID placeholder
	appUUIDValue := "${{ secrets.COOLIFY_APP_UUID }}"
	if ciAppUUID != "" {
		appUUIDValue = ciAppUUID
	}

	// Generate workflow content
	workflow := fmt.Sprintf(`name: Deploy to Coolify

on:
  push:
    branches: [%s]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Coolify
        run: |
          curl -X POST \
            -H "Authorization: Bearer ${{ secrets.COOLIFY_TOKEN }}" \
            "${{ secrets.COOLIFY_URL }}/api/v1/deploy?uuid=%s"
        env:
          COOLIFY_TOKEN: ${{ secrets.COOLIFY_TOKEN }}
          COOLIFY_URL: ${{ secrets.COOLIFY_URL }}
`, ciBranch, appUUIDValue)

	// Write file
	if err := os.WriteFile(workflowFile, []byte(workflow), 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	ui.Success(fmt.Sprintf("Created %s", workflowFile))
	ui.Spacer()
	ui.Info("Required GitHub Secrets:")
	ui.List([]string{
		"COOLIFY_URL - Your Coolify instance URL (e.g., https://coolify.example.com)",
		"COOLIFY_TOKEN - Your Coolify API token",
	})
	if ciAppUUID == "" {
		ui.List([]string{
			"COOLIFY_APP_UUID - Your application UUID",
		})
	}
	ui.Spacer()
	ui.NextSteps([]string{
		"Add secrets to your GitHub repository settings",
		"Push to " + ciBranch + " branch to trigger deployment",
	})

	return nil
}
