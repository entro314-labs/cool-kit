package service

import (
	"context"
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/api"
)

// TeamService handles team-related operations
type TeamService struct {
	client *api.Client
}

// Team represents a Coolify team
type Team struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// NewTeamService creates a new team service
func NewTeamService(client *api.Client) *TeamService {
	return &TeamService{client: client}
}

// List retrieves all teams
func (s *TeamService) List(ctx context.Context) ([]Team, error) {
	var teams []Team
	err := s.client.Get("teams", &teams)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	return teams, nil
}

// Get retrieves a team by ID
func (s *TeamService) Get(ctx context.Context, id string) (*Team, error) {
	var team Team
	err := s.client.Get(fmt.Sprintf("teams/%s", id), &team)
	if err != nil {
		return nil, fmt.Errorf("failed to get team %s: %w", id, err)
	}
	return &team, nil
}

// Current retrieves the currently authenticated team
func (s *TeamService) Current(ctx context.Context) (*Team, error) {
	var team Team
	err := s.client.Get("teams/current", &team)
	if err != nil {
		return nil, fmt.Errorf("failed to get current team: %w", err)
	}
	return &team, nil
}

// ListMembers retrieves members of a specific team
func (s *TeamService) ListMembers(ctx context.Context, teamID string) ([]TeamMember, error) {
	var members []TeamMember
	err := s.client.Get(fmt.Sprintf("teams/%s/members", teamID), &members)
	if err != nil {
		return nil, fmt.Errorf("failed to list members for team %s: %w", teamID, err)
	}
	return members, nil
}

// CurrentMembers retrieves members of the currently authenticated team
func (s *TeamService) CurrentMembers(ctx context.Context) ([]TeamMember, error) {
	var members []TeamMember
	err := s.client.Get("teams/current/members", &members)
	if err != nil {
		return nil, fmt.Errorf("failed to list current team members: %w", err)
	}
	return members, nil
}
