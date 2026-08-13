package cmd

import (
	"context"
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cool-kit",
	Short: "Cool Kit - Complete Coolify toolkit",
	Long: `Cool Kit is a comprehensive toolkit for Coolify with three pillars:

Pillar 1: Deploy Coolify
  Install and setup Coolify on any infrastructure with automated configuration

Pillar 2: Deploy Apps TO Coolify
  Deploy applications with Vercel-like experience - smart framework detection,
  automatic service provisioning (databases, caches), and one-command deployments

Pillar 3: Enhance Coolify
  Optimize and enhance existing Coolify installations with best practices

	Features:
	  • Smart framework detection (Next.js, Laravel, Node.js, and 30+ more)
	  • Automatic service provisioning (PostgreSQL, MySQL, MongoDB, Redis, etc.)
	  • Multi-instance management (switch between Coolify instances)
	  • Production and preview deployments
	  • Environment variable management
	  • Deployment monitoring and rollbacks`,
	PersistentPreRunE: initRuntimeContext,
	RunE:              runMainTUI,
	SilenceUsage:      true,
}

func initRuntimeContext(cmd *cobra.Command, _ []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to read verbose flag: %w", err)
	}
	SetVerbose(verbose)

	instanceName, err := cmd.Flags().GetString("instance")
	if err != nil {
		return fmt.Errorf("failed to read instance flag: %w", err)
	}
	SetInstanceOverride(instanceName)

	if err := config.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	return nil
}

// runMainTUI runs the main TUI menu and dispatches to subcommands
func runMainTUI(cmd *cobra.Command, args []string) error {
	selection, err := ui.RunMainMenu()
	if err != nil {
		return err
	}

	switch selection {
	case ui.SelectionInstall:
		return runInstall(cmd, args)
	case ui.SelectionDeploy:
		return runDeploy(cmd.Context())
	case ui.SelectionInit:
		return runInit(cmd, args)
	case ui.SelectionLink:
		return runLink(cmd, args)
	case ui.SelectionLogs:
		return runLogs(cmd, args)
	case ui.SelectionStatus:
		return runHealth(cmd, args)
	case ui.SelectionInstances:
		return runInstancesList(cmd, args)
	case ui.SelectionServices:
		return runServicesList(cmd, args)
	case ui.SelectionEnv:
		return runEnvLs(cmd, args)
	case ui.SelectionConfig:
		return configShowCmd.RunE(configShowCmd, args)
	case ui.SelectionBackup:
		return runBackup(cmd, args)
	case ui.SelectionBadge:
		return runBadgeInteractive()
	case ui.SelectionCI:
		return runCIGithub(cmd, args)
	case ui.SelectionHelp:
		return cmd.Help()
	case ui.SelectionExit:
		return nil
	}
	return nil
}

func Execute(ctx context.Context, version, commit, date, builtBy string) error {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s, by: %s)", version, commit, date, builtBy)

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().String("instance", "", "Target Coolify instance (overrides context)")
	rootCmd.PersistentFlags().StringP("format", "o", "table", "Output format (table, json, pretty)")

	// Pillar 1: Deploy Coolify
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(azureCmd)
	rootCmd.AddCommand(awsCmd)
	rootCmd.AddCommand(gcpCmd)
	rootCmd.AddCommand(hetznerCmd)
	rootCmd.AddCommand(digitaloceanCmd)
	rootCmd.AddCommand(baremetalCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(localCmd)

	// Pillar 2: Deploy TO Coolify
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(linkCmd)

	// Instance & Auth
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(instancesCmd)

	// Management
	rootCmd.AddCommand(servicesCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(teamCmd)
	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(githubCmd)

	// AI Integration
	rootCmd.AddCommand(mcpCmd)

	// Utilities
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(badgeCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(ciCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(destroyCmd)
}
