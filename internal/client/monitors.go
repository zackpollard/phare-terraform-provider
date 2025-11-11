// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Monitor represents a Phare uptime monitor
type Monitor struct {
	ID                    *int               `json:"id,omitempty"`
	Name                  string             `json:"name"`
	Protocol              string             `json:"protocol"`
	Request               MonitorRequest     `json:"request"`
	Interval              int                `json:"interval"`
	Timeout               int                `json:"timeout"`
	IncidentConfirmations int                `json:"incident_confirmations"`
	RecoveryConfirmations int                `json:"recovery_confirmations"`
	Regions               []string           `json:"regions"`
	SuccessAssertions     []SuccessAssertion `json:"success_assertions,omitempty"`
	Paused                *bool              `json:"paused,omitempty"`
	CreatedAt             *string            `json:"created_at,omitempty"`
	UpdatedAt             *string            `json:"updated_at,omitempty"`
}

// MonitorRequest represents the request configuration for a monitor
type MonitorRequest struct {
	// HTTP fields
	Method          *string         `json:"method,omitempty"`
	URL             *string         `json:"url,omitempty"`
	TLSSkipVerify   *bool           `json:"tls_skip_verify,omitempty"`
	Body            *string         `json:"body,omitempty"`
	FollowRedirects *bool           `json:"follow_redirects,omitempty"`
	UserAgentSecret *string         `json:"user_agent_secret,omitempty"`
	Headers         []RequestHeader `json:"headers,omitempty"`

	// TCP fields
	Host       *string `json:"host,omitempty"`
	Port       *string `json:"port,omitempty"`
	Connection *string `json:"connection,omitempty"`
}

// RequestHeader represents an HTTP header
type RequestHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SuccessAssertion represents a success assertion for a monitor
type SuccessAssertion struct {
	Type     string  `json:"type"`
	Operator *string `json:"operator,omitempty"`
	Value    *string `json:"value,omitempty"`
	Property *string `json:"property,omitempty"`
}

// MonitorListResponse represents the response from listing monitors
type MonitorListResponse struct {
	Data []Monitor `json:"data"`
}

// MonitorResponse represents the response from creating/getting a monitor
type MonitorResponse struct {
	Data Monitor `json:"data"`
}

// CreateMonitor creates a new uptime monitor
func (c *Client) CreateMonitor(ctx context.Context, monitor *Monitor) (*Monitor, error) {
	respBody, err := c.doRequest(ctx, "POST", "/uptime/monitors", monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}

	var created Monitor
	if err := json.Unmarshal(respBody, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &created, nil
}

// GetMonitor retrieves a monitor by ID
func (c *Client) GetMonitor(ctx context.Context, id int) (*Monitor, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/uptime/monitors/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	var monitor Monitor
	if err := json.Unmarshal(respBody, &monitor); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &monitor, nil
}

// UpdateMonitor updates an existing monitor
func (c *Client) UpdateMonitor(ctx context.Context, id int, monitor *Monitor) (*Monitor, error) {
	respBody, err := c.doRequest(ctx, "POST", fmt.Sprintf("/uptime/monitors/%d", id), monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to update monitor: %w", err)
	}

	var updated Monitor
	if err := json.Unmarshal(respBody, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &updated, nil
}

// DeleteMonitor deletes a monitor
func (c *Client) DeleteMonitor(ctx context.Context, id int) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/uptime/monitors/%d", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete monitor: %w", err)
	}

	return nil
}

// PauseMonitor pauses a monitor
func (c *Client) PauseMonitor(ctx context.Context, id int) error {
	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("/uptime/monitors/%d/pause", id), nil)
	if err != nil {
		return fmt.Errorf("failed to pause monitor: %w", err)
	}

	return nil
}

// ResumeMonitor resumes a paused monitor
func (c *Client) ResumeMonitor(ctx context.Context, id int) error {
	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("/uptime/monitors/%d/resume", id), nil)
	if err != nil {
		return fmt.Errorf("failed to resume monitor: %w", err)
	}

	return nil
}

// ListMonitors lists all monitors
func (c *Client) ListMonitors(ctx context.Context) ([]Monitor, error) {
	respBody, err := c.doRequest(ctx, "GET", "/uptime/monitors", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}

	var resp MonitorListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Data, nil
}
