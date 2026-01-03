package service

import (
	"context"
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/api"
)

// GitHubAppService handles GitHub App-related operations
type GitHubAppService struct {
	client *api.Client
}

// GitHubApp represents a GitHub App integration
type GitHubApp struct {
	ID             int    `json:"id,omitempty"`
	UUID           string `json:"uuid,omitempty"`
	Name           string `json:"name,omitempty"`
	Organization   string `json:"organization,omitempty"`
	AppID          int    `json:"app_id,omitempty"`
	InstallationID int    `json:"installation_id,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	HtmlURL        string `json:"html_url,omitempty"`
	ApiURL         string `json:"api_url,omitempty"`
	CustomPort     int    `json:"custom_port,omitempty"`
	CustomUser     string `json:"custom_user,omitempty"`
	IsSystemWide   bool   `json:"is_system_wide,omitempty"`
	PrivateKeyUUID string `json:"private_key_uuid,omitempty"`
	TeamID         int    `json:"team_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Private  bool   `json:"private,omitempty"`
	HTMLURL  string `json:"html_url,omitempty"`
}

// GitHubBranch represents a GitHub branch
type GitHubBranch struct {
	Name      string `json:"name,omitempty"`
	Protected bool   `json:"protected,omitempty"`
}

// GitHubAppCreateRequest is the request body for creating a GitHub App
type GitHubAppCreateRequest struct {
	Name           string `json:"name"`
	Organization   string `json:"organization,omitempty"`
	AppID          int    `json:"app_id"`
	InstallationID int    `json:"installation_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	WebhookSecret  string `json:"webhook_secret,omitempty"`
	PrivateKeyUUID string `json:"private_key_uuid"`
	HtmlURL        string `json:"html_url"`
	ApiURL         string `json:"api_url"`
	CustomPort     int    `json:"custom_port,omitempty"`
	CustomUser     string `json:"custom_user,omitempty"`
	IsSystemWide   bool   `json:"is_system_wide,omitempty"`
}

// GitHubAppUpdateRequest is the request body for updating a GitHub App
type GitHubAppUpdateRequest struct {
	Name           string `json:"name,omitempty"`
	Organization   string `json:"organization,omitempty"`
	InstallationID int    `json:"installation_id,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty"`
	WebhookSecret  string `json:"webhook_secret,omitempty"`
	CustomPort     int    `json:"custom_port,omitempty"`
	CustomUser     string `json:"custom_user,omitempty"`
}

// NewGitHubAppService creates a new GitHub App service
func NewGitHubAppService(client *api.Client) *GitHubAppService {
	return &GitHubAppService{
		client: client,
	}
}

// List retrieves all GitHub Apps
func (s *GitHubAppService) List(ctx context.Context) ([]GitHubApp, error) {
	var apps []GitHubApp
	err := s.client.Get("github-apps", &apps)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub Apps: %w", err)
	}
	return apps, nil
}

// Get retrieves a specific GitHub App by UUID
func (s *GitHubAppService) Get(ctx context.Context, uuid string) (*GitHubApp, error) {
	var app GitHubApp
	err := s.client.Get(fmt.Sprintf("github-apps/%s", uuid), &app)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub App %s: %w", uuid, err)
	}
	return &app, nil
}

// Create creates a new GitHub App
func (s *GitHubAppService) Create(ctx context.Context, req *GitHubAppCreateRequest) (*GitHubApp, error) {
	var app GitHubApp
	err := s.client.Post("github-apps", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub App: %w", err)
	}
	return &app, nil
}

// Update updates an existing GitHub App
func (s *GitHubAppService) Update(ctx context.Context, uuid string, req *GitHubAppUpdateRequest) error {
	type response struct {
		Message string `json:"message"`
	}
	var resp response
	err := s.client.Patch(fmt.Sprintf("github-apps/%s", uuid), req, &resp)
	if err != nil {
		return fmt.Errorf("failed to update GitHub App %s: %w", uuid, err)
	}
	return nil
}

// Delete deletes a GitHub App
func (s *GitHubAppService) Delete(ctx context.Context, uuid string) error {
	err := s.client.Delete(fmt.Sprintf("github-apps/%s", uuid))
	if err != nil {
		return fmt.Errorf("failed to delete GitHub App %s: %w", uuid, err)
	}
	return nil
}

// ListRepositories lists all repositories accessible by a GitHub App
func (s *GitHubAppService) ListRepositories(ctx context.Context, appUUID string) ([]GitHubRepository, error) {
	type response struct {
		Repositories []GitHubRepository `json:"repositories"`
	}
	var resp response
	err := s.client.Get(fmt.Sprintf("github-apps/%s/repositories", appUUID), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for GitHub App %s: %w", appUUID, err)
	}
	return resp.Repositories, nil
}

// ListBranches lists all branches for a repository
func (s *GitHubAppService) ListBranches(ctx context.Context, appUUID string, owner, repo string) ([]GitHubBranch, error) {
	type response struct {
		Branches []GitHubBranch `json:"branches"`
	}
	var resp response
	err := s.client.Get(fmt.Sprintf("github-apps/%s/repositories/%s/%s/branches", appUUID, owner, repo), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for %s/%s: %w", owner, repo, err)
	}
	return resp.Branches, nil
}
