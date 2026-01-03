package api

import "fmt"

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
func (c *Client) ListDatabases() ([]Database, error) {
	var databases []Database
	err := c.Get("/databases", &databases)
	return databases, err
}

// GetDatabase returns a database by UUID
func (c *Client) GetDatabase(uuid string) (*Database, error) {
	var database Database
	err := c.Get("/databases/"+uuid, &database)
	return &database, err
}

// CreateDatabase creates a new database
func (c *Client) CreateDatabase(config map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post("/databases", config, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateDatabase updates a database
func (c *Client) UpdateDatabase(uuid string, updates map[string]interface{}) error {
	return c.Patch("/databases/"+uuid, updates, nil)
}

// DeleteDatabase deletes a database
func (c *Client) DeleteDatabase(uuid string) error {
	return c.Delete("/databases/" + uuid)
}

// StartDatabase starts a database
func (c *Client) StartDatabase(uuid string) error {
	return c.Post(fmt.Sprintf("/databases/%s/start", uuid), nil, nil)
}

// StopDatabase stops a database
func (c *Client) StopDatabase(uuid string) error {
	return c.Post(fmt.Sprintf("/databases/%s/stop", uuid), nil, nil)
}

// RestartDatabase restarts a database
func (c *Client) RestartDatabase(uuid string) error {
	return c.Post(fmt.Sprintf("/databases/%s/restart", uuid), nil, nil)
}

// ListServices returns all services
func (c *Client) ListServices() ([]Service, error) {
	var services []Service
	err := c.Get("/services", &services)
	return services, err
}

// GetService returns a service by UUID
func (c *Client) GetService(uuid string) (*Service, error) {
	var service Service
	err := c.Get("/services/"+uuid, &service)
	return &service, err
}

// CreateService creates a new service
func (c *Client) CreateService(config map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Post("/services", config, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateService updates a service
func (c *Client) UpdateService(uuid string, updates map[string]interface{}) error {
	return c.Patch("/services/"+uuid, updates, nil)
}

// DeleteService deletes a service
func (c *Client) DeleteService(uuid string) error {
	return c.Delete("/services/" + uuid)
}

// StartService starts a service
func (c *Client) StartService(uuid string) error {
	return c.Post(fmt.Sprintf("/services/%s/start", uuid), nil, nil)
}

// StopService stops a service
func (c *Client) StopService(uuid string) error {
	return c.Post(fmt.Sprintf("/services/%s/stop", uuid), nil, nil)
}

// RestartService restarts a service
func (c *Client) RestartService(uuid string) error {
	return c.Post(fmt.Sprintf("/services/%s/restart", uuid), nil, nil)
}
