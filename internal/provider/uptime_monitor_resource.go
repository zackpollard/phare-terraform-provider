// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/phare/terraform-provider-phare/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UptimeMonitorResource{}
var _ resource.ResourceWithImportState = &UptimeMonitorResource{}

func NewUptimeMonitorResource() resource.Resource {
	return &UptimeMonitorResource{}
}

// UptimeMonitorResource defines the resource implementation.
type UptimeMonitorResource struct {
	client *client.Client
}

// UptimeMonitorResourceModel describes the resource data model.
type UptimeMonitorResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Protocol              types.String `tfsdk:"protocol"`
	HTTPRequest           types.Object `tfsdk:"http_request"`
	TCPRequest            types.Object `tfsdk:"tcp_request"`
	Interval              types.Int64  `tfsdk:"interval"`
	Timeout               types.Int64  `tfsdk:"timeout"`
	IncidentConfirmations types.Int64  `tfsdk:"incident_confirmations"`
	RecoveryConfirmations types.Int64  `tfsdk:"recovery_confirmations"`
	Regions               types.List   `tfsdk:"regions"`
	SuccessAssertions     types.List   `tfsdk:"success_assertions"`
	Paused                types.Bool   `tfsdk:"paused"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

type HTTPRequestModel struct {
	Method          types.String `tfsdk:"method"`
	URL             types.String `tfsdk:"url"`
	TLSSkipVerify   types.Bool   `tfsdk:"tls_skip_verify"`
	Body            types.String `tfsdk:"body"`
	FollowRedirects types.Bool   `tfsdk:"follow_redirects"`
	UserAgentSecret types.String `tfsdk:"user_agent_secret"`
	Headers         types.List   `tfsdk:"headers"`
}

type TCPRequestModel struct {
	Host          types.String `tfsdk:"host"`
	Port          types.String `tfsdk:"port"`
	Connection    types.String `tfsdk:"connection"`
	TLSSkipVerify types.Bool   `tfsdk:"tls_skip_verify"`
}

type RequestHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type SuccessAssertionModel struct {
	Type     types.String `tfsdk:"type"`
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
	Property types.String `tfsdk:"property"`
}

func (r *UptimeMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uptime_monitor"
}

func (r *UptimeMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Phare uptime monitor for HTTP or TCP endpoints.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the monitor",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the monitor (2-30 characters)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 30),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Monitoring protocol: `http` or `tcp`",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("http", "tcp"),
				},
			},
			"http_request": schema.SingleNestedAttribute{
				MarkdownDescription: "HTTP request configuration (required when protocol is `http`)",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						MarkdownDescription: "HTTP method",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("HEAD", "GET", "POST", "PUT", "PATCH", "OPTIONS"),
						},
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "URL to monitor (max 255 characters)",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtMost(255),
						},
					},
					"tls_skip_verify": schema.BoolAttribute{
						MarkdownDescription: "Skip SSL certificate verification",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"body": schema.StringAttribute{
						MarkdownDescription: "Request body for POST, PUT, PATCH (max 500 characters)",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.LengthAtMost(500),
						},
					},
					"follow_redirects": schema.BoolAttribute{
						MarkdownDescription: "Follow HTTP redirects",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
					},
					"user_agent_secret": schema.StringAttribute{
						MarkdownDescription: "Secret value for User-Agent header authentication",
						Optional:            true,
						Sensitive:           true,
					},
					"headers": schema.ListNestedAttribute{
						MarkdownDescription: "Additional HTTP headers (max 10)",
						Optional:            true,
						Validators: []validator.List{
							listvalidator.SizeAtMost(10),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Header name",
									Required:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Header value",
									Required:            true,
								},
							},
						},
					},
				},
			},
			"tcp_request": schema.SingleNestedAttribute{
				MarkdownDescription: "TCP request configuration (required when protocol is `tcp`)",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						MarkdownDescription: "TCP hostname or IP address",
						Required:            true,
					},
					"port": schema.StringAttribute{
						MarkdownDescription: "TCP port",
						Required:            true,
					},
					"connection": schema.StringAttribute{
						MarkdownDescription: "Connection type: `plain` or `tls`",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("plain", "tls"),
						},
					},
					"tls_skip_verify": schema.BoolAttribute{
						MarkdownDescription: "Skip TLS certificate verification",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
			"interval": schema.Int64Attribute{
				MarkdownDescription: "Monitoring interval in seconds (30, 60, 120, 180, 300, 600, 900, 1800, 3600)",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(30, 60, 120, 180, 300, 600, 900, 1800, 3600),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Monitoring timeout in milliseconds (1000-30000)",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 15000, 20000, 25000, 30000),
				},
			},
			"incident_confirmations": schema.Int64Attribute{
				MarkdownDescription: "Number of failed checks required to create an incident (1-5)",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
				},
			},
			"recovery_confirmations": schema.Int64Attribute{
				MarkdownDescription: "Number of successful checks required to resolve an incident (1-5)",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
				},
			},
			"regions": schema.ListAttribute{
				MarkdownDescription: "List of regions where monitoring checks are performed (1-6 regions)",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 6),
					listvalidator.ValueStringsAre(stringvalidator.OneOf(
						"as-jpn-hnd", "as-sgp-sin", "as-tha-bkk",
						"eu-deu-fra", "eu-gbr-lhr", "eu-swe-arn", "ng-nld-ams",
						"na-mex-mex", "na-usa-iad", "na-usa-sea",
						"oc-aus-syd", "sa-bra-gru",
					)),
				},
			},
			"success_assertions": schema.ListNestedAttribute{
				MarkdownDescription: "List of assertions that must be true for check success",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Assertion type: `status_code`, `response_header`, or `response_body`",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("status_code", "response_header", "response_body"),
							},
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Comparison operator",
							Optional:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Expected value",
							Optional:            true,
						},
						"property": schema.StringAttribute{
							MarkdownDescription: "Property name (for response_header type)",
							Optional:            true,
						},
					},
				},
			},
			"paused": schema.BoolAttribute{
				MarkdownDescription: "Whether the monitor is paused",
				Optional:            true,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the monitor was created",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the monitor was last updated",
				Computed:            true,
			},
		},
	}
}

func (r *UptimeMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UptimeMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UptimeMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	monitor, diags := r.terraformToAPIModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating uptime monitor", map[string]any{"name": data.Name.ValueString()})

	// Create monitor via API
	created, err := r.client.CreateMonitor(ctx, monitor)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create monitor", err.Error())
		return
	}

	// Get the created monitor ID
	if created.ID == nil {
		resp.Diagnostics.AddError("Failed to create monitor", "API did not return a monitor ID")
		return
	}

	// Handle pause state if needed
	if !data.Paused.IsNull() && data.Paused.ValueBool() {
		if err := r.client.PauseMonitor(ctx, *created.ID); err != nil {
			resp.Diagnostics.AddError("Failed to pause monitor", err.Error())
			return
		}
	}

	// Read back the monitor to get all fields (created_at, updated_at, etc.)
	fullMonitor, err := r.client.GetMonitor(ctx, *created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created monitor", err.Error())
		return
	}

	// Convert API response back to Terraform model
	diags = r.apiToTerraformModel(ctx, fullMonitor, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UptimeMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UptimeMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading uptime monitor", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid monitor ID", fmt.Sprintf("Failed to parse monitor ID: %s", err.Error()))
		return
	}

	monitor, err := r.client.GetMonitor(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read monitor", err.Error())
		return
	}

	diags := r.apiToTerraformModel(ctx, monitor, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UptimeMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UptimeMonitorResourceModel
	var state UptimeMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	monitor, diags := r.terraformToAPIModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating uptime monitor", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid monitor ID", fmt.Sprintf("Failed to parse monitor ID: %s", err.Error()))
		return
	}

	updated, err := r.client.UpdateMonitor(ctx, id, monitor)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update monitor", err.Error())
		return
	}

	// Handle pause/resume state changes
	statePaused := !state.Paused.IsNull() && state.Paused.ValueBool()
	planPaused := !data.Paused.IsNull() && data.Paused.ValueBool()

	if planPaused && !statePaused {
		if err := r.client.PauseMonitor(ctx, id); err != nil {
			resp.Diagnostics.AddError("Failed to pause monitor", err.Error())
			return
		}
		updated.Paused = &[]bool{true}[0]
	} else if !planPaused && statePaused {
		if err := r.client.ResumeMonitor(ctx, id); err != nil {
			resp.Diagnostics.AddError("Failed to resume monitor", err.Error())
			return
		}
		updated.Paused = &[]bool{false}[0]
	}

	diags = r.apiToTerraformModel(ctx, updated, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UptimeMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UptimeMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting uptime monitor", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid monitor ID", fmt.Sprintf("Failed to parse monitor ID: %s", err.Error()))
		return
	}

	if err := r.client.DeleteMonitor(ctx, id); err != nil {
		resp.Diagnostics.AddError("Failed to delete monitor", err.Error())
		return
	}
}

func (r *UptimeMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions to convert between Terraform and API models will be added in next file
