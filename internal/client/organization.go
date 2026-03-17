package client

import (
	"encoding/json"
	"fmt"
)

// Organization represents a Pritunl organization
type Organization struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	AuthAPI    bool   `json:"auth_api,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
	AuthSecret string `json:"auth_secret,omitempty"`
	UserCount  int    `json:"user_count,omitempty"`
}

// ListOrganizations returns all organizations
func (c *Client) ListOrganizations() ([]Organization, error) {
	resp, err := c.Get("/organization")
	if err != nil {
		return nil, err
	}

	var orgs []Organization
	if err := json.Unmarshal(resp, &orgs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organizations: %w", err)
	}

	return orgs, nil
}

// GetOrganization returns a single organization by ID
func (c *Client) GetOrganization(id string) (*Organization, error) {
	resp, err := c.Get("/organization/" + id)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(resp, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
	}

	return &org, nil
}

// GetOrganizationByName returns a single organization by name
func (c *Client) GetOrganizationByName(name string) (*Organization, error) {
	orgs, err := c.ListOrganizations()
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if org.Name == name {
			return &org, nil
		}
	}

	return nil, fmt.Errorf("organization with name %q not found", name)
}

// CreateOrganization creates a new organization
func (c *Client) CreateOrganization(org *Organization) (*Organization, error) {
	resp, err := c.Post("/organization", org)
	if err != nil {
		return nil, err
	}

	var created Organization
	if err := json.Unmarshal(resp, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created organization: %w", err)
	}

	return &created, nil
}

// UpdateOrganization updates an existing organization
func (c *Client) UpdateOrganization(id string, org *Organization) (*Organization, error) {
	resp, err := c.Put("/organization/"+id, org)
	if err != nil {
		return nil, err
	}

	var updated Organization
	if err := json.Unmarshal(resp, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated organization: %w", err)
	}

	return &updated, nil
}

// DeleteOrganization deletes an organization by ID
func (c *Client) DeleteOrganization(id string) error {
	_, err := c.Delete("/organization/" + id)
	return err
}
