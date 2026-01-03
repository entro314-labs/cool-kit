package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/entro314-labs/cool-kit/internal/api"
	"github.com/entro314-labs/cool-kit/internal/config"
	"github.com/entro314-labs/cool-kit/internal/mcp"
	"github.com/entro314-labs/cool-kit/internal/ui"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP (Model Context Protocol) server",
	Long: `Start the MCP server for AI assistant integration.

The MCP server allows AI assistants like Claude Desktop, Cursor, and other
MCP-compatible tools to interact with your Coolify deployments.

The server communicates via stdio using JSON-RPC 2.0 protocol.

To use with Claude Desktop, add to your config:
{
  "mcpServers": {
    "coolify": {
      "command": "cool-kit",
      "args": ["mcp"]
    }
  }
}

Available tools exposed via MCP:
  • list_applications     - List all Coolify applications
  • get_application       - Get details of a specific application
  • get_application_logs  - Retrieve application logs
  • start_application     - Start an application
  • stop_application      - Stop an application  
  • restart_application   - Restart an application
  • deploy_application    - Trigger a deployment
  • list_deployments      - List deployments for an application
  • get_deployment        - Get deployment details with logs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCP()
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP() error {
	// Load global config to get Coolify connection
	globalCfg, err := config.LoadGlobal()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Not logged in. Run 'cool-kit login' first.")
		return err
	}

	// Check required fields
	if globalCfg.CoolifyURL == "" || globalCfg.CoolifyToken == "" {
		fmt.Fprintln(os.Stderr, "Error: Coolify URL or token not configured. Run 'cool-kit login' first.")
		return fmt.Errorf("missing URL or token")
	}

	// Create API client
	client := api.NewClient(globalCfg.CoolifyURL, globalCfg.CoolifyToken)

	// Verify connection
	if err := client.HealthCheck(); err != nil {
		ui.Error(fmt.Sprintf("Failed to connect to Coolify: %v", err))
		return err
	}

	// Start MCP server
	server := mcp.NewMCPServer(client)
	return server.Start()
}
