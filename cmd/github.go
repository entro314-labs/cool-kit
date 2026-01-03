package cmd

import (
	"context"
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/service"
	"github.com/spf13/cobra"
)

var githubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "Manage GitHub App integrations",
	Long:    "Manage GitHub App integrations in your Coolify instance for private repo deployments.",
}

var githubListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all GitHub Apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewGitHubAppService(client)
		apps, err := svc.List(context.Background())
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, apps)
	},
}

var githubGetCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get a GitHub App by UUID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewGitHubAppService(client)
		app, err := svc.Get(context.Background(), args[0])
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, app)
	},
}

var githubReposCmd = &cobra.Command{
	Use:   "repos <app-uuid>",
	Short: "List repositories accessible by a GitHub App",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		svc := service.NewGitHubAppService(client)
		repos, err := svc.ListRepositories(context.Background(), args[0])
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, repos)
	},
}

var githubBranchesCmd = &cobra.Command{
	Use:   "branches <app-uuid> <owner/repo>",
	Short: "List branches for a repository",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		// Parse owner/repo
		ownerRepo := args[1]
		var owner, repo string
		if n, _ := fmt.Sscanf(ownerRepo, "%s/%s", &owner, &repo); n != 2 {
			// Split manually
			parts := splitOwnerRepo(ownerRepo)
			if len(parts) != 2 {
				return fmt.Errorf("invalid format: use 'owner/repo'")
			}
			owner, repo = parts[0], parts[1]
		}

		svc := service.NewGitHubAppService(client)
		branches, err := svc.ListBranches(context.Background(), args[0], owner, repo)
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		return formatOutput(format, branches)
	},
}

var githubDeleteCmd = &cobra.Command{
	Use:     "delete <uuid>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a GitHub App",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete GitHub App '%s'? This cannot be undone.\n", args[0])
			fmt.Print("Type 'yes' to confirm: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		svc := service.NewGitHubAppService(client)
		if err := svc.Delete(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("✅ GitHub App '%s' deleted\n", args[0])
		return nil
	},
}

func splitOwnerRepo(ownerRepo string) []string {
	for i := 0; i < len(ownerRepo); i++ {
		if ownerRepo[i] == '/' {
			return []string{ownerRepo[:i], ownerRepo[i+1:]}
		}
	}
	return []string{ownerRepo}
}

func init() {
	// Add format flags
	githubListCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	githubGetCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	githubReposCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	githubBranchesCmd.Flags().String("format", "table", "Output format: table, json, pretty")
	githubDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	// Wire commands
	githubCmd.AddCommand(githubListCmd)
	githubCmd.AddCommand(githubGetCmd)
	githubCmd.AddCommand(githubReposCmd)
	githubCmd.AddCommand(githubBranchesCmd)
	githubCmd.AddCommand(githubDeleteCmd)
}
