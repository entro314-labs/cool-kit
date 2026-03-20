package api

import "context"

// ListServers returns all servers
func (c *Client) ListServers(ctx context.Context) ([]Server, error) {
	var servers []Server
	err := c.Get(ctx, "/servers", &servers)
	return servers, err
}

// GetServer returns a server by UUID
func (c *Client) GetServer(ctx context.Context, uuid string) (*Server, error) {
	var server Server
	err := c.Get(ctx, "/servers/"+uuid, &server)
	return &server, err
}
