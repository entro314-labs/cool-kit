package api

import (
	"context"
	"fmt"
)

// Database represents a Coolify database
type Database struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Image       string `json:"image"`
	IsPublic    bool   `json:"is_public"`
}

// Service represents a Coolify service
type Service struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Image       string `json:"image"`
	IsPublic    bool   `json:"is_public"`
}

// CreateDatabaseRequest is the request body for creating a database
type CreateDatabaseRequest struct {
	ProjectUUID     string `json:"project_uuid"`
	ServerUUID      string `json:"server_uuid"`
	EnvironmentName string `json:"environment_name,omitempty"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type"`
	Image           string `json:"image,omitempty"`
	IsPublic        bool   `json:"is_public,omitempty"`
}

// CreateServiceRequest is the request body for creating a service
type CreateServiceRequest struct {
	ProjectUUID     string `json:"project_uuid"`
	ServerUUID      string `json:"server_uuid"`
	EnvironmentName string `json:"environment_name,omitempty"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Type            string `json:"type"`
	Image           string `json:"image,omitempty"`
	IsPublic        bool   `json:"is_public,omitempty"`
}

// CreateDatabaseResponse is the response from creating a database
type CreateDatabaseResponse struct {
	UUID string `json:"uuid"`
}

// CreateServiceResponse is the response from creating a service
type CreateServiceResponse struct {
	UUID string `json:"uuid"`
}

// ListDatabases returns all databases
func (c *Client) ListDatabases(ctx context.Context) ([]Database, error) {
	var databases []Database
	err := c.Get(ctx, "/databases", &databases)
	return databases, err
}

// GetDatabase returns a database by UUID
func (c *Client) GetDatabase(ctx context.Context, uuid string) (*Database, error) {
	var database Database
	err := c.Get(ctx, "/databases/"+uuid, &database)
	return &database, err
}

// CreateDatabase creates a new database
func CreateDatabase[T any](c *Client, ctx context.Context, config T) (*CreateDatabaseResponse, error) {
	var result CreateDatabaseResponse
	err := c.Post(ctx, "/databases", config, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateDatabase wrapper for method-style call with type inference
func (c *Client) CreateDatabase(ctx context.Context, config any) (*CreateDatabaseResponse, error) {
	return CreateDatabase(c, ctx, config)
}

// UpdateDatabase updates a database
func (c *Client) UpdateDatabase(ctx context.Context, uuid string, updates UpdateDatabaseRequest) error {
	return c.Patch(ctx, "/databases/"+uuid, updates, nil)
}

// DeleteDatabase deletes a database
func (c *Client) DeleteDatabase(ctx context.Context, uuid string) error {
	return c.Delete(ctx, "/databases/"+uuid)
}

// StartDatabase starts a database
func (c *Client) StartDatabase(ctx context.Context, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/databases/%s/start", uuid), nil, nil)
}

// StopDatabase stops a database
func (c *Client) StopDatabase(ctx context.Context, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/databases/%s/stop", uuid), nil, nil)
}

// RestartDatabase restarts a database
func (c *Client) RestartDatabase(ctx context.Context, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/databases/%s/restart", uuid), nil, nil)
}

// ListServices returns all services
func (c *Client) ListServices(ctx context.Context) ([]Service, error) {
	var services []Service
	err := c.Get(ctx, "/services", &services)
	return services, err
}

// GetService returns a service by UUID
func (c *Client) GetService(ctx context.Context, uuid string) (*Service, error) {
	var service Service
	err := c.Get(ctx, "/services/"+uuid, &service)
	return &service, err
}

// CreateService creates a new service
func CreateService[T any](c *Client, ctx context.Context, config T) (*CreateServiceResponse, error) {
	var result CreateServiceResponse
	err := c.Post(ctx, "/services", config, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateService wrapper
func (c *Client) CreateService(ctx context.Context, config any) (*CreateServiceResponse, error) {
	return CreateService(c, ctx, config)
}

// UpdateService updates a service
func (c *Client) UpdateService(ctx context.Context, uuid string, updates UpdateServiceRequest) error {
	return c.Patch(ctx, "/services/"+uuid, updates, nil)
}

// DeleteService deletes a service
func (c *Client) DeleteService(ctx context.Context, uuid string) error {
	return c.Delete(ctx, "/services/"+uuid)
}

// StartService starts a service
func (c *Client) StartService(ctx context.Context, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/services/%s/start", uuid), nil, nil)
}

// StopService stops a service
func (c *Client) StopService(ctx context.Context, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/services/%s/stop", uuid), nil, nil)
}

// RestartService restarts a service
func (c *Client) RestartService(ctx context.Context, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/services/%s/restart", uuid), nil, nil)
}
