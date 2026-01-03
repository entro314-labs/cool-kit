package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/entro314-labs/cool-kit/internal/api"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/output"
	"github.com/entro314-labs/cool-kit/internal/service"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:     "team",
	Aliases: []string{"teams"},
	Short:   "Manage teams",
	Long:    "Manage Coolify teams and team members.",
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewTeamService(client)
		teams, err := svc.List(context.Background())
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, teams)
	},
}

var teamCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current team",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewTeamService(client)
		team, err := svc.Current(context.Background())
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, team)
	},
}

var teamMembersCmd = &cobra.Command{
	Use:   "members [team_id]",
	Short: "List team members",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewTeamService(client)
		var members []service.TeamMember
		var listErr error

		if len(args) > 0 {
			members, listErr = svc.ListMembers(context.Background(), args[0])
		} else {
			members, listErr = svc.CurrentMembers(context.Background())
		}
		if listErr != nil {
			return listErr
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, members)
	},
}

// getAPIClient creates an API client from current configuration
func getAPIClient() (*api.Client, error) {
	cfg, err := config.LoadGlobal()
	if err != nil {
		return nil, fmt.Errorf("not logged in: run 'cool-kit login' first")
	}

	if cfg.CoolifyURL == "" || cfg.CoolifyToken == "" {
		return nil, fmt.Errorf("no Coolify instance configured. Run 'cool-kit login' first")
	}

	return api.NewClient(cfg.CoolifyURL, cfg.CoolifyToken), nil
}

// formatOutput formats and prints output
func formatOutput(format string, data interface{}) error {
	if format == "" {
		format = output.FormatTable
	}
	opts := output.Options{Writer: os.Stdout}
	formatter, err := output.NewFormatter(format, opts)
	if err != nil {
		return err
	}
	return formatter.Format(data)
}

func init() {
	// Add format flag
	teamListCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	teamCurrentCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	teamMembersCmd.Flags().String("format", "table", "Output format: table, json, pretty")

	// Wire commands
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamCurrentCmd)
	teamCmd.AddCommand(teamMembersCmd)
}
