package service

import (
	"context"
	"fmt"

	"github.com/entro314-labs/cool-kit/internal/api"
)

// PrivateKeyService handles private key-related operations
type PrivateKeyService struct {
	client *api.Client
}

// PrivateKey represents a Coolify private key
type PrivateKey struct {
	ID          int    `json:"id,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	PrivateKey  string `json:"private_key,omitempty"`
	IsGitKey    bool   `json:"is_git_key,omitempty"`
	TeamID      int    `json:"team_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// PrivateKeyCreateRequest is the request body for creating a private key
type PrivateKeyCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PrivateKey  string `json:"private_key"`
}

// NewPrivateKeyService creates a new private key service
func NewPrivateKeyService(client *api.Client) *PrivateKeyService {
	return &PrivateKeyService{
		client: client,
	}
}

// List retrieves all private keys
func (s *PrivateKeyService) List(ctx context.Context) ([]PrivateKey, error) {
	var keys []PrivateKey
	err := s.client.Get("security/keys", &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to list private keys: %w", err)
	}
	return keys, nil
}

// Get retrieves a specific private key by UUID
func (s *PrivateKeyService) Get(ctx context.Context, uuid string) (*PrivateKey, error) {
	var key PrivateKey
	err := s.client.Get(fmt.Sprintf("security/keys/%s", uuid), &key)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key %s: %w", uuid, err)
	}
	return &key, nil
}

// Create creates a new private key
func (s *PrivateKeyService) Create(ctx context.Context, req PrivateKeyCreateRequest) (*PrivateKey, error) {
	var key PrivateKey
	err := s.client.Post("security/keys", req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}
	return &key, nil
}

// Delete deletes a private key by UUID
func (s *PrivateKeyService) Delete(ctx context.Context, uuid string) error {
	err := s.client.Delete(fmt.Sprintf("security/keys/%s", uuid))
	if err != nil {
		return fmt.Errorf("failed to delete private key %s: %w", uuid, err)
	}
	return nil
}
