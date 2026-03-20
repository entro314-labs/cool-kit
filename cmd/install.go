package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/orchestrator"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Coolify on a server (Pillar 1)",
	Long: `Install Coolify on a server with automated configuration.

Part of Pillar 1: Deploy Coolify - Install Coolify on any infrastructure

This command will guide you through installing Coolify on:
- Azure
- AWS
- Google Cloud Platform (GCP)
- Bare Metal servers
- Local Docker environment`,
	RunE: runInstall,
}

func init() {
	installCmd.AddCommand(installAzureCmd)
	installCmd.AddCommand(installAWSCmd)
	installCmd.AddCommand(installGCPCmd)
	installCmd.AddCommand(installBareMetalCmd)
	installCmd.AddCommand(installLocalCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Start interactive TUI to select provider
	model := ui.NewModel() // This model should handle provider selection
	program := tea.NewProgram(model, tea.WithContext(cmd.Context()))

	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	m := finalModel.(ui.Model)

	if m.Provider != "" {
		return performInstall(cmd.Context(), m.Provider, m.Configuration)
	}

	return nil
}

func performInstall(ctx context.Context, provider string, overrides map[string]string) error {
	cfg := config.Get()
	if cfg == nil {
		// Try to init defaults if nil
		if err := config.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		cfg = config.Get()
	}

	// Apply overrides to config if needed
	if len(overrides) > 0 {
		if err := cfg.Merge(overrides); err != nil {
			return fmt.Errorf("failed to apply overrides: %w", err)
		}
	}

	o := orchestrator.NewOrchestrator(cfg, provider)
	result, err := o.Deploy(ctx)
	if err != nil {
		// Don't wrap the error again - TUI already showed it
		return err
	}

	if result.Success {
		return nil
	}
	return fmt.Errorf("installation failed")
}

var installAzureCmd = &cobra.Command{
	Use:          "azure",
	Short:        "Install on Azure",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return performInstall(cmd.Context(), "azure", nil)
	},
}

var installAWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "Install on AWS",
	RunE: func(cmd *cobra.Command, args []string) error {
		return performInstall(cmd.Context(), "aws", nil)
	},
}

var installGCPCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Install on GCP",
	RunE: func(cmd *cobra.Command, args []string) error {
		return performInstall(cmd.Context(), "gcp", nil)
	},
}

var installBareMetalCmd = &cobra.Command{
	Use:   "baremetal",
	Short: "Install on Bare Metal",
	RunE: func(cmd *cobra.Command, args []string) error {
		return performInstall(cmd.Context(), "baremetal", nil)
	},
}

var installLocalCmd = &cobra.Command{
	Use:   "local",
	Short: "Install Locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		return performInstall(cmd.Context(), "local", nil)
	},
}
