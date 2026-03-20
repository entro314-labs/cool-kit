// Package mcp implements the Model Context Protocol server.
package mcp

import "encoding/json"

// JSON-RPC 2.0 Request
type JsonRpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      *json.Number    `json:"id,omitempty"` // Use json.Number to handle both int and float IDs correctly
}

// JSON-RPC 2.0 Response
type JsonRpcResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
	ID      *json.Number  `json:"id,omitempty"`
}

// JSON-RPC 2.0 Error
type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC 2.0 Errors
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Initialize Result
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

type Capabilities struct {
	Tools map[string]any `json:"tools"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool List Result
type ToolListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

type ToolDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required,omitempty"`
}

// Tool Call Params
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Call Tool Result
type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Tool Arguments Structs
type BaseArgs struct {
	UUID string `json:"uuid"`
}

type BaseAppArgs struct {
	AppUUID string `json:"app_uuid"`
}

type GetLogsArgs struct {
	UUID     string `json:"uuid"`
	Lines    int    `json:"lines"` // Changed from float64 to int as we can decode directly
	GrepText string `json:"grep_text,omitempty"`
}

type DeployAppArgs struct {
	UUID  string `json:"uuid"`
	Force bool   `json:"force"`
}

type GetDeploymentArgs struct {
	DeploymentUUID string `json:"deployment_uuid"`
}
