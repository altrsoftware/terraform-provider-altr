// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateRepoSidecarBinding creates a new repo sidecar binding
func (c *Client) CreateRepoSidecarBinding(sidecarID, repoName string, port int) error {
	resp, err := c.makeRequest(http.MethodPost, fmt.Sprintf("/sidecars/%s/bindings/ports/%d/repos/%s", url.PathEscape(sidecarID), port, url.PathEscape(repoName)), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to create repo sidecar binding: %w", err)
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to create repo sidecar binding: %w", err)
	}

	return nil
}

// GetRepoSidecarBinding retrieves a specific repo sidecar binding
func (c *Client) GetRepoSidecarBinding(sidecarID, repoName string, port int) (*RepoSidecarBinding, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/sidecars/%s/bindings/ports/%d/repos/%s", url.PathEscape(sidecarID), port, url.PathEscape(repoName)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get repo sidecar binding: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var output GetRepoBindOutput
	if err := handleAPIResponse(resp, &output); err != nil {
		return nil, fmt.Errorf("failed to get repo sidecar binding: %w", err)
	}

	return &output.RepoSidecarBinding, nil
}

// DeleteRepoSidecarBinding deletes a repo sidecar binding
func (c *Client) DeleteRepoSidecarBinding(sidecarID, repoName string, port int) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/sidecars/%s/bindings/ports/%d/repos/%s", url.PathEscape(sidecarID), port, url.PathEscape(repoName)), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete repo sidecar binding: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete repo sidecar binding: %w", err)
	}

	return nil
}

// ListSidecarBindings lists all bindings for a given sidecar
func (c *Client) ListSidecarBindings(sidecarID string) ([]RepoSidecarBinding, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/sidecars/%s/bindings", url.PathEscape(sidecarID)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to list sidecar bindings: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return []RepoSidecarBinding{}, nil
	}

	var output ListBindingsOutput
	if err := handleAPIResponse(resp, &output); err != nil {
		return nil, fmt.Errorf("failed to list sidecar bindings: %w", err)
	}

	return output.RepoBindings, nil
}

// ListRepoBindings lists all bindings for a given repo
func (c *Client) ListRepoBindings(repoName string) ([]RepoSidecarBinding, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/repos/%s/bindings", url.PathEscape(repoName)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to list repo bindings: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return []RepoSidecarBinding{}, nil
	}

	var output ListBindingsOutput
	if err := handleAPIResponse(resp, &output); err != nil {
		return nil, fmt.Errorf("failed to list repo bindings: %w", err)
	}

	return output.RepoBindings, nil
}
