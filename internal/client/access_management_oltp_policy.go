// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// AccessManagementOLTPPolicy represents an access management OLTP policy
type AccessManagementOLTPPolicy struct {
	ID               string                     `json:"policy_id"`
	Name             string                     `json:"policy_name"`
	Description      string                     `json:"description"`
	DatabaseTypeName string                     `json:"database_type_name"`
	DatabaseType     int64                      `json:"database_type"`
	CaseSensitivity  string                     `json:"case_sensitivity"`
	RepoName         string                     `json:"repo_name"`
	CreatedAt        string                     `json:"created_at"`
	UpdatedAt        string                     `json:"updated_at"`
	Rules            []AccessManagementOLTPRule `json:"rules"`
}

type AccessManagementOLTPRule struct {
	Type    string                       `json:"type"`
	Actors  []AccessManagementOLTPActor  `json:"actors"`
	Objects []AccessManagementOLTPObject `json:"objects"`
}

type AccessManagementOLTPActor struct {
	Type        string   `json:"type"`
	Condition   string   `json:"condition"`
	Identifiers []string `json:"identifiers"`
}

type AccessManagementOLTPObject struct {
	Type        string                           `json:"type"`
	Identifiers []AccessManagementOLTPIdentifier `json:"identifiers"`
}

type AccessManagementOLTPIdentifier struct {
	Database AccessManagementOLTPIdentifierPart `json:"database"`
	Schema   AccessManagementOLTPIdentifierPart `json:"schema"`
	Table    AccessManagementOLTPIdentifierPart `json:"table"`
	Column   AccessManagementOLTPIdentifierPart `json:"column"`
}

type AccessManagementOLTPIdentifierPart struct {
	Name     string `json:"name"`
	Wildcard bool   `json:"wildcard"`
}

// CreateAccessManagementOLTPPolicyInput represents the input for creating an access management OLTP policy
type CreateAccessManagementOLTPPolicyInput struct {
	Name             string                     `json:"policy_name"`
	Description      string                     `json:"description"`
	DatabaseTypeName string                     `json:"database_type_name"`
	DatabaseType     int64                      `json:"database_type"`
	CaseSensitivity  string                     `json:"case_sensitivity"`
	RepoName         string                     `json:"repo_name"`
	Rules            []AccessManagementOLTPRule `json:"rules"`
}

// UpdateAccessManagementOLTPPolicyInput represents the input for updating an access management OLTP policy
type UpdateAccessManagementOLTPPolicyInput struct {
	Name        string                     `json:"policy_name"`
	Description string                     `json:"description"`
	Rules       []AccessManagementOLTPRule `json:"rules"`
}

// CreateAccessManagementOLTPPolicy creates a new access management OLTP policy
func (c *Client) CreateAccessManagementOLTPPolicy(input CreateAccessManagementOLTPPolicyInput) (*AccessManagementOLTPPolicy, error) {
	resp, err := c.makeRequest(http.MethodPost, "/unified-policy/management/policy/accessManagement/oltp", input, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to create access management OLTP policy: %w", err)
	}

	var response struct {
		Data struct {
			Policy   AccessManagementOLTPPolicy `json:"policy"`
			PolicyID string                     `json:"policy_id"`
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

// GetAccessManagementOLTPPolicy retrieves an access management OLTP policy by ID
func (c *Client) GetAccessManagementOLTPPolicy(policyID string) (*AccessManagementOLTPPolicy, error) {
	fmt.Println("getting id", policyID)

	resp, err := c.makeRequest(http.MethodGet, "/unified-policy/management/policy/"+url.PathEscape(policyID), nil, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to get access management OLTP policy: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var response struct {
		Data AccessManagementOLTPPolicy `json:"data"`
	}

	if err := handleAPIResponse(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to get access management OLTP policy: %w", err)
	}

	return &response.Data, nil
}

// UpdateAccessManagementOLTPPolicy updates an existing access management OLTP policy
// Update is not currently supported
func (c *Client) UpdateAccessManagementOLTPPolicy(policyID string, input UpdateAccessManagementOLTPPolicyInput) (*AccessManagementOLTPPolicy, error) {
	resp, err := c.makeRequest(http.MethodPatch, "/unified-policy/management/access-management/oltp/"+url.PathEscape(policyID), input, "external")
	if err != nil {
		return nil, fmt.Errorf("failed to update access management OLTP policy: %w", err)
	}

	var policy AccessManagementOLTPPolicy
	if err := handleAPIResponse(resp, &policy); err != nil {
		return nil, fmt.Errorf("failed to update access management OLTP policy: %w", err)
	}

	return &policy, nil
}

// DeleteAccessManagementOLTPPolicy deletes an access management OLTP policy
func (c *Client) DeleteAccessManagementOLTPPolicy(policyID string) error {
	resp, err := c.makeRequest(http.MethodDelete, "/unified-policy/management/policy/"+url.PathEscape(policyID), nil, "external")
	if err != nil {
		return fmt.Errorf("failed to delete access management OLTP policy: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete access management OLTP policy: %w", err)
	}

	return nil
}
