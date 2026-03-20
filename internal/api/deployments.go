package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Deploy triggers a deployment for an application
// pr parameter: 0 for production, >0 for preview deployment
// Deploy triggers a deployment for an application
// pr parameter: 0 for production, >0 for preview deployment
func (c *Client) Deploy(ctx context.Context, uuid string, force bool, pr int) (*DeployResponse, error) {
	params := map[string]string{
		"uuid": uuid,
	}
	if force {
		params["force"] = "true"
	}
	if pr > 0 {
		params["pr"] = fmt.Sprintf("%d", pr)
	}
	var resp DeployResponse
	err := c.GetWithParams(ctx, "/deploy", params, &resp)
	return &resp, err
}

// DeployByTag triggers a deployment by tag
func (c *Client) DeployByTag(ctx context.Context, tag string, force bool) (*DeployResponse, error) {
	params := map[string]string{
		"tag": tag,
	}
	if force {
		params["force"] = "true"
	}
	var resp DeployResponse
	err := c.GetWithParams(ctx, "/deploy", params, &resp)
	return &resp, err
}

// GetDeploymentLogs returns logs for a deployment
func (c *Client) GetDeploymentLogs(ctx context.Context, appUUID string) (string, error) {
	var resp DeploymentLogsResponse
	err := c.Get(ctx, fmt.Sprintf("/applications/%s/logs", appUUID), &resp)
	return resp.Logs, err
}

// Deployment represents a deployment in Coolify
// Note: Coolify API returns some IDs as strings
type Deployment struct {
	ID               ID     `json:"id"`
	UUID             string `json:"uuid"`
	Status           string `json:"status"`
	ApplicationID    ID     `json:"application_id"`
	ApplicationUUID  string `json:"application_uuid"`
	DeploymentUUID   string `json:"deployment_uuid"`
	PullRequestID    ID     `json:"pull_request_id"`
	ForceRebuild     bool   `json:"force_rebuild"`
	CommitMessage    string `json:"commit_message"`
	Commit           string `json:"commit"`
	GitCommitSha     string `json:"git_commit_sha"`
	GitType          string `json:"git_type"`
	OnlyThisServer   bool   `json:"only_this_server"`
	RollbackToUUID   string `json:"rollback_to"`
	CurrentProcessID string `json:"current_process_id"`
	DestinationID    ID     `json:"destination_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// DeploymentLogsResponse represents the response from the deployment logs endpoint
type DeploymentLogsResponse struct {
	Logs string `json:"logs"`
}

// ListDeployments returns deployments for an application
func (c *Client) ListDeployments(ctx context.Context, appUUID string) ([]Deployment, error) {
	var deployments []Deployment
	err := c.Get(ctx, fmt.Sprintf("/deployments?application_uuid=%s", appUUID), &deployments)
	return deployments, err
}

// DeploymentDetail contains full deployment info including logs
// Note: Coolify API returns some IDs as strings, so we use json.Number/interface{} for flexibility
type DeploymentDetail struct {
	ID              ID     `json:"id"`
	ApplicationID   ID     `json:"application_id"`
	DeploymentUUID  string `json:"deployment_uuid"`
	Status          string `json:"status"`
	Logs            string `json:"logs"`
	Commit          string `json:"commit"`
	CommitMessage   string `json:"commit_message"`
	ServerID        ID     `json:"server_id"`
	ApplicationName string `json:"application_name"`
	ServerName      string `json:"server_name"`
	DeploymentURL   string `json:"deployment_url"`
	DestinationID   ID     `json:"destination_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// GetDeployment returns a specific deployment by UUID with full details
func (c *Client) GetDeployment(ctx context.Context, deploymentUUID string) (*DeploymentDetail, error) {
	var deployment DeploymentDetail
	err := c.Get(ctx, fmt.Sprintf("/deployments/%s", deploymentUUID), &deployment)
	return &deployment, err
}

// LogEntry represents a single log entry from Coolify
type LogEntry struct {
	Output    string `json:"output"`
	Type      string `json:"type"`
	Hidden    bool   `json:"hidden"`
	Timestamp string `json:"timestamp"`
	Order     int    `json:"order"`
}

// ParseLogs extracts readable log output from Coolify's JSON log format
func ParseLogs(rawLogs string) string {
	if rawLogs == "" {
		return ""
	}

	// Use a JSON decoder to handle concatenated JSON objects/arrays
	reader := strings.NewReader(rawLogs)
	decoder := json.NewDecoder(reader)

	var allEntries []LogEntry

	for {
		// Try to decode the next JSON value
		// It could be a single LogEntry, or an array []LogEntry
		// We'll decode into json.RawMessage first to see what we have
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			// If we can't decode as JSON, checking if we can skip or if it's mixed content
			// For robustness, if decode fails, we might just stop or skip a byte
			// strict JSON parsing might fail on partial logs
			break
		}

		// Try to unmarshal as array of LogEntry
		var entries []LogEntry
		if err := json.Unmarshal(raw, &entries); err == nil {
			allEntries = append(allEntries, entries...)
			continue
		}

		// Try to unmarshal as single LogEntry
		var entry LogEntry
		if err := json.Unmarshal(raw, &entry); err == nil {
			allEntries = append(allEntries, entry)
			continue
		}
	}

	if len(allEntries) > 0 {
		return formatLogEntries(allEntries)
	}

	// Fallback mechanism: sometimes logs are just lines of strings if not valid JSON
	// Check if it looks like JSON at all
	trimmed := strings.TrimSpace(rawLogs)
	if !strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "{") {
		return rawLogs
	}

	return rawLogs
}

func formatLogEntries(entries []LogEntry) string {
	var lines []string
	for _, e := range entries {
		// Skip hidden entries and empty output
		if e.Hidden || e.Output == "" {
			continue
		}
		lines = append(lines, e.Output)
	}
	return strings.Join(lines, "\n")
}

// GetBuildLogs returns build logs for a specific deployment
func (c *Client) GetBuildLogs(ctx context.Context, deploymentUUID string) (string, error) {
	deployment, err := c.GetDeployment(ctx, deploymentUUID)
	if err != nil {
		return "", err
	}
	return deployment.Logs, nil
}

// HealthCheck validates the API connection
func (c *Client) HealthCheck(ctx context.Context) error {
	var resp HealthCheckResponse
	// Try to get the version endpoint instead since healthcheck might not be available
	err := c.Get(ctx, "/version", &resp)
	if err != nil {
		// If version fails, try listing teams as a validation
		var teams []Team
		err = c.Get(ctx, "/teams", &teams)
	}
	return err
}
