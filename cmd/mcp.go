package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

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
		return runMCP(cmd.Context())
	},
}

func init() {
}

func runMCP(ctx context.Context) error {
	client, instance, err := getCurrentCoolifyClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Not logged in. Run 'cool-kit login' first.")
		return err
	}

	// Verify connection
	if err := client.HealthCheck(ctx); err != nil {
		ui.Error(fmt.Sprintf("Failed to connect to Coolify instance %q: %v", instance.Name, err))
		return err
	}

	// Start MCP server
	server := mcp.NewMCPServer(client)
	return server.Start(ctx)
}
