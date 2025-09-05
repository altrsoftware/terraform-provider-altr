package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateRepoUser creates a new repo user
func (c *Client) CreateRepoUser(repoName string, input CreateRepoUserInput) (*RepoUser, error) {
	resp, err := c.makeRequest(http.MethodPost, fmt.Sprintf("/repos/%s/users", url.PathEscape(repoName)), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to create repo user: %w", err)
	}

	var repoUser RepoUser
	if err := handleAPIResponse(resp, &repoUser); err != nil {
		return nil, fmt.Errorf("failed to create repo user: %w", err)
	}

	return &repoUser, nil
}

// GetRepoUser retrieves a repo user by repo name and username
func (c *Client) GetRepoUser(repoName, username string) (*RepoUser, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/repos/%s/users/%s", url.PathEscape(repoName), url.PathEscape(username)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to get repo user: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	var repoUser RepoUser
	if err := handleAPIResponse(resp, &repoUser); err != nil {
		return nil, fmt.Errorf("failed to get repo user: %w", err)
	}

	return &repoUser, nil
}

// UpdateRepoUser updates an existing repo user
func (c *Client) UpdateRepoUser(repoName, username string, input UpdateRepoUserInput) (*RepoUser, error) {
	resp, err := c.makeRequest(http.MethodPatch, fmt.Sprintf("/repos/%s/users/%s", url.PathEscape(repoName), url.PathEscape(username)), input, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to update repo user: %w", err)
	}

	var repoUser RepoUser
	if err := handleAPIResponse(resp, &repoUser); err != nil {
		return nil, fmt.Errorf("failed to update repo user: %w", err)
	}

	return &repoUser, nil
}

// DeleteRepoUser deletes a repo user
func (c *Client) DeleteRepoUser(repoName, username string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/repos/%s/users/%s", url.PathEscape(repoName), url.PathEscape(username)), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to delete repo user: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete repo user: %w", err)
	}

	return nil
}
