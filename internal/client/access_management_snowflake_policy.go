package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// AccessManagementSnowflakePolicy represents an access management Snowflake policy
type AccessManagementSnowflakePolicy struct {
	ID           string                          `json:"policy_id"`
	Name         string                          `json:"policy_name"`
	Description  string                          `json:"description"`
	CreatedAt    string                          `json:"created_at"`
	UpdatedAt    string                          `json:"updated_at"`
	Rules        []AccessManagementSnowflakeRule `json:"rules"` // This is for the POST, the other rule arrays are for GET
	PendingRules []AccessManagementSnowflakeRule `json:"rules_pending"`
	AppliedRules []AccessManagementSnowflakeRule `json:"rules_applied"`
	FailedRules  []AccessManagementSnowflakeRule `json:"rules_failed"`
}

type AccessManagementSnowflakeRule struct {
	Actors        []AccessManagementSnowflakeActor        `json:"actors"`
	Objects       []AccessManagementSnowflakeObject       `json:"objects,omitempty"`
	TaggedObjects []AccessManagementSnowflakeTaggedObject `json:"tagged_objects,omitempty"`
	Access        []AccessManagementSnowflakeAccess       `json:"access"`
}

type AccessManagementSnowflakeActor struct {
	Type        string   `json:"type"`
	Condition   string   `json:"condition,omitempty"`
	Identifiers []string `json:"identifiers"`
}

type AccessManagementSnowflakeObject struct {
	Type                      string                                               `json:"type"`
	Condition                 string                                               `json:"condition"`
	Identifiers               []string                                             `json:"identifiers,omitempty"`
	FullyQualifiedIdentifiers []AccessManagementSnowflakeFullyQualifiedIdentifiers `json:"fully_qualified_identifiers,omitempty"`
}

type AccessManagementSnowflakeFullyQualifiedIdentifiers struct {
	Database string `json:"database,omitempty"`
	Schema   string `json:"schema,omitempty"`
	Table    string `json:"table,omitempty"`
	View     string `json:"view,omitempty"`
}

type AccessManagementSnowflakeTaggedObject struct {
	CheckAgainst []string                              `json:"check_against"`
	TaggedWith   []AccessManagementSnowflakeTaggedWith `json:"tagged_with"`
	TagCondition string                                `json:"tag_condition"`
}

type AccessManagementSnowflakeTaggedWith struct {
	Database string `json:"database"`
	Schema   string `json:"schema"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

type AccessManagementSnowflakeAccess struct {
	Name string `json:"name"`
}

type AccessManagementPolicyMaintenance struct {
	Rate  string `json:"rate,omitempty" tfsdk:"rate"`
	Value string `json:"value,omitempty" tfsdk:"value"`
}

// CreateAccessManagementSnowflakePolicyInput represents the input for creating an access management Snowflake policy
type CreateAccessManagementSnowflakePolicyInput struct {
	Name              string                             `json:"policy_name"`
	Description       string                             `json:"description"`
	Rules             []AccessManagementSnowflakeRule    `json:"rules"`
	ConnectionIds     []int64                            `json:"connection_ids"`
	PolicyMaintenance *AccessManagementPolicyMaintenance `json:"policy_maintenance,omitempty"`
}

// UpdateAccessManagementSnowflakePolicyInput represents the input for updating an access management Snowflake policy
type UpdateAccessManagementSnowflakePolicyInput struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	Rules       []AccessManagementSnowflakeRule `json:"rules"`
}

// CreateAccessManagementSnowflakePolicy creates a new access management Snowflake policy
func (c *Client) CreateAccessManagementSnowflakePolicy(input CreateAccessManagementSnowflakePolicyInput) (*AccessManagementSnowflakePolicy, error) {
	resp, err := c.makeRequest(http.MethodPost, "/unified-policy/management/policy/accessManagement/snowflake", input, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to create access management Snowflake policy: %w", err)
	}

	fmt.Printf("*using input %+v\n", input)

	var response struct {
		Data struct {
			Policy   AccessManagementSnowflakePolicy `json:"policy"`
			PolicyID string                          `json:"policy_id"`
		} `json:"data"`
	}

	// Parse the response into the temporary structure
	if err := handleAPIResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response.Data.Policy.ID = response.Data.PolicyID

	// Return the parsed policy
	return &response.Data.Policy, nil
}

// GetAccessManagementSnowflakePolicy retrieves an access management Snowflake policy by ID
func (c *Client) GetAccessManagementSnowflakePolicy(policyID string) (*AccessManagementSnowflakePolicy, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/unified-policy/management/policy/%s", url.PathEscape(policyID)), nil, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to get access management Snowflake policy: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var response struct {
		Data AccessManagementSnowflakePolicy `json:"data"`
	}

	if err := handleAPIResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to get access management Snowflake policy: %w", err)
	}

	return &response.Data, nil
}

// UpdateAccessManagementSnowflakePolicy updates an existing access management Snowflake policy
// NOT SUPPORTED
func (c *Client) UpdateAccessManagementSnowflakePolicy(policyID string, input UpdateAccessManagementSnowflakePolicyInput) (*AccessManagementSnowflakePolicy, error) {
	resp, err := c.makeRequest(http.MethodPatch, fmt.Sprintf("/unified-policy/management/access-management/snowflake/%s", url.PathEscape(policyID)), input, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to update access management Snowflake policy: %w", err)
	}

	var policy AccessManagementSnowflakePolicy
	if err := handleAPIResponse(resp, &policy); err != nil {
		return nil, fmt.Errorf("failed to update access management Snowflake policy: %w", err)
	}

	return &policy, nil
}

// DeleteAccessManagementSnowflakePolicy deletes an access management Snowflake policy
func (c *Client) DeleteAccessManagementSnowflakePolicy(policyID string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/unified-policy/management/policy/%s", url.PathEscape(policyID)), nil, "external")
	if err != nil {
		return fmt.Errorf("failed to delete access management Snowflake policy: %w", err)
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete access management Snowflake policy: %w", err)
	}

	return nil
}
