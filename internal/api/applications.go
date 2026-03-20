package api

import (
	"context"
	"fmt"
	"net/http"
)

// ListApplications returns all applications
func (c *Client) ListApplications(ctx context.Context) ([]Application, error) {
	var applications []Application
	err := c.Get(ctx, "/applications", &applications)
	return applications, err
}

// GetApplication returns an application by UUID
func (c *Client) GetApplication(ctx context.Context, uuid string) (*Application, error) {
	var app Application
	err := c.Get(ctx, "/applications/"+uuid, &app)
	return &app, err
}

// CreatePublicApp creates an application from a public git repository
func (c *Client) CreatePublicApp(ctx context.Context, req *CreatePublicAppRequest) (*CreateAppResponse, error) {
	var resp CreateAppResponse
	err := c.Post(ctx, "/applications/public", req, &resp)
	return &resp, err
}

// CreatePublicApplication creates a new application based on a public git repository (CAGC pattern)
func (c *Client) CreatePublicApplication(ctx context.Context, app Application) (*CreateResponse, error) {
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, "/applications/public", app, &response)
	return &response, err
}

// CreateDockerImageApp creates an application from a Docker registry image
func (c *Client) CreateDockerImageApp(ctx context.Context, req *CreateDockerImageAppRequest) (*CreateAppResponse, error) {
	var resp CreateAppResponse
	err := c.Post(ctx, "/applications/dockerimage", req, &resp)
	return &resp, err
}

// CreateDockerImageApplication creates a new application based on a Docker image (CAGC pattern)
func (c *Client) CreateDockerImageApplication(ctx context.Context, app Application) (*CreateResponse, error) {
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, "/applications/dockerimage", app, &response)
	return &response, err
}

// UpdateApplication updates an application
func (c *Client) UpdateApplication(ctx context.Context, uuid string, updates UpdateApplicationRequest) error {
	return c.Patch(ctx, "/applications/"+uuid, updates, nil)
}

// DeleteApplication deletes an application
func (c *Client) DeleteApplication(ctx context.Context, uuid string) error {
	return c.Delete(ctx, "/applications/"+uuid)
}

// StartApplication starts an application (CAGC pattern)
func (c *Client) StartApplication(ctx context.Context, uuid string, force, instantDeploy bool) (*DeploymentResponse, error) {
	path := fmt.Sprintf("/applications/%s/start?force=%t&instant_deploy=%t", uuid, force, instantDeploy)
	var response DeploymentResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

// StopApplication stops an application (CAGC pattern)
func (c *Client) StopApplication(ctx context.Context, uuid string) (*CreateResponse, error) {
	path := fmt.Sprintf("/applications/%s/stop", uuid)
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

// RestartApplication restarts an application (CAGC pattern)
func (c *Client) RestartApplication(ctx context.Context, uuid string) (*DeploymentResponse, error) {
	path := fmt.Sprintf("/applications/%s/restart", uuid)
	var response DeploymentResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

// GetApplicationLogs gets logs for an application (CAGC pattern)
func (c *Client) GetApplicationLogs(ctx context.Context, uuid string, lines int) (*LogsResponse, error) {
	path := fmt.Sprintf("/applications/%s/logs", uuid)
	if lines > 0 {
		path = fmt.Sprintf("%s?lines=%d", path, lines)
	}
	var response LogsResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

// ExecuteCommand executes a command on an application's container (CAGC pattern)
func (c *Client) ExecuteCommand(ctx context.Context, uuid string, command string) (*CommandResponse, error) {
	path := fmt.Sprintf("/applications/%s/execute", uuid)
	req := map[string]string{"command": command}
	var response CommandResponse
	err := c.doRequest(ctx, http.MethodPost, path, req, &response)
	return &response, err
}

// GetApplicationEnvVars returns environment variables for an application
func (c *Client) GetApplicationEnvVars(ctx context.Context, uuid string) ([]EnvVar, error) {
	var envVars []EnvVar
	err := c.Get(ctx, fmt.Sprintf("/applications/%s/envs", uuid), &envVars)
	return envVars, err
}

// ListApplicationEnvs lists all environment variables for an application (CAGC pattern)
func (c *Client) ListApplicationEnvs(ctx context.Context, uuid string) ([]EnvironmentVariable, error) {
	path := fmt.Sprintf("/applications/%s/envs", uuid)
	var envs []EnvironmentVariable
	err := c.doRequest(ctx, http.MethodGet, path, nil, &envs)
	return envs, err
}

// CreateApplicationEnvVar creates an environment variable for an application
func (c *Client) CreateApplicationEnvVar(ctx context.Context, uuid, key, value string, isBuildTime, isPreview bool) (*EnvVar, error) {
	body := EnvironmentVariable{
		Key:         key,
		Value:       value,
		IsPreview:   isPreview,
		IsBuildTime: isBuildTime,
	}
	// Note: API returns EnvVar, but input is EnvironmentVariable style.
	var envVar EnvVar
	err := c.Post(ctx, fmt.Sprintf("/applications/%s/envs", uuid), body, &envVar)
	return &envVar, err
}

// CreateApplicationEnv creates a new environment variable for an application (CAGC pattern)
func (c *Client) CreateApplicationEnv(ctx context.Context, appUUID string, env EnvironmentVariable) (*CreateResponse, error) {
	path := fmt.Sprintf("/applications/%s/envs", appUUID)
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, path, env, &response)
	return &response, err
}

// UpdateApplicationEnv updates an environment variable for an application (CAGC pattern)
func (c *Client) UpdateApplicationEnv(ctx context.Context, appUUID string, env EnvironmentVariable) (*CreateResponse, error) {
	path := fmt.Sprintf("/applications/%s/envs", appUUID)
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPatch, path, env, &response)
	return &response, err
}

// DeleteApplicationEnvVar deletes an environment variable
func (c *Client) DeleteApplicationEnvVar(ctx context.Context, appUUID, envUUID string) error {
	return c.Delete(ctx, fmt.Sprintf("/applications/%s/envs/%s", appUUID, envUUID))
}

// DeleteApplicationEnv deletes an environment variable for an application (CAGC pattern)
func (c *Client) DeleteApplicationEnv(ctx context.Context, appUUID string, envUUID string) (*CreateResponse, error) {
	path := fmt.Sprintf("/applications/%s/envs/%s", appUUID, envUUID)
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodDelete, path, nil, &response)
	return &response, err
}

// UpdateApplicationEnvsBulk updates multiple environment variables for an application (CAGC pattern)
// This is the KEY feature we need for automatic service provisioning!
func (c *Client) UpdateApplicationEnvsBulk(ctx context.Context, appUUID string, envs []EnvironmentVariable) (*CreateResponse, error) {
	path := fmt.Sprintf("/applications/%s/envs/bulk", appUUID)
	req := map[string][]EnvironmentVariable{"data": envs}
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPatch, path, req, &response)
	return &response, err
}

// ListGitHubApps returns all GitHub Apps configured in Coolify
func (c *Client) ListGitHubApps(ctx context.Context) ([]GitHubApp, error) {
	var apps []GitHubApp
	err := c.Get(ctx, "/github-apps", &apps)
	return apps, err
}

// CreatePrivateGitHubApp creates an application from a private GitHub repository using a GitHub App
func (c *Client) CreatePrivateGitHubApp(ctx context.Context, req *CreatePrivateGitHubAppRequest) (*CreateAppResponse, error) {
	var resp CreateAppResponse
	err := c.Post(ctx, "/applications/private-github-app", req, &resp)
	return &resp, err
}

// CreatePrivateGithubAppApplication creates a new application based on a private repo through Github App (CAGC pattern)
func (c *Client) CreatePrivateGithubAppApplication(ctx context.Context, app Application) (*CreateResponse, error) {
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, "/applications/private-github-app", app, &response)
	return &response, err
}

// CreatePrivateDeployKeyApplication creates a new application based on a private repo through Deploy Key (CAGC pattern)
func (c *Client) CreatePrivateDeployKeyApplication(ctx context.Context, app Application) (*CreateResponse, error) {
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, "/applications/private-deploy-key", app, &response)
	return &response, err
}

// CreateDockerfileApplication creates a new application based on a Dockerfile (CAGC pattern)
func (c *Client) CreateDockerfileApplication(ctx context.Context, app Application) (*CreateResponse, error) {
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, "/applications/dockerfile", app, &response)
	return &response, err
}

// CreateDockerComposeApplication creates a new application based on a docker-compose file (CAGC pattern)
func (c *Client) CreateDockerComposeApplication(ctx context.Context, app Application) (*CreateResponse, error) {
	var response CreateResponse
	err := c.doRequest(ctx, http.MethodPost, "/applications/dockercompose", app, &response)
	return &response, err
}
