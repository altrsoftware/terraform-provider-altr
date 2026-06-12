// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateAgentTask creates a new task for an agent
func (c *Client) CreateAgentTask(agentID string, input CreateAgentTaskInput) (*AgentTask, error) {
	resp, err := c.makeRequest(http.MethodPost, fmt.Sprintf("/agents/%s/tasks", url.PathEscape(agentID)), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to create agent task: %w", err)
	}

	var task AgentTask
	if err := handleAPIResponse(resp, &task); err != nil {
		return nil, fmt.Errorf("failed to create agent task: %w", err)
	}

	return &task, nil
}

// GetAgentTask retrieves a task by agent ID and task ID
func (c *Client) GetAgentTask(agentID, taskID string) (*AgentTask, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/agents/%s/tasks/%s", url.PathEscape(agentID), url.PathEscape(taskID)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get agent task: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var task AgentTask
	if err := handleAPIResponse(resp, &task); err != nil {
		return nil, fmt.Errorf("failed to get agent task: %w", err)
	}

	return &task, nil
}

// UpdateAgentTask updates an existing agent task
func (c *Client) UpdateAgentTask(agentID, taskID string, input UpdateAgentTaskInput) (*AgentTask, error) {
	resp, err := c.makeRequest(http.MethodPatch, fmt.Sprintf("/agents/%s/tasks/%s", url.PathEscape(agentID), url.PathEscape(taskID)), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to update agent task: %w", err)
	}

	var task AgentTask
	if err := handleAPIResponse(resp, &task); err != nil {
		return nil, fmt.Errorf("failed to update agent task: %w", err)
	}

	return &task, nil
}

// DeleteAgentTask deletes an agent task
func (c *Client) DeleteAgentTask(agentID, taskID string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/agents/%s/tasks/%s", url.PathEscape(agentID), url.PathEscape(taskID)), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete agent task: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete agent task: %w", err)
	}

	return nil
}
