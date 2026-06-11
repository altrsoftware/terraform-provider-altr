// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateServiceUser creates a new service user for a repository
func (c *Client) CreateServiceUser(repoName string, input CreateServiceUserInput) (*ServiceUser, error) {
	resp, err := c.makeRequest(http.MethodPost, fmt.Sprintf("/repos/%s/serviceusers", url.PathEscape(repoName)), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to create service user: %w", err)
	}

	var serviceUser ServiceUser
	if err := handleAPIResponse(resp, &serviceUser); err != nil {
		return nil, fmt.Errorf("failed to create service user: %w", err)
	}

	return &serviceUser, nil
}

// GetServiceUser retrieves a service user by repo name and username
func (c *Client) GetServiceUser(repoName, username string) (*ServiceUser, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/repos/%s/serviceusers/%s", url.PathEscape(repoName), url.PathEscape(username)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get service user: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var serviceUser ServiceUser
	if err := handleAPIResponse(resp, &serviceUser); err != nil {
		return nil, fmt.Errorf("failed to get service user: %w", err)
	}

	return &serviceUser, nil
}

// UpdateServiceUser updates an existing service user
func (c *Client) UpdateServiceUser(repoName, username string, input UpdateServiceUserInput) (*ServiceUser, error) {
	resp, err := c.makeRequest(http.MethodPatch, fmt.Sprintf("/repos/%s/serviceusers/%s", url.PathEscape(repoName), url.PathEscape(username)), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to update service user: %w", err)
	}

	var serviceUser ServiceUser
	if err := handleAPIResponse(resp, &serviceUser); err != nil {
		return nil, fmt.Errorf("failed to update service user: %w", err)
	}

	return &serviceUser, nil
}

// DeleteServiceUser deletes a service user
func (c *Client) DeleteServiceUser(repoName, username string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/repos/%s/serviceusers/%s", url.PathEscape(repoName), url.PathEscape(username)), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete service user: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete service user: %w", err)
	}

	return nil
}
