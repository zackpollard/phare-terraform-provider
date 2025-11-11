// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/phare/terraform-provider-phare/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UptimeIncidentDataSource{}

func NewUptimeIncidentDataSource() datasource.DataSource {
	return &UptimeIncidentDataSource{}
}

// UptimeIncidentDataSource defines the data source implementation.
type UptimeIncidentDataSource struct {
	client *client.Client
}

// UptimeIncidentDataSourceModel describes the data source data model.
type UptimeIncidentDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.Int64  `tfsdk:"project_id"`
	Title               types.String `tfsdk:"title"`
	Slug                types.String `tfsdk:"slug"`
	Impact              types.String `tfsdk:"impact"`
	State               types.String `tfsdk:"state"`
	Description         types.String `tfsdk:"description"`
	ExcludeFromDowntime types.Bool   `tfsdk:"exclude_from_downtime"`
	Status              types.String `tfsdk:"status"`
	IncidentAt          types.String `tfsdk:"incident_at"`
	RecoveryAt          types.String `tfsdk:"recovery_at"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (d *UptimeIncidentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uptime_incident"
}

func (d *UptimeIncidentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a Phare uptime incident (status page incident).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the incident",
				Required:            true,
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the project this incident belongs to",
				Computed:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The title of the incident",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The URL-friendly slug for the incident",
				Computed:            true,
			},
			"impact": schema.StringAttribute{
				MarkdownDescription: "The impact level of the incident (e.g., majorOutage, degradedPerformance)",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The current state of the incident (e.g., investigating, identified, monitoring)",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the incident",
				Computed:            true,
			},
			"exclude_from_downtime": schema.BoolAttribute{
				MarkdownDescription: "Whether this incident is excluded from downtime calculations",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the incident (ongoing or resolved)",
				Computed:            true,
			},
			"incident_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the incident occurred",
				Computed:            true,
			},
			"recovery_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the incident was recovered (if resolved)",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the incident was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the incident was last updated",
				Computed:            true,
			},
		},
	}
}

func (d *UptimeIncidentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *UptimeIncidentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UptimeIncidentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading uptime incident", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid incident ID", fmt.Sprintf("Failed to parse incident ID: %s", err.Error()))
		return
	}

	incident, err := d.client.GetIncident(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read incident", err.Error())
		return
	}

	// Convert API model to Terraform model
	if incident.ID != nil {
		data.ID = types.StringValue(fmt.Sprintf("%d", *incident.ID))
	}

	if incident.ProjectID != nil {
		data.ProjectID = types.Int64Value(int64(*incident.ProjectID))
	} else {
		data.ProjectID = types.Int64Null()
	}

	data.Title = types.StringValue(incident.Title)
	data.Slug = types.StringValue(incident.Slug)
	data.Impact = types.StringValue(incident.Impact)
	data.State = types.StringValue(incident.State)
	data.Description = types.StringValue(incident.Description)
	data.ExcludeFromDowntime = types.BoolValue(incident.ExcludeFromDowntime)
	data.Status = types.StringValue(incident.Status)
	data.IncidentAt = types.StringValue(incident.IncidentAt)

	if incident.RecoveryAt != nil {
		data.RecoveryAt = types.StringValue(*incident.RecoveryAt)
	} else {
		data.RecoveryAt = types.StringNull()
	}

	if incident.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*incident.CreatedAt)
	}
	if incident.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*incident.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
