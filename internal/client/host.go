package client

import (
	"encoding/json"
	"fmt"
)

// Host represents a Pritunl host
type Host struct {
	ID              string  `json:"id,omitempty"`
	Name            string  `json:"name,omitempty"`
	Hostname        string  `json:"hostname,omitempty"`
	Status          string  `json:"status,omitempty"`
	Uptime          int64   `json:"uptime,omitempty"`
	UsersOnline     int     `json:"users_online,omitempty"`
	PublicAddr      string  `json:"public_addr,omitempty"`
	PublicAddr6     string  `json:"public_addr6,omitempty"`
	RoutedSubnet6   string  `json:"routed_subnet6,omitempty"`
	RoutedSubnet6WG string  `json:"routed_subnet6_wg,omitempty"`
	LocalAddr       string  `json:"local_addr,omitempty"`
	LocalAddr6      string  `json:"local_addr6,omitempty"`
	LinkAddr        string  `json:"link_addr,omitempty"`
	SyncAddress     string  `json:"sync_address,omitempty"`
	Availability    string  `json:"availability,omitempty"`
	AvailabilityG   string  `json:"availability_group,omitempty"`
	CPUUsage        float64 `json:"cpu_usage,omitempty"`
	MemUsage        float64 `json:"mem_usage,omitempty"`
	Version         string  `json:"version,omitempty"`
}

// ListHosts returns all hosts
func (c *Client) ListHosts() ([]Host, error) {
	resp, err := c.Get("/host")
	if err != nil {
		return nil, err
	}

	var hosts []Host
	if err := json.Unmarshal(resp, &hosts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hosts: %w", err)
	}

	return hosts, nil
}

// GetHost returns a single host by ID
func (c *Client) GetHost(id string) (*Host, error) {
	resp, err := c.Get("/host/" + id)
	if err != nil {
		return nil, err
	}

	var host Host
	if err := json.Unmarshal(resp, &host); err != nil {
		return nil, fmt.Errorf("failed to unmarshal host: %w", err)
	}

	return &host, nil
}

// GetHostByName returns a single host by name
func (c *Client) GetHostByName(name string) (*Host, error) {
	hosts, err := c.ListHosts()
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.Name == name {
			return &host, nil
		}
	}

	return nil, fmt.Errorf("host with name %q not found", name)
}
