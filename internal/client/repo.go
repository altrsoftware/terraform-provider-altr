// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateRepo creates a new repo
func (c *Client) CreateRepo(input CreateRepoInput) (*Repo, error) {
	resp, err := c.makeRequest(http.MethodPost, "/repos", input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}

	var repo Repo
	if err := handleAPIResponse(resp, &repo); err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}

	return &repo, nil
}

// GetRepo retrieves a repo by name
func (c *Client) GetRepo(repoName string) (*Repo, error) {
	resp, err := c.makeRequest(http.MethodGet, "/repos/"+url.PathEscape(repoName), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var repo Repo
	if err := handleAPIResponse(resp, &repo); err != nil {
		return nil, fmt.Errorf("failed to get repo: %w", err)
	}

	return &repo, nil
}

// UpdateRepo updates an existing repo
func (c *Client) UpdateRepo(repoName string, input UpdateRepoInput) (*Repo, error) {
	resp, err := c.makeRequest(http.MethodPatch, "/repos/"+url.PathEscape(repoName), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to update repo: %w", err)
	}

	var repo Repo
	if err := handleAPIResponse(resp, &repo); err != nil {
		return nil, fmt.Errorf("failed to update repo: %w", err)
	}

	return &repo, nil
}

// DeleteRepo deletes a repo
func (c *Client) DeleteRepo(repoName string) error {
	resp, err := c.makeRequest(http.MethodDelete, "/repos/"+url.PathEscape(repoName), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete repo: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete repo: %w", err)
	}

	return nil
}
