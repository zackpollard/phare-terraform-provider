// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// StatusPage represents a Phare status page
type StatusPage struct {
	ID                  *int              `json:"id,omitempty"`
	Name                string            `json:"name"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	SearchEngineIndexed bool              `json:"search_engine_indexed"`
	WebsiteURL          string            `json:"website_url"`
	Subdomain           *string           `json:"subdomain,omitempty"`
	Domain              *string           `json:"domain,omitempty"`
	Timeframe           *int              `json:"timeframe,omitempty"`
	Colors              StatusPageColors  `json:"colors"`
	Components          []StatusComponent `json:"components"`
	Logo                *string           `json:"logo,omitempty"`
	Favicon             *string           `json:"favicon,omitempty"`
	CreatedAt           *string           `json:"created_at,omitempty"`
	UpdatedAt           *string           `json:"updated_at,omitempty"`
}

// StatusPageColors represents the color scheme for a status page
type StatusPageColors struct {
	Operational         string `json:"operational"`
	DegradedPerformance string `json:"degradedPerformance"`
	PartialOutage       string `json:"partialOutage"`
	MajorOutage         string `json:"majorOutage"`
	Maintenance         string `json:"maintenance"`
	Empty               string `json:"empty"`
}

// StatusComponent represents a component on a status page
type StatusComponent struct {
	ComponentableType string `json:"componentable_type"`
	ComponentableID   int    `json:"componentable_id"`
}

// StatusPageListResponse represents the response from listing status pages
type StatusPageListResponse struct {
	Data []StatusPage `json:"data"`
}

// StatusPageResponse represents the response from creating/getting a status page
type StatusPageResponse struct {
	Data StatusPage `json:"data"`
}

// CreateStatusPage creates a new status page
func (c *Client) CreateStatusPage(ctx context.Context, page *StatusPage) (*StatusPage, error) {
	respBody, err := c.doRequest(ctx, "POST", "/uptime/status-pages", page)
	if err != nil {
		return nil, fmt.Errorf("failed to create status page: %w", err)
	}

	var created StatusPage
	if err := json.Unmarshal(respBody, &created); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &created, nil
}

// GetStatusPage retrieves a status page by ID
func (c *Client) GetStatusPage(ctx context.Context, id int) (*StatusPage, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/uptime/status-pages/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status page: %w", err)
	}

	var page StatusPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &page, nil
}

// UpdateStatusPage updates an existing status page
func (c *Client) UpdateStatusPage(ctx context.Context, id int, page *StatusPage) (*StatusPage, error) {
	respBody, err := c.doRequest(ctx, "POST", fmt.Sprintf("/uptime/status-pages/%d", id), page)
	if err != nil {
		return nil, fmt.Errorf("failed to update status page: %w", err)
	}

	var updated StatusPage
	if err := json.Unmarshal(respBody, &updated); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &updated, nil
}

// DeleteStatusPage deletes a status page
func (c *Client) DeleteStatusPage(ctx context.Context, id int) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/uptime/status-pages/%d", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete status page: %w", err)
	}

	return nil
}

// ListStatusPages lists all status pages
func (c *Client) ListStatusPages(ctx context.Context) ([]StatusPage, error) {
	respBody, err := c.doRequest(ctx, "GET", "/uptime/status-pages", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list status pages: %w", err)
	}

	var resp StatusPageListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Data, nil
}
