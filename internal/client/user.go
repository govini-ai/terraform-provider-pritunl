package client

import (
	"encoding/json"
	"fmt"
)

// User represents a Pritunl user
type User struct {
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name"`
	Type             string   `json:"type,omitempty"`
	AuthType         string   `json:"auth_type,omitempty"`
	DNSServers       []string `json:"dns_servers,omitempty"`
	DNSSuffix        string   `json:"dns_suffix,omitempty"`
	DnsMappingKind   string   `json:"dns_mapping_kind,omitempty"`
	Disabled         bool     `json:"disabled,omitempty"`
	NetworkLinks     []string `json:"network_links,omitempty"`
	PortForwarding   []string `json:"port_forwarding,omitempty"`
	Email            string   `json:"email,omitempty"`
	Status           bool     `json:"status,omitempty"`
	OtpSecret        string   `json:"otp_secret,omitempty"`
	ClientToClient   bool     `json:"client_to_client,omitempty"`
	MacAddresses     []string `json:"mac_addresses,omitempty"`
	YubicoID         string   `json:"yubico_id,omitempty"`
	SSO              string   `json:"sso,omitempty"`
	BypassSecondary  bool     `json:"bypass_secondary,omitempty"`
	Groups           []string `json:"groups,omitempty"`
	Audit            bool     `json:"audit,omitempty"`
	Gravatar         bool     `json:"gravatar,omitempty"`
	OtpAuth          bool     `json:"otp_auth,omitempty"`
	DeviceAuth       bool     `json:"device_auth,omitempty"`
	Organization     string   `json:"organization,omitempty"`
	OrganizationName string   `json:"organization_name,omitempty"`
	Pin              bool     `json:"pin,omitempty"`
	Servers          []string `json:"servers,omitempty"`
}

// ListUsers returns all users in an organization
func (c *Client) ListUsers(orgID string) ([]User, error) {
	resp, err := c.Get("/user/" + orgID)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	return users, nil
}

// GetUser returns a single user by ID
func (c *Client) GetUser(orgID, userID string) (*User, error) {
	resp, err := c.Get("/user/" + orgID + "/" + userID)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetUserByName returns a single user by name in an organization
func (c *Client) GetUserByName(orgID, name string) (*User, error) {
	users, err := c.ListUsers(orgID)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Name == name {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user with name %q not found in organization %q", name, orgID)
}

// CreateUser creates a new user in an organization
func (c *Client) CreateUser(orgID string, user *User) (*User, error) {
	resp, err := c.Post("/user/"+orgID, user)
	if err != nil {
		return nil, err
	}

	// The API returns an array with a single user
	var users []User
	if err := json.Unmarshal(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user returned from create")
	}

	return &users[0], nil
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(orgID, userID string, user *User) (*User, error) {
	resp, err := c.Put("/user/"+orgID+"/"+userID, user)
	if err != nil {
		return nil, err
	}

	var updated User
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated user: %w", err)
	}

	return &updated, nil
}

// DeleteUser deletes a user from an organization
func (c *Client) DeleteUser(orgID, userID string) error {
	_, err := c.Delete("/user/" + orgID + "/" + userID)
	return err
}

// GetUserKey returns the user's profile key/configuration
func (c *Client) GetUserKey(orgID, userID string) ([]byte, error) {
	return c.Get("/key/" + orgID + "/" + userID)
}
