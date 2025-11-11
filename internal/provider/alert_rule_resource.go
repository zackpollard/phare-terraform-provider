// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/phare/terraform-provider-phare/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AlertRuleResource{}
var _ resource.ResourceWithImportState = &AlertRuleResource{}

func NewAlertRuleResource() resource.Resource {
	return &AlertRuleResource{}
}

// AlertRuleResource defines the resource implementation.
type AlertRuleResource struct {
	client *client.Client
}

// AlertRuleResourceModel describes the resource data model.
type AlertRuleResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Event         types.String `tfsdk:"event"`
	IntegrationID types.Int64  `tfsdk:"integration_id"`
	RateLimit     types.Int64  `tfsdk:"rate_limit"`
	EventSettings types.Object `tfsdk:"event_settings"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type AlertEventSettingsModel struct {
	Type types.String `tfsdk:"type"`
}

func (r *AlertRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_rule"
}

func (r *AlertRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Phare alert rule that triggers notifications based on platform events.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the alert rule",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event": schema.StringAttribute{
				MarkdownDescription: "The event that triggers this alert rule",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"uptime.monitor.created",
						"uptime.monitor.updated",
						"uptime.monitor.deleted",
						"uptime.incident.created",
						"uptime.incident.acknowledged",
						"uptime.incident.resolved",
						"uptime.status_page.created",
						"uptime.status_page.updated",
						"uptime.status_page.deleted",
						"platform.integration.health.unhealthy",
					),
				},
			},
			"integration_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the integration to send alerts to",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"rate_limit": schema.Int64Attribute{
				MarkdownDescription: "Rate limit in minutes (0, 5, 30, 60, 120, 360, 1440)",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(0, 5, 30, 60, 120, 360, 1440),
				},
			},
			"event_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Settings for when the alert should trigger",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Trigger type (e.g., 'all' to trigger for all events)",
						Required:            true,
					},
				},
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "Optional project ID to scope the alert rule to a specific project",
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the alert rule was created",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the alert rule was last updated",
				Computed:            true,
			},
		},
	}
}

func (r *AlertRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *AlertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AlertRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract event_settings
	var eventSettings AlertEventSettingsModel
	resp.Diagnostics.Append(data.EventSettings.As(ctx, &eventSettings, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := &client.AlertRule{
		Event:         data.Event.ValueString(),
		IntegrationID: int(data.IntegrationID.ValueInt64()),
		RateLimit:     int(data.RateLimit.ValueInt64()),
		EventSettings: client.AlertEventSettings{
			Type: eventSettings.Type.ValueString(),
		},
	}

	if !data.ProjectID.IsNull() {
		projectID := int(data.ProjectID.ValueInt64())
		rule.ProjectID = &projectID
	}

	tflog.Debug(ctx, "Creating alert rule", map[string]any{"event": data.Event.ValueString()})

	created, err := r.client.CreateAlertRule(ctx, rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create alert rule", err.Error())
		return
	}

	// Get the created rule ID
	if created.ID == nil {
		resp.Diagnostics.AddError("Failed to create alert rule", "API did not return an alert rule ID")
		return
	}

	// Read back the alert rule to get all fields
	fullRule, err := r.client.GetAlertRule(ctx, *created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created alert rule", err.Error())
		return
	}

	r.apiToTerraformModel(fullRule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AlertRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading alert rule", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid alert rule ID", fmt.Sprintf("Failed to parse alert rule ID: %s", err.Error()))
		return
	}

	rule, err := r.client.GetAlertRule(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read alert rule", err.Error())
		return
	}

	r.apiToTerraformModel(rule, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AlertRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract event_settings
	var eventSettings AlertEventSettingsModel
	resp.Diagnostics.Append(data.EventSettings.As(ctx, &eventSettings, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := &client.AlertRule{
		Event:         data.Event.ValueString(),
		IntegrationID: int(data.IntegrationID.ValueInt64()),
		RateLimit:     int(data.RateLimit.ValueInt64()),
		EventSettings: client.AlertEventSettings{
			Type: eventSettings.Type.ValueString(),
		},
	}

	if !data.ProjectID.IsNull() {
		projectID := int(data.ProjectID.ValueInt64())
		rule.ProjectID = &projectID
	}

	tflog.Debug(ctx, "Updating alert rule", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid alert rule ID", fmt.Sprintf("Failed to parse alert rule ID: %s", err.Error()))
		return
	}

	updated, err := r.client.UpdateAlertRule(ctx, id, rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update alert rule", err.Error())
		return
	}

	r.apiToTerraformModel(updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AlertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AlertRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting alert rule", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid alert rule ID", fmt.Sprintf("Failed to parse alert rule ID: %s", err.Error()))
		return
	}

	if err := r.client.DeleteAlertRule(ctx, id); err != nil {
		resp.Diagnostics.AddError("Failed to delete alert rule", err.Error())
		return
	}
}

func (r *AlertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AlertRuleResource) apiToTerraformModel(rule *client.AlertRule, data *AlertRuleResourceModel) {
	if rule.ID != nil {
		data.ID = types.StringValue(fmt.Sprintf("%d", *rule.ID))
	}
	data.Event = types.StringValue(rule.Event)
	data.IntegrationID = types.Int64Value(int64(rule.IntegrationID))
	data.RateLimit = types.Int64Value(int64(rule.RateLimit))

	// Convert event_settings - only if returned by API (currently not returned)
	// If not returned, we keep the value from the plan/state
	if rule.EventSettings.Type != "" {
		eventSettingsObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"type": types.StringType,
			},
			map[string]attr.Value{
				"type": types.StringValue(rule.EventSettings.Type),
			},
		)
		data.EventSettings = eventSettingsObj
	}
	// If EventSettings is empty, we don't overwrite data.EventSettings,
	// which preserves the value from the plan/state

	if rule.ProjectID != nil {
		data.ProjectID = types.Int64Value(int64(*rule.ProjectID))
	} else {
		data.ProjectID = types.Int64Null()
	}

	if rule.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*rule.CreatedAt)
	}
	if rule.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*rule.UpdatedAt)
	}
}
