package api

import (
	"context"
	"fmt"
)

// ListProjects returns all projects
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var projects []Project
	err := c.Get(ctx, "/projects", &projects)
	return projects, err
}

// GetProject returns a project by UUID
func (c *Client) GetProject(ctx context.Context, uuid string) (*Project, error) {
	var project Project
	err := c.Get(ctx, "/projects/"+uuid, &project)
	return &project, err
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	body := map[string]string{
		"name":        name,
		"description": description,
	}
	var project Project
	err := c.Post(ctx, "/projects", body, &project)
	return &project, err
}

// CreateEnvironment creates a new environment in a project
func (c *Client) CreateEnvironment(ctx context.Context, projectUUID, name string) (*Environment, error) {
	body := map[string]string{
		"name": name,
	}
	var env Environment
	err := c.Post(ctx, fmt.Sprintf("/projects/%s/environments", projectUUID), body, &env)
	return &env, err
}

// DeleteProject deletes a project by UUID
func (c *Client) DeleteProject(ctx context.Context, uuid string) error {
	return c.Delete(ctx, "/projects/"+uuid)
}
