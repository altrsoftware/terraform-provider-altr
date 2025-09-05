package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// ImpersonationPolicy represents an impersonation policy
type ImpersonationPolicy struct {
	ID          string              `json:"policy_id"`
	Name        string              `json:"policy_name"`
	Description string              `json:"description"`
	RepoName    string              `json:"repo_name"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
	Rules       []ImpersonationRule `json:"rules"` // List of rules for the impersonation policy
}

// Actor represents an actor in the impersonation policy
type Actor struct {
	Type        string   `json:"type"`        // e.g., "idp_user", "idp_group"
	Identifiers []string `json:"identifiers"` // List of user or group identifiers
	Condition   string   `json:"condition"`   // e.g., "equals"
}

type ImpersonationRule struct {
	Actors  []Actor `json:"actors"`  // List of actors for the rule
	Targets []Actor `json:"targets"` // List of target users or groups
}

// CreateImpersonationPolicyInput represents the input for creating an impersonation policy
type CreateImpersonationPolicyInput struct {
	Name        string              `json:"policy_name"`
	Description string              `json:"description"`
	RepoName    string              `json:"repo_name"`
	Rules       []ImpersonationRule `json:"rules"` // List of rules for the impersonation policy
}

// UpdateImpersonationPolicyInput represents the input for updating an impersonation policy
type UpdateImpersonationPolicyInput struct {
	Name        string              `json:"policy_name"`
	Description string              `json:"description"`
	Rules       []ImpersonationRule `json:"rules"` // List of rules for the impersonation policy
}

// CreateImpersonationPolicy creates a new impersonation policy
func (c *Client) CreateImpersonationPolicy(input CreateImpersonationPolicyInput) (*ImpersonationPolicy, error) {
	resp, err := c.makeRequest(http.MethodPost, "/unified-policy/management/policy/impersonation", input, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to create impersonation policy: %w", err)
	}

	// Define a temporary structure to parse the response
	var response struct {
		Data struct {
			Policy   ImpersonationPolicy `json:"policy"`
			PolicyID string              `json:"policy_id"`
		} `json:"data"`
	}

	// Parse the response into the temporary structure
	if err := handleAPIResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse impersonation policy response: %w", err)
	}

	response.Data.Policy.ID = response.Data.PolicyID

	// Return the parsed policy
	return &response.Data.Policy, nil
}

// GetImpersonationPolicy retrieves an impersonation policy by ID
func (c *Client) GetImpersonationPolicy(policyID string) (*ImpersonationPolicy, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/unified-policy/management/policy/%s", url.PathEscape(policyID)), nil, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to get impersonation policy: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var response struct {
		Data ImpersonationPolicy `json:"data"`
	}

	if err := handleAPIResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to get impersonation policy: %w", err)
	}

	return &response.Data, nil
}

// UpdateImpersonationPolicy updates an existing impersonation policy
// NOT IMPLEMENTED
func (c *Client) UpdateImpersonationPolicy(policyID string, input UpdateImpersonationPolicyInput) (*ImpersonationPolicy, error) {
	resp, err := c.makeRequest(http.MethodPatch, fmt.Sprintf("/unified-policy/management/policy/impersonation/%s", url.PathEscape(policyID)), input, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to update impersonation policy: %w", err)
	}

	var policy ImpersonationPolicy
	if err := handleAPIResponse(resp, &policy); err != nil {
		return nil, fmt.Errorf("failed to update impersonation policy: %w", err)
	}

	return &policy, nil
}

// DeleteImpersonationPolicy deletes an impersonation policy
func (c *Client) DeleteImpersonationPolicy(policyID string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/unified-policy/management/policy/%s", url.PathEscape(policyID)), nil, "external")
	if err != nil {
		return fmt.Errorf("failed to delete impersonation policy: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete impersonation policy: %w", err)
	}

	return nil
}
