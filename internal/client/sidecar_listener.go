package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// RegisterSidecarListener registers a new sidecar listener port
func (c *Client) RegisterSidecarListener(sidecarID string, input RegisterSidecarListenerInput) error {
	resp, err := c.makeRequest(http.MethodPost, fmt.Sprintf("/sidecars/%s/ports", url.PathEscape(sidecarID)), input, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to register sidecar listener: %w", err)
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to register sidecar listener: %w", err)
	}

	return nil
}

// GetSidecarListener retrieves a specific sidecar listener by sidecar ID and port
func (c *Client) GetSidecarListener(sidecarID string, port int) (*ListenerPort, error) {
	// Get all listeners for the sidecar and find the specific port
	listeners, err := c.ListSidecarListeners(sidecarID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sidecar listener: %w", err)
	}

	for _, listener := range listeners {
		if listener.Port == port {
			return &listener, nil
		}
	}

	return nil, nil // Listener not found
}

// ListSidecarListeners lists all listeners for a given sidecar
func (c *Client) ListSidecarListeners(sidecarID string) ([]ListenerPort, error) {
	resp, err := c.makeRequest(http.MethodGet, fmt.Sprintf("/sidecars/%s/ports", url.PathEscape(sidecarID)), nil, "sidecar")
	if err != nil {
		return nil, fmt.Errorf("failed to list sidecar listeners: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return []ListenerPort{}, nil
	}

	var output ListSidecarListenersOutput
	if err := handleAPIResponse(resp, &output); err != nil {
		return nil, fmt.Errorf("failed to list sidecar listeners: %w", err)
	}

	return output.SidecarListeners, nil
}

// DeregisterSidecarListener removes a sidecar listener
func (c *Client) DeregisterSidecarListener(sidecarID string, port int) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/sidecars/%s/ports/%s", url.PathEscape(sidecarID), strconv.Itoa(port)), nil, "sidecar")
	if err != nil {
		return fmt.Errorf("failed to deregister sidecar listener: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to deregister sidecar listener: %w", err)
	}

	return nil
}
