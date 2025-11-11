// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// AlertRule represents a Phare alert rule
type AlertRule struct {
	ID            *int               `json:"id,omitempty"`
	Event         string             `json:"event"`
	IntegrationID int                `json:"integration_id"`
	RateLimit     int                `json:"rate_limit"`
	EventSettings AlertEventSettings `json:"event_settings"`
	ProjectID     *int               `json:"project_id,omitempty"`
	CreatedAt     *string            `json:"created_at,omitempty"`
	UpdatedAt     *string            `json:"updated_at,omitempty"`
}

// AlertEventSettings represents the event settings for an alert rule
type AlertEventSettings struct {
	Type string `json:"type"`
}

// AlertRuleListResponse represents the response from listing alert rules
type AlertRuleListResponse struct {
	Data []AlertRule `json:"data"`
}

// AlertRuleResponse represents the response from creating/getting an alert rule
type AlertRuleResponse struct {
	Data AlertRule `json:"data"`
}

// CreateAlertRule creates a new alert rule
func (c *Client) CreateAlertRule(ctx context.Context, rule *AlertRule) (*AlertRule, error) {
	respBody, err := c.doRequest(ctx, "POST", "/alert-rules", rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	var created AlertRule
	if err := json.Unmarshal(respBody, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &created, nil
}

// GetAlertRule retrieves an alert rule by ID
func (c *Client) GetAlertRule(ctx context.Context, id int) (*AlertRule, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/alert-rules/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	var rule AlertRule
	if err := json.Unmarshal(respBody, &rule); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &rule, nil
}

// UpdateAlertRule updates an existing alert rule
func (c *Client) UpdateAlertRule(ctx context.Context, id int, rule *AlertRule) (*AlertRule, error) {
	respBody, err := c.doRequest(ctx, "POST", fmt.Sprintf("/alert-rules/%d", id), rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update alert rule: %w", err)
	}

	var updated AlertRule
	if err := json.Unmarshal(respBody, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &updated, nil
}

// DeleteAlertRule deletes an alert rule
func (c *Client) DeleteAlertRule(ctx context.Context, id int) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/alert-rules/%d", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	return nil
}

// ListAlertRules lists all alert rules
func (c *Client) ListAlertRules(ctx context.Context) ([]AlertRule, error) {
	respBody, err := c.doRequest(ctx, "GET", "/alert-rules", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert rules: %w", err)
	}

	var resp AlertRuleListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Data, nil
}
