package cmd

import (
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of Coolify",
	Long:  "Remove stored Coolify credentials from this machine.",
	RunE:  runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	ui.Section("Logout")

	confirm, err := ui.Confirm("Remove all stored credentials?")
	if err != nil {
		return err
	}
	if !confirm {
		ui.Dim("Canceled")
		return nil
	}

	ui.Spacer()
	globalErr := config.ClearGlobal()
	instanceErr := config.ClearInstances()
	if globalErr != nil && instanceErr != nil {
		ui.Warning("No credentials found")
		return nil
	}
	if globalErr != nil {
		ui.Warning("Failed to clear global credentials")
	}
	if instanceErr != nil {
		ui.Warning("Failed to clear instance credentials")
	}

	ui.Success("Logged out successfully")
	ui.Spacer()
	ui.Dim("Run 'cool-kit login' to authenticate again")
	return nil
}
