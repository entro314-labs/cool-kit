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
		ui.Dim("Cancelled")
		return nil
	}

	ui.Spacer()
	if err := config.ClearGlobal(); err != nil {
		// Ignore error if file doesn't exist
		ui.Warning("No credentials found")
		return nil
	}

	ui.Success("Logged out successfully")
	ui.Spacer()
	ui.Dim("Run 'cdp login' to authenticate again")
	return nil
}
