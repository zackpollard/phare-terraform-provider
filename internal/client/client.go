// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.phare.io"
	DefaultTimeout = 30 * time.Second
)

// Client represents a Phare API client
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Phare API client
func NewClient(apiToken, baseURL string) (*Client, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("api_token is required")
	}

	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// ErrorResponse represents a Phare API error response
type ErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

// doRequest performs an HTTP request with proper authentication and error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}

		if len(errResp.Errors) > 0 {
			return nil, fmt.Errorf("API error (status %d): %s - validation errors: %+v",
				resp.StatusCode, errResp.Message, errResp.Errors)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Message)
	}

	return respBody, nil
}
