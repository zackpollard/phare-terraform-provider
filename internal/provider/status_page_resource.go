// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/phare/terraform-provider-phare/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StatusPageResource{}
var _ resource.ResourceWithImportState = &StatusPageResource{}

func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

// StatusPageResource defines the resource implementation.
type StatusPageResource struct {
	client *client.Client
}

// StatusPageResourceModel describes the resource data model.
type StatusPageResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Title               types.String `tfsdk:"title"`
	Description         types.String `tfsdk:"description"`
	SearchEngineIndexed types.Bool   `tfsdk:"search_engine_indexed"`
	WebsiteURL          types.String `tfsdk:"website_url"`
	Subdomain           types.String `tfsdk:"subdomain"`
	Domain              types.String `tfsdk:"domain"`
	Timeframe           types.Int64  `tfsdk:"timeframe"`
	Colors              types.Object `tfsdk:"colors"`
	Components          types.List   `tfsdk:"components"`
	Logo                types.String `tfsdk:"logo"`
	Favicon             types.String `tfsdk:"favicon"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

type StatusPageColorsModel struct {
	Operational         types.String `tfsdk:"operational"`
	DegradedPerformance types.String `tfsdk:"degraded_performance"`
	PartialOutage       types.String `tfsdk:"partial_outage"`
	MajorOutage         types.String `tfsdk:"major_outage"`
	Maintenance         types.String `tfsdk:"maintenance"`
	Empty               types.String `tfsdk:"empty"`
}

type StatusComponentModel struct {
	ComponentableType types.String `tfsdk:"componentable_type"`
	ComponentableID   types.Int64  `tfsdk:"componentable_id"`
}

func (r *StatusPageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *StatusPageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Phare status page for displaying uptime information publicly.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the status page",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Internal name of the status page (2-30 characters)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 30),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Public title displayed on the status page (2-250 characters)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 250),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description shown on the status page (2-250 characters)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 250),
				},
			},
			"search_engine_indexed": schema.BoolAttribute{
				MarkdownDescription: "Whether search engines should index this status page",
				Required:            true,
			},
			"website_url": schema.StringAttribute{
				MarkdownDescription: "URL of the website this status page is for (max 250 characters)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(250),
				},
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: "Subdomain for the status page (e.g., 'status' for status.phare.io, creates {subdomain}.status.phare.io)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 30),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "Custom domain for the status page",
				Optional:            true,
			},
			"timeframe": schema.Int64Attribute{
				MarkdownDescription: "Number of days of history to display (30, 60, or 90)",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(30, 60, 90),
				},
			},
			"colors": schema.SingleNestedAttribute{
				MarkdownDescription: "Color scheme for different status states",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"operational": schema.StringAttribute{
						MarkdownDescription: "Color for operational status (hex color code)",
						Required:            true,
					},
					"degraded_performance": schema.StringAttribute{
						MarkdownDescription: "Color for degraded performance status (hex color code)",
						Required:            true,
					},
					"partial_outage": schema.StringAttribute{
						MarkdownDescription: "Color for partial outage status (hex color code)",
						Required:            true,
					},
					"major_outage": schema.StringAttribute{
						MarkdownDescription: "Color for major outage status (hex color code)",
						Required:            true,
					},
					"maintenance": schema.StringAttribute{
						MarkdownDescription: "Color for maintenance status (hex color code)",
						Required:            true,
					},
					"empty": schema.StringAttribute{
						MarkdownDescription: "Color for empty/unknown status (hex color code)",
						Required:            true,
					},
				},
			},
			"components": schema.ListNestedAttribute{
				MarkdownDescription: "List of monitors to display as components on the status page",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"componentable_type": schema.StringAttribute{
							MarkdownDescription: "Type of component (e.g., 'uptime/monitor')",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("uptime/monitor"),
							},
						},
						"componentable_id": schema.Int64Attribute{
							MarkdownDescription: "ID of the monitor to display",
							Required:            true,
						},
					},
				},
			},
			"logo": schema.StringAttribute{
				MarkdownDescription: "Logo file path or URL (jpeg, png, or svg)",
				Optional:            true,
			},
			"favicon": schema.StringAttribute{
				MarkdownDescription: "Favicon file path or URL (ico, png, or svg)",
				Optional:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the status page was created",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the status page was last updated",
				Computed:            true,
			},
		},
	}
}

func (r *StatusPageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	page, diags := r.terraformToAPIModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating status page", map[string]any{"name": data.Name.ValueString()})

	created, err := r.client.CreateStatusPage(ctx, page)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create status page", err.Error())
		return
	}

	// Get the created page ID
	if created.ID == nil {
		resp.Diagnostics.AddError("Failed to create status page", "API did not return a status page ID")
		return
	}

	// Read back the status page to get all fields
	fullPage, err := r.client.GetStatusPage(ctx, *created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created status page", err.Error())
		return
	}

	diags = r.apiToTerraformModel(ctx, fullPage, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading status page", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid status page ID", fmt.Sprintf("Failed to parse status page ID: %s", err.Error()))
		return
	}

	page, err := r.client.GetStatusPage(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read status page", err.Error())
		return
	}

	diags := r.apiToTerraformModel(ctx, page, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, diags := r.terraformToAPIModel(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating status page", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid status page ID", fmt.Sprintf("Failed to parse status page ID: %s", err.Error()))
		return
	}

	updated, err := r.client.UpdateStatusPage(ctx, id, page)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update status page", err.Error())
		return
	}

	diags = r.apiToTerraformModel(ctx, updated, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StatusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StatusPageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting status page", map[string]any{"id": data.ID.ValueString()})

	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid status page ID", fmt.Sprintf("Failed to parse status page ID: %s", err.Error()))
		return
	}

	if err := r.client.DeleteStatusPage(ctx, id); err != nil {
		resp.Diagnostics.AddError("Failed to delete status page", err.Error())
		return
	}
}

func (r *StatusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions for model conversion will be in a separate file
