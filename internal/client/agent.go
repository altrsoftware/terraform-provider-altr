// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateAgent creates a new agent
func (c *Client) CreateAgent(input CreateAgentInput) (*Agent, error) {
	resp, err := c.makeRequest(http.MethodPost, "/agents", input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	var agent Agent
	if err := handleAPIResponse(resp, &agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &agent, nil
}

// GetAgent retrieves an agent by ID
func (c *Client) GetAgent(agentID string) (*Agent, error) {
	resp, err := c.makeRequest(http.MethodGet, "/agents/"+url.PathEscape(agentID), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var agent Agent
	if err := handleAPIResponse(resp, &agent); err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// UpdateAgent updates an existing agent
func (c *Client) UpdateAgent(agentID string, input UpdateAgentInput) (*Agent, error) {
	resp, err := c.makeRequest(http.MethodPatch, "/agents/"+url.PathEscape(agentID), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	var agent Agent
	if err := handleAPIResponse(resp, &agent); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return &agent, nil
}

// DeleteAgent deletes an agent
func (c *Client) DeleteAgent(agentID string) error {
	resp, err := c.makeRequest(http.MethodDelete, "/agents/"+url.PathEscape(agentID), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}
