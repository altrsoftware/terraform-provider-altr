// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateSidecar creates a new sidecar
func (c *Client) CreateSidecar(input CreateSidecarInput) (*Sidecar, error) {
	resp, err := c.makeRequest(http.MethodPost, "/sidecars", input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar: %w", err)
	}

	var sidecar Sidecar
	if err := handleAPIResponse(resp, &sidecar); err != nil {
		return nil, fmt.Errorf("failed to create sidecar: %w", err)
	}

	return &sidecar, nil
}

// GetSidecar retrieves a sidecar by ID
func (c *Client) GetSidecar(sidecarID string) (*Sidecar, error) {
	resp, err := c.makeRequest(http.MethodGet, "/sidecars/"+url.PathEscape(sidecarID), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get sidecar: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var sidecar Sidecar
	if err := handleAPIResponse(resp, &sidecar); err != nil {
		return nil, fmt.Errorf("failed to get sidecar: %w", err)
	}

	return &sidecar, nil
}

// UpdateSidecar updates an existing sidecar
func (c *Client) UpdateSidecar(sidecarID string, input UpdateSidecarInput) (*Sidecar, error) {
	resp, err := c.makeRequest(http.MethodPatch, "/sidecars/"+url.PathEscape(sidecarID), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to update sidecar: %w", err)
	}

	var sidecar Sidecar
	if err := handleAPIResponse(resp, &sidecar); err != nil {
		return nil, fmt.Errorf("failed to update sidecar: %w", err)
	}

	return &sidecar, nil
}

// DeleteSidecar deletes a sidecar
func (c *Client) DeleteSidecar(sidecarID string) error {
	resp, err := c.makeRequest(http.MethodDelete, "/sidecars/"+url.PathEscape(sidecarID), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete sidecar: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete sidecar: %w", err)
	}

	return nil
}
