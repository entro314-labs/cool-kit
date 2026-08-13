package cmd

import (
	"fmt"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/ui"
	"github.com/spf13/cobra"
)

var badgeCmd = &cobra.Command{
	Use:   "badge [app-uuid]",
	Short: "Generate deployment status badge for README",
	Long: `Generate a deployment status badge URL for use in README files.

The badge shows live deployment status:
  • Green  - Successfully deployed
  • Red    - Deployment failed
  • Blue   - In progress
  • Gray   - Queued

Examples:
  cool-kit badge abc123def456
  cool-kit badge abc123def456 --format=markdown
  cool-kit badge abc123def456 --format=html`,
	Args: cobra.ExactArgs(1),
	RunE: runBadge,
}

var (
	badgeFormat  string
	badgeService string
)

func init() {
	badgeCmd.Flags().StringVarP(&badgeFormat, "format", "f", "markdown", "Output format (markdown, html, url)")
	badgeCmd.Flags().StringVar(&badgeService, "service", "", "Badge service URL (default: uses Coolify instance)")
}

func runBadge(cmd *cobra.Command, args []string) error {
	appUUID := args[0]

	// Determine badge URL
	var badgeURL string
	if badgeService != "" {
		badgeURL = fmt.Sprintf("%s/badge/%s", strings.TrimSuffix(badgeService, "/"), appUUID)
	} else {
		if err := checkLogin(); err != nil {
			return err
		}
		instance, err := getCurrentInstance()
		if err != nil {
			return err
		}
		// Default: assume badge service runs on same domain or use placeholder
		baseURL := strings.TrimSuffix(instance.FQDN, "/")
		baseURL = strings.TrimSuffix(baseURL, "/api/v1")
		badgeURL = fmt.Sprintf("%s/badge/%s", baseURL, appUUID)
	}

	// Generate output based on format
	var output string
	switch badgeFormat {
	case "markdown":
		output = fmt.Sprintf("![Deployment Status](%s)", badgeURL)
	case "html":
		output = fmt.Sprintf(`<img src="%s" alt="Deployment Status">`, badgeURL)
	case "url":
		output = badgeURL
	default:
		return fmt.Errorf("unsupported format: %s (use markdown, html, or url)", badgeFormat)
	}

	ui.Success("Badge generated")
	ui.Spacer()
	ui.Print(output)
	ui.Spacer()
	ui.Dim("Add this to your README to display live deployment status")

	return nil
}

// runBadgeInteractive prompts for app-uuid and generates badge
func runBadgeInteractive() error {
	appUUID, err := ui.Input("Enter application UUID", "")
	if err != nil {
		return err
	}
	if appUUID == "" {
		ui.Dim("Canceled - no UUID provided")
		return nil
	}
	return runBadge(badgeCmd, []string{appUUID})
}
