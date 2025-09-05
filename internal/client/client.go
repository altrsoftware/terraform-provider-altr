package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	externalURL string // URL for external API calls
	sidecarURL  string // URL for sidecar API calls
	auth        string
}

func NewClient(orgID, apiKey, secret, baseURL string) (*Client, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", apiKey, secret)))
	// T
	if strings.Contains(baseURL, "{orgID}") {
		baseURL = strings.Replace(baseURL, "{orgID}", orgID, -1)
	}
	if strings.Contains(baseURL, "altrnet") {
		externalURL := strings.Replace(baseURL, "altrnet", "api", 1) + "/v1"
		sidecarURL := strings.Replace(baseURL, "altrnet", "sc-control", 1) + "/v1"

		return &Client{
			httpClient:  &http.Client{Timeout: 30 * time.Second},
			baseURL:     baseURL,
			externalURL: externalURL, // For altrnet, external and sidecar URLs
			sidecarURL:  sidecarURL,
			auth:        auth,
		}, nil
	} else {
		return nil, fmt.Errorf("base URL must contain 'altrnet' for altrnet API")
	}
}

func (c *Client) makeRequest(method, endpoint string, body interface{}, apiGateway string) (*http.Response, error) {

	url := ""
	if apiGateway == "" || apiGateway == "altrnet" {
		url = c.baseURL + endpoint
	} else if apiGateway == "external" {
		url = c.externalURL + endpoint
	} else if apiGateway == "sidecar" {
		url = c.sidecarURL + endpoint
	} else {
		return nil, fmt.Errorf("unknown API gateway: %s", apiGateway)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	tflog.Info(context.Background(), "Making request", map[string]interface{}{
		"url":    url,
		"method": method,
		"body":   body,
	})

	req.Header.Set("Authorization", "Basic "+c.auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	tflog.Trace(context.Background(), "Making request", map[string]interface{}{
		"url":     url,
		"method":  method,
		"body":    body,
		"headers": req.Header,
	})

	return c.httpClient.Do(req)
}

// Helper function to handle API responses
func handleAPIResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if v != nil {
			return json.NewDecoder(resp.Body).Decode(v)
		}
		return nil
	}

	resBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	tflog.Trace(context.Background(), "API response", map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(resBytes),
	})

	// Check if we have an API error
	var apiError APIError
	if err := json.Unmarshal(resBytes, &apiError); err != nil {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	apiError.StatusCode = resp.StatusCode
	if apiError.Response.Message != "" || apiError.Response.ErrorCode != 0 {
		return fmt.Errorf("API error (status %d): %w", resp.StatusCode, apiError)
	}

	// We aren't working with an APIError response. Let's try to parse the response as an APIErrorResponse.
	var apiResponse APIErrorResponse
	if err := json.Unmarshal(resBytes, &apiResponse); err != nil {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if apiResponse.Message != "" || apiResponse.ErrorCode != 0 {
		apiError.Response = apiResponse
		return fmt.Errorf("API error (status %d): %w", resp.StatusCode, apiError)
	}

	// Error response is unknown. set the message to the response body.
	apiError.Response.Message = string(resBytes)

	return fmt.Errorf("API error (status %d): %w", resp.StatusCode, apiError)
}
