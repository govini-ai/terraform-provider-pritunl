package client

import (
	"encoding/json"
	"fmt"
)

// Server represents a Pritunl VPN server
type Server struct {
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name"`
	Status           string   `json:"status,omitempty"`
	Uptime           int64    `json:"uptime,omitempty"`
	UsersOnline      int      `json:"users_online,omitempty"`
	DevicesOnline    int      `json:"devices_online,omitempty"`
	UserCount        int      `json:"user_count,omitempty"`
	Network          string   `json:"network"`
	NetworkWG        string   `json:"network_wg,omitempty"`
	NetworkMode      string   `json:"network_mode,omitempty"`
	NetworkStart     string   `json:"network_start,omitempty"`
	NetworkEnd       string   `json:"network_end,omitempty"`
	RestrictRoutes   bool     `json:"restrict_routes,omitempty"`
	IPv6             bool     `json:"ipv6,omitempty"`
	IPv6Firewall     bool     `json:"ipv6_firewall,omitempty"`
	BindAddress      string   `json:"bind_address,omitempty"`
	Protocol         string   `json:"protocol"`
	Port             int      `json:"port"`
	PortWG           int      `json:"port_wg,omitempty"`
	DNSServers       []string `json:"dns_servers,omitempty"`
	SearchDomain     string   `json:"search_domain,omitempty"`
	InterClient      bool     `json:"inter_client"`
	PingInterval     int      `json:"ping_interval,omitempty"`
	PingTimeout      int      `json:"ping_timeout,omitempty"`
	LinkPingInterval int      `json:"link_ping_interval,omitempty"`
	LinkPingTimeout  int      `json:"link_ping_timeout,omitempty"`
	InactiveTimeout  int      `json:"inactive_timeout,omitempty"`
	SessionTimeout   int      `json:"session_timeout,omitempty"`
	AllowedDevices   string   `json:"allowed_devices,omitempty"`
	MaxClients       int      `json:"max_clients,omitempty"`
	MaxDevices       int      `json:"max_devices,omitempty"`
	ReplicaCount     int      `json:"replica_count,omitempty"`
	VxLan            bool     `json:"vxlan,omitempty"`
	DNSMapping       bool     `json:"dns_mapping,omitempty"`
	Debug            bool     `json:"debug,omitempty"`
	OtpAuth          bool     `json:"otp_auth,omitempty"`
	LzoCompression   bool     `json:"lzo_compression,omitempty"`
	Cipher           string   `json:"cipher,omitempty"`
	Hash             string   `json:"hash,omitempty"`
	BlockOutsideDNS  bool     `json:"block_outside_dns,omitempty"`
	JumboFrames      bool     `json:"jumbo_frames,omitempty"`
	PreConnectMsg    string   `json:"pre_connect_msg,omitempty"`
	MSFixTimeout     int      `json:"mss_fix,omitempty"`
	Multihome        bool     `json:"multihome,omitempty"`
	Groups           []string `json:"groups,omitempty"`
	Organizations    []string `json:"organizations,omitempty"`
	WG               bool     `json:"wg,omitempty"`
}

// ServerOrganization represents an organization attached to a server
type ServerOrganization struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Server string `json:"server"`
}

// ListServers returns all servers
func (c *Client) ListServers() ([]Server, error) {
	resp, err := c.Get("/server")
	if err != nil {
		return nil, err
	}

	var servers []Server
	if err := json.Unmarshal(resp, &servers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal servers: %w", err)
	}

	return servers, nil
}

// GetServer returns a single server by ID
func (c *Client) GetServer(id string) (*Server, error) {
	resp, err := c.Get("/server/" + id)
	if err != nil {
		return nil, err
	}

	var server Server
	if err := json.Unmarshal(resp, &server); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server: %w", err)
	}

	return &server, nil
}

// GetServerByName returns a single server by name
func (c *Client) GetServerByName(name string) (*Server, error) {
	servers, err := c.ListServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Name == name {
			return &server, nil
		}
	}

	return nil, fmt.Errorf("server with name %q not found", name)
}

// CreateServer creates a new server
func (c *Client) CreateServer(server *Server) (*Server, error) {
	resp, err := c.Post("/server", server)
	if err != nil {
		return nil, err
	}

	var created Server
	if err := json.Unmarshal(resp, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created server: %w", err)
	}

	return &created, nil
}

// UpdateServer updates an existing server
func (c *Client) UpdateServer(id string, server *Server) (*Server, error) {
	resp, err := c.Put("/server/"+id, server)
	if err != nil {
		return nil, err
	}

	var updated Server
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated server: %w", err)
	}

	return &updated, nil
}

// DeleteServer deletes a server by ID
func (c *Client) DeleteServer(id string) error {
	_, err := c.Delete("/server/" + id)
	return err
}

// GetServerOrganizations returns organizations attached to a server
func (c *Client) GetServerOrganizations(serverID string) ([]ServerOrganization, error) {
	resp, err := c.Get("/server/" + serverID + "/organization")
	if err != nil {
		return nil, err
	}

	var orgs []ServerOrganization
	if err := json.Unmarshal(resp, &orgs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server organizations: %w", err)
	}

	return orgs, nil
}

// AttachOrganization attaches an organization to a server
func (c *Client) AttachOrganization(serverID, orgID string) error {
	_, err := c.Put("/server/"+serverID+"/organization/"+orgID, nil)
	return err
}

// DetachOrganization detaches an organization from a server
func (c *Client) DetachOrganization(serverID, orgID string) error {
	_, err := c.Delete("/server/" + serverID + "/organization/" + orgID)
	return err
}

// StartServer starts a server
func (c *Client) StartServer(id string) error {
	_, err := c.Put("/server/"+id+"/operation/start", nil)
	return err
}

// StopServer stops a server
func (c *Client) StopServer(id string) error {
	_, err := c.Put("/server/"+id+"/operation/stop", nil)
	return err
}

// WithServerStopped executes an operation while ensuring the server is stopped.
// If the server was online, it will be stopped before the operation and started after.
// Returns any error from the operation, or from stop/start if the operation succeeded.
func (c *Client) WithServerStopped(serverID string, operation func() error) error {
	server, err := c.GetServer(serverID)
	if err != nil {
		return fmt.Errorf("unable to get server status: %w", err)
	}

	wasOnline := server.Status == "online"
	if wasOnline {
		if err := c.StopServer(serverID); err != nil {
			return fmt.Errorf("unable to stop server: %w", err)
		}
	}

	// Execute the operation
	opErr := operation()

	// Restart server if it was online before
	if wasOnline {
		if startErr := c.StartServer(serverID); startErr != nil {
			if opErr != nil {
				return opErr // Return the operation error as primary
			}
			return fmt.Errorf("operation succeeded but failed to restart server: %w", startErr)
		}
	}

	return opErr
}
