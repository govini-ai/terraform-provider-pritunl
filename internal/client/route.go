package client

import (
	"encoding/json"
	"fmt"
)

// Route represents a route on a Pritunl server
type Route struct {
	ID           string `json:"id,omitempty"`
	Network      string `json:"network"`
	Comment      string `json:"comment,omitempty"`
	Metric       int    `json:"metric,omitempty"`
	Nat          bool   `json:"nat,omitempty"`
	NatInterface string `json:"nat_interface,omitempty"`
	NatNetmap    string `json:"nat_netmap,omitempty"`
	Advertise    bool   `json:"advertise,omitempty"`
	VpcRegion    string `json:"vpc_region,omitempty"`
	VpcID        string `json:"vpc_id,omitempty"`
	NetGateway   bool   `json:"net_gateway,omitempty"`
	Server       string `json:"server,omitempty"`
}

// ListRoutes returns all routes for a server
func (c *Client) ListRoutes(serverID string) ([]Route, error) {
	resp, err := c.Get("/server/" + serverID + "/route")
	if err != nil {
		return nil, err
	}

	var routes []Route
	if err := json.Unmarshal(resp, &routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}

	return routes, nil
}

// GetRoute returns a single route by ID
func (c *Client) GetRoute(serverID, routeID string) (*Route, error) {
	routes, err := c.ListRoutes(serverID)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		if route.ID == routeID {
			return &route, nil
		}
	}

	return nil, fmt.Errorf("route with ID %q not found on server %q", routeID, serverID)
}

// CreateRoute creates a new route on a server
func (c *Client) CreateRoute(serverID string, route *Route) (*Route, error) {
	resp, err := c.Post("/server/"+serverID+"/route", route)
	if err != nil {
		return nil, err
	}

	var created Route
	if err := json.Unmarshal(resp, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created route: %w", err)
	}

	return &created, nil
}

// UpdateRoute updates an existing route
func (c *Client) UpdateRoute(serverID, routeID string, route *Route) (*Route, error) {
	resp, err := c.Put("/server/"+serverID+"/route/"+routeID, route)
	if err != nil {
		return nil, err
	}

	var updated Route
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated route: %w", err)
	}

	return &updated, nil
}

// DeleteRoute deletes a route from a server
func (c *Client) DeleteRoute(serverID, routeID string) error {
	_, err := c.Delete("/server/" + serverID + "/route/" + routeID)
	return err
}
