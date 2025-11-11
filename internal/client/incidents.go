// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Incident represents a Phare uptime incident (status page incident)
type Incident struct {
	ID                  *int    `json:"id,omitempty"`
	ProjectID           *int    `json:"project_id,omitempty"`
	Title               string  `json:"title"`
	Slug                string  `json:"slug"`
	Impact              string  `json:"impact"`
	State               string  `json:"state"`
	Description         string  `json:"description"`
	ExcludeFromDowntime bool    `json:"exclude_from_downtime"`
	Status              string  `json:"status"`
	IncidentAt          string  `json:"incident_at"`
	RecoveryAt          *string `json:"recovery_at,omitempty"`
	CreatedAt           *string `json:"created_at,omitempty"`
	UpdatedAt           *string `json:"updated_at,omitempty"`
}

// IncidentListResponse represents the response from listing incidents
type IncidentListResponse struct {
	Data []Incident `json:"data"`
}

// IncidentResponse represents the response from getting an incident
type IncidentResponse struct {
	Data Incident `json:"data"`
}

// GetIncident retrieves an incident by ID
func (c *Client) GetIncident(ctx context.Context, id int) (*Incident, error) {
	respBody, err := c.doRequest(ctx, "GET", fmt.Sprintf("/uptime/incidents/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	var incident Incident
	if err := json.Unmarshal(respBody, &incident); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &incident, nil
}

// DeleteIncident deletes an incident
func (c *Client) DeleteIncident(ctx context.Context, id int) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/uptime/incidents/%d", id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	return nil
}

// ListIncidents lists all incidents
func (c *Client) ListIncidents(ctx context.Context) ([]Incident, error) {
	respBody, err := c.doRequest(ctx, "GET", "/uptime/incidents", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	var resp IncidentListResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Data, nil
}
