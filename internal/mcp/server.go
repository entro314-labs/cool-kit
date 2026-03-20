// Package mcp implements the Model Context Protocol server.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/entro314-labs/cool-kit/internal/api"
)

// MCPServer represents the MCP (Model Context Protocol) server
// This allows AI assistants (Claude Desktop, Cursor, etc.) to interact with Coolify
type MCPServer struct {
	client *api.Client
}

// NewMCPServer creates a new MCP server with the given Coolify API client
func NewMCPServer(client *api.Client) *MCPServer {
	return &MCPServer{client: client}
}

// Start starts the MCP server, listening on stdin for JSON-RPC 2.0 messages
func (s *MCPServer) Start(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if err := s.handleMessage(ctx, line); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
	return scanner.Err()
}

// handleMessage processes a JSON-RPC 2.0 message
func (s *MCPServer) handleMessage(ctx context.Context, msg string) error {
	var req JsonRpcRequest
	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return s.sendError(nil, ParseError, "Parse error", nil)
	}

	// Validate request
	if req.JSONRPC != "2.0" {
		// Valid JSON-RPC 2.0 but wrong version, still treat as valid request format for error reporting
		return s.sendError(req.ID, InvalidRequest, "Invalid JSON-RPC version", nil)
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.ID)
	case "tools/list":
		return s.handleToolsList(req.ID)
	case "tools/call":
		return s.handleToolsCall(ctx, req.ID, req.Params)
	default:
		return s.sendError(req.ID, MethodNotFound, "Method not found", req.Method)
	}
}

// handleInitialize handles the MCP initialize request
func (s *MCPServer) handleInitialize(id *json.Number) error {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: Capabilities{
			Tools: map[string]any{}, // Empty map as required by spec if supporting tools
		},
		ServerInfo: ServerInfo{
			Name:    "cool-kit-mcp",
			Version: "1.0.0",
		},
	}

	return s.sendResponse(id, result)
}

// handleToolsList returns the list of available MCP tools
func (s *MCPServer) handleToolsList(id *json.Number) error {
	tools := []ToolDefinition{
		{
			Name:        "list_applications",
			Description: "List all Coolify applications with their UUIDs and names",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]any{},
			},
		},
		{
			Name:        "get_application",
			Description: "Get details of a specific Coolify application by UUID",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				Required: []string{"uuid"},
			},
		},
		{
			Name:        "get_application_logs",
			Description: "Retrieve logs for a Coolify application for debugging purposes",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
					"lines": map[string]any{
						"type":        "integer",
						"description": "Number of lines to retrieve (default: 100)",
					},
					"grep_text": map[string]any{
						"type":        "string",
						"description": "Text to filter/grep in logs",
					},
				},
				Required: []string{"uuid"},
			},
		},
		{
			Name:        "start_application",
			Description: "Start a Coolify application",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				Required: []string{"uuid"},
			},
		},
		{
			Name:        "stop_application",
			Description: "Stop a Coolify application",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				Required: []string{"uuid"},
			},
		},
		{
			Name:        "restart_application",
			Description: "Restart a Coolify application",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				Required: []string{"uuid"},
			},
		},
		{
			Name:        "deploy_application",
			Description: "Trigger a deployment for a Coolify application",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
					"force": map[string]any{
						"type":        "boolean",
						"description": "Force rebuild without cache",
					},
				},
				Required: []string{"uuid"},
			},
		},
		{
			Name:        "list_deployments",
			Description: "List deployments for a Coolify application",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"app_uuid": map[string]any{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				Required: []string{"app_uuid"},
			},
		},
		{
			Name:        "get_deployment",
			Description: "Get details and logs of a specific deployment",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]any{
					"deployment_uuid": map[string]any{
						"type":        "string",
						"description": "Deployment UUID",
					},
				},
				Required: []string{"deployment_uuid"},
			},
		},
	}

	result := ToolListResult{
		Tools: tools,
	}
	return s.sendResponse(id, result)
}

// handleToolsCall dispatches tool calls to the appropriate handler
func (s *MCPServer) handleToolsCall(ctx context.Context, id *json.Number, paramsRaw json.RawMessage) error {
	var params ToolCallParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		return s.sendError(id, InvalidParams, "Invalid params definition", nil)
	}

	var result any
	var err error

	switch params.Name {
	case "list_applications":
		result, err = s.listApplications(ctx)
	case "get_application":
		var args BaseArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.getApplication(ctx, args)
	case "get_application_logs":
		var args GetLogsArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.getApplicationLogs(ctx, args)
	case "start_application":
		var args BaseArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.startApplication(ctx, args)
	case "stop_application":
		var args BaseArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.stopApplication(ctx, args)
	case "restart_application":
		var args BaseArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.restartApplication(ctx, args)
	case "deploy_application":
		var args DeployAppArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.deployApplication(ctx, args)
	case "list_deployments":
		var args BaseAppArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.listDeployments(ctx, args)
	case "get_deployment":
		var args GetDeploymentArgs
		if jsonErr := json.Unmarshal(params.Arguments, &args); jsonErr != nil {
			return s.sendError(id, InvalidParams, "Invalid arguments", nil)
		}
		result, err = s.getDeployment(ctx, args)
	default:
		return s.sendError(id, MethodNotFound, "Tool not found", params.Name)
	}

	if err != nil {
		// Return application error as a tool error result rather than JSON-RPC error
		// This allows the model to see the error output
		return s.sendToolResult(id, fmt.Sprintf("Error: %v", err), true)
	}

	// Format result as text content for the tool result
	resultText, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		resultText = fmt.Appendf(nil, "%v", result)
	}

	return s.sendToolResult(id, string(resultText), false)
}

func (s *MCPServer) sendToolResult(id *json.Number, text string, isError bool) error {
	result := CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: isError,
	}
	return s.sendResponse(id, result)
}

// Tool implementations

func (s *MCPServer) listApplications(ctx context.Context) (any, error) {
	apps, err := s.client.ListApplications(ctx)
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for _, app := range apps {
		var fqdn string
		if app.Fqdn != nil {
			fqdn = *app.Fqdn
		}
		result = append(result, map[string]any{
			"uuid":   app.UUID,
			"name":   app.Name,
			"status": app.Status,
			"fqdn":   fqdn,
		})
	}
	return result, nil
}

func (s *MCPServer) getApplication(ctx context.Context, args BaseArgs) (any, error) {
	if args.UUID == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	app, err := s.client.GetApplication(ctx, args.UUID)
	if err != nil {
		return nil, err
	}

	var fqdn string
	if app.Fqdn != nil {
		fqdn = *app.Fqdn
	}

	return map[string]any{
		"uuid":       app.UUID,
		"name":       app.Name,
		"status":     app.Status,
		"fqdn":       fqdn,
		"git_repo":   app.GitRepository,
		"git_branch": app.GitBranch,
	}, nil
}

func (s *MCPServer) getApplicationLogs(ctx context.Context, args GetLogsArgs) (any, error) {
	if args.UUID == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	lines := args.Lines
	if lines <= 0 {
		lines = 100
	}

	logs, err := s.client.GetApplicationLogs(ctx, args.UUID, lines)
	if err != nil {
		return nil, err
	}

	logText := logs.Logs

	// Apply grep filter if specified
	if args.GrepText != "" {
		logLines := strings.Split(logText, "\n")
		var filteredLines []string
		for _, line := range logLines {
			if strings.Contains(line, args.GrepText) {
				filteredLines = append(filteredLines, line)
			}
		}
		logText = strings.Join(filteredLines, "\n")
	}

	return map[string]any{
		"uuid": args.UUID,
		"logs": logText,
	}, nil
}

func (s *MCPServer) startApplication(ctx context.Context, args BaseArgs) (any, error) {
	if args.UUID == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.StartApplication(ctx, args.UUID, false, false)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"message":         resp.Message,
		"deployment_uuid": resp.DeploymentUUID,
	}, nil
}

func (s *MCPServer) stopApplication(ctx context.Context, args BaseArgs) (any, error) {
	if args.UUID == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.StopApplication(ctx, args.UUID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"message": resp.Message,
	}, nil
}

func (s *MCPServer) restartApplication(ctx context.Context, args BaseArgs) (any, error) {
	if args.UUID == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.RestartApplication(ctx, args.UUID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"message":         resp.Message,
		"deployment_uuid": resp.DeploymentUUID,
	}, nil
}

func (s *MCPServer) deployApplication(ctx context.Context, args DeployAppArgs) (any, error) {
	if args.UUID == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.Deploy(ctx, args.UUID, args.Force, 0)
	if err != nil {
		return nil, err
	}

	// DeployResponse contains array of deployments
	var message, deploymentUUID string
	if len(resp.Deployments) > 0 {
		message = resp.Deployments[0].Message
		deploymentUUID = resp.Deployments[0].DeploymentUUID
	}

	return map[string]any{
		"message":         message,
		"deployment_uuid": deploymentUUID,
	}, nil
}

func (s *MCPServer) listDeployments(ctx context.Context, args BaseAppArgs) (any, error) {
	if args.AppUUID == "" {
		return nil, fmt.Errorf("app_uuid is required")
	}

	deployments, err := s.client.ListDeployments(ctx, args.AppUUID)
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for _, d := range deployments {
		result = append(result, map[string]any{
			"uuid":            d.UUID,
			"deployment_uuid": d.DeploymentUUID,
			"status":          d.Status,
			"commit":          d.Commit,
			"commit_message":  d.CommitMessage,
			"created_at":      d.CreatedAt,
		})
	}
	return result, nil
}

func (s *MCPServer) getDeployment(ctx context.Context, args GetDeploymentArgs) (any, error) {
	if args.DeploymentUUID == "" {
		return nil, fmt.Errorf("deployment_uuid is required")
	}

	deployment, err := s.client.GetDeployment(ctx, args.DeploymentUUID)
	if err != nil {
		return nil, err
	}

	// Parse logs for readability
	parsedLogs := api.ParseLogs(deployment.Logs)

	return map[string]any{
		"deployment_uuid":  deployment.DeploymentUUID,
		"status":           deployment.Status,
		"application_name": deployment.ApplicationName,
		"server_name":      deployment.ServerName,
		"commit":           deployment.Commit,
		"commit_message":   deployment.CommitMessage,
		"created_at":       deployment.CreatedAt,
		"logs":             parsedLogs,
	}, nil
}

// sendResponse sends a JSON-RPC 2.0 response to stdout
func (s *MCPServer) sendResponse(id *json.Number, result any) error {
	response := JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// sendError sends a JSON-RPC 2.0 error response
func (s *MCPServer) sendError(id *json.Number, code int, message string, data any) error {
	response := JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JsonRpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	dataBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Println(string(dataBytes))
	return nil
}
