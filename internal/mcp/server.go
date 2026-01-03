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
func (s *MCPServer) Start() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if err := s.handleMessage(line); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
	return scanner.Err()
}

// handleMessage processes a JSON-RPC 2.0 message
func (s *MCPServer) handleMessage(msg string) error {
	var req map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &req); err != nil {
		return s.sendError(-32700, "Parse error", nil)
	}

	id, ok := req["id"].(float64)
	if !ok {
		// id is optional or might be missing, but for JSON-RPC usually required for request/response matching
		// We'll proceed with 0 if missing/invalid, assuming notification or we can't reply properly anyway
	}
	method, ok := req["method"].(string)
	if !ok {
		return s.sendError(int(id), "Invalid Request", nil)
	}

	switch method {
	case "initialize":
		return s.handleInitialize(id)
	case "tools/list":
		return s.handleToolsList(id)
	case "tools/call":
		return s.handleToolsCall(id, req)
	default:
		return s.sendError(int(id), "Method not found", method)
	}
}

// handleInitialize handles the MCP initialize request
func (s *MCPServer) handleInitialize(id float64) error {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "cool-kit-mcp",
				"version": "1.0.0",
			},
		},
	}
	return s.sendResponse(response)
}

// handleToolsList returns the list of available MCP tools
func (s *MCPServer) handleToolsList(id float64) error {
	tools := []map[string]interface{}{
		{
			"name":        "list_applications",
			"description": "List all Coolify applications with their UUIDs and names",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "get_application",
			"description": "Get details of a specific Coolify application by UUID",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				"required": []string{"uuid"},
			},
		},
		{
			"name":        "get_application_logs",
			"description": "Retrieve logs for a Coolify application for debugging purposes",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
					"lines": map[string]interface{}{
						"type":        "integer",
						"description": "Number of lines to retrieve (default: 100)",
					},
					"grep_text": map[string]interface{}{
						"type":        "string",
						"description": "Text to filter/grep in logs",
					},
				},
				"required": []string{"uuid"},
			},
		},
		{
			"name":        "start_application",
			"description": "Start a Coolify application",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				"required": []string{"uuid"},
			},
		},
		{
			"name":        "stop_application",
			"description": "Stop a Coolify application",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				"required": []string{"uuid"},
			},
		},
		{
			"name":        "restart_application",
			"description": "Restart a Coolify application",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				"required": []string{"uuid"},
			},
		},
		{
			"name":        "deploy_application",
			"description": "Trigger a deployment for a Coolify application",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force rebuild without cache",
					},
				},
				"required": []string{"uuid"},
			},
		},
		{
			"name":        "list_deployments",
			"description": "List deployments for a Coolify application",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"app_uuid": map[string]interface{}{
						"type":        "string",
						"description": "Application UUID",
					},
				},
				"required": []string{"app_uuid"},
			},
		},
		{
			"name":        "get_deployment",
			"description": "Get details and logs of a specific deployment",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"deployment_uuid": map[string]interface{}{
						"type":        "string",
						"description": "Deployment UUID",
					},
				},
				"required": []string{"deployment_uuid"},
			},
		},
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": tools,
		},
	}
	return s.sendResponse(response)
}

// handleToolsCall dispatches tool calls to the appropriate handler
func (s *MCPServer) handleToolsCall(id float64, req map[string]interface{}) error {
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		// params are optional
		params = make(map[string]interface{})
	}
	name, ok := params["name"].(string)
	if !ok {
		return s.sendError(int(id), "Invalid tool name", nil)
	}
	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	var result interface{}
	var err error

	switch name {
	case "list_applications":
		result, err = s.listApplications()
	case "get_application":
		result, err = s.getApplication(args)
	case "get_application_logs":
		result, err = s.getApplicationLogs(args)
	case "start_application":
		result, err = s.startApplication(args)
	case "stop_application":
		result, err = s.stopApplication(args)
	case "restart_application":
		result, err = s.restartApplication(args)
	case "deploy_application":
		result, err = s.deployApplication(args)
	case "list_deployments":
		result, err = s.listDeployments(args)
	case "get_deployment":
		result, err = s.getDeployment(args)
	default:
		return s.sendError(int(id), "Tool not found", name)
	}

	if err != nil {
		return s.sendError(int(id), err.Error(), nil)
	}

	// Format result as text content
	resultText, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		resultText = []byte(fmt.Sprintf("%v", result))
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": string(resultText),
				},
			},
		},
	}
	return s.sendResponse(response)
}

// Tool implementations

func (s *MCPServer) listApplications() (interface{}, error) {
	apps, err := s.client.ListApplications()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, app := range apps {
		var fqdn string
		if app.Fqdn != nil {
			fqdn = *app.Fqdn
		}
		result = append(result, map[string]interface{}{
			"uuid":   app.UUID,
			"name":   app.Name,
			"status": app.Status,
			"fqdn":   fqdn,
		})
	}
	return result, nil
}

func (s *MCPServer) getApplication(args map[string]interface{}) (interface{}, error) {
	uuid, ok := args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid is required")
	}
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	app, err := s.client.GetApplication(uuid)
	if err != nil {
		return nil, err
	}

	var fqdn string
	if app.Fqdn != nil {
		fqdn = *app.Fqdn
	}

	return map[string]interface{}{
		"uuid":       app.UUID,
		"name":       app.Name,
		"status":     app.Status,
		"fqdn":       fqdn,
		"git_repo":   app.GitRepository,
		"git_branch": app.GitBranch,
	}, nil
}

func (s *MCPServer) getApplicationLogs(args map[string]interface{}) (interface{}, error) {
	uuid, ok := args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid is required")
	}
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	lines := 100
	if l, ok := args["lines"].(float64); ok {
		lines = int(l)
	}

	logs, err := s.client.GetApplicationLogs(context.Background(), uuid, lines)
	if err != nil {
		return nil, err
	}

	logText := logs.Logs

	// Apply grep filter if specified
	if grepText, ok := args["grep_text"].(string); ok && grepText != "" {
		logLines := strings.Split(logText, "\n")
		var filteredLines []string
		for _, line := range logLines {
			if strings.Contains(line, grepText) {
				filteredLines = append(filteredLines, line)
			}
		}
		logText = strings.Join(filteredLines, "\n")
	}

	return map[string]interface{}{
		"uuid": uuid,
		"logs": logText,
	}, nil
}

func (s *MCPServer) startApplication(args map[string]interface{}) (interface{}, error) {
	uuid, ok := args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid is required")
	}
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.StartApplication(context.Background(), uuid, false, false)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":         resp.Message,
		"deployment_uuid": resp.DeploymentUUID,
	}, nil
}

func (s *MCPServer) stopApplication(args map[string]interface{}) (interface{}, error) {
	uuid, ok := args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid is required")
	}
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.StopApplication(context.Background(), uuid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": resp.Message,
	}, nil
}

func (s *MCPServer) restartApplication(args map[string]interface{}) (interface{}, error) {
	uuid, ok := args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid is required")
	}
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	resp, err := s.client.RestartApplication(context.Background(), uuid)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":         resp.Message,
		"deployment_uuid": resp.DeploymentUUID,
	}, nil
}

func (s *MCPServer) deployApplication(args map[string]interface{}) (interface{}, error) {
	uuid, ok := args["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid is required")
	}
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	force := false
	if f, ok := args["force"].(bool); ok {
		force = f
	}

	resp, err := s.client.Deploy(uuid, force, 0)
	if err != nil {
		return nil, err
	}

	// DeployResponse contains array of deployments
	var message, deploymentUUID string
	if len(resp.Deployments) > 0 {
		message = resp.Deployments[0].Message
		deploymentUUID = resp.Deployments[0].DeploymentUUID
	}

	return map[string]interface{}{
		"message":         message,
		"deployment_uuid": deploymentUUID,
	}, nil
}

func (s *MCPServer) listDeployments(args map[string]interface{}) (interface{}, error) {
	appUUID, ok := args["app_uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("app_uuid must be a string")
	}
	if appUUID == "" {
		return nil, fmt.Errorf("app_uuid is required")
	}

	deployments, err := s.client.ListDeployments(appUUID)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, d := range deployments {
		result = append(result, map[string]interface{}{
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

func (s *MCPServer) getDeployment(args map[string]interface{}) (interface{}, error) {
	deploymentUUID, ok := args["deployment_uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("deployment_uuid must be a string")
	}
	if deploymentUUID == "" {
		return nil, fmt.Errorf("deployment_uuid is required")
	}

	deployment, err := s.client.GetDeployment(deploymentUUID)
	if err != nil {
		return nil, err
	}

	// Parse logs for readability
	parsedLogs := api.ParseLogs(deployment.Logs)

	return map[string]interface{}{
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
func (s *MCPServer) sendResponse(response map[string]interface{}) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// sendError sends a JSON-RPC 2.0 error response
func (s *MCPServer) sendError(id int, message string, data interface{}) error {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    -32601,
			"message": message,
		},
	}
	if data != nil {
		response["error"].(map[string]interface{})["data"] = data
	}
	return s.sendResponse(response)
}
