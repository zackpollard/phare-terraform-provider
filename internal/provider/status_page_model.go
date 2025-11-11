// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/phare/terraform-provider-phare/internal/client"
)

// terraformToAPIModel converts Terraform model to API client model
func (r *StatusPageResource) terraformToAPIModel(ctx context.Context, data *StatusPageResourceModel) (*client.StatusPage, diag.Diagnostics) {
	var diags diag.Diagnostics

	page := &client.StatusPage{
		Name:                data.Name.ValueString(),
		Title:               data.Title.ValueString(),
		Description:         data.Description.ValueString(),
		SearchEngineIndexed: data.SearchEngineIndexed.ValueBool(),
		WebsiteURL:          data.WebsiteURL.ValueString(),
	}

	if !data.Subdomain.IsNull() {
		page.Subdomain = stringPtr(data.Subdomain.ValueString())
	}
	if !data.Domain.IsNull() {
		page.Domain = stringPtr(data.Domain.ValueString())
	}
	if !data.Timeframe.IsNull() {
		timeframe := int(data.Timeframe.ValueInt64())
		page.Timeframe = &timeframe
	}
	if !data.Logo.IsNull() {
		page.Logo = stringPtr(data.Logo.ValueString())
	}
	if !data.Favicon.IsNull() {
		page.Favicon = stringPtr(data.Favicon.ValueString())
	}

	// Convert colors
	var colors StatusPageColorsModel
	diags.Append(data.Colors.As(ctx, &colors, basetypes.ObjectAsOptions{})...)

	page.Colors = client.StatusPageColors{
		Operational:         colors.Operational.ValueString(),
		DegradedPerformance: colors.DegradedPerformance.ValueString(),
		PartialOutage:       colors.PartialOutage.ValueString(),
		MajorOutage:         colors.MajorOutage.ValueString(),
		Maintenance:         colors.Maintenance.ValueString(),
		Empty:               colors.Empty.ValueString(),
	}

	// Convert components
	var components []StatusComponentModel
	diags.Append(data.Components.ElementsAs(ctx, &components, false)...)

	page.Components = make([]client.StatusComponent, len(components))
	for i, c := range components {
		page.Components[i] = client.StatusComponent{
			ComponentableType: c.ComponentableType.ValueString(),
			ComponentableID:   int(c.ComponentableID.ValueInt64()),
		}
	}

	return page, diags
}

// apiToTerraformModel converts API client model to Terraform model
func (r *StatusPageResource) apiToTerraformModel(ctx context.Context, page *client.StatusPage, data *StatusPageResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if page.ID != nil {
		data.ID = types.StringValue(fmt.Sprintf("%d", *page.ID))
	}
	data.Name = types.StringValue(page.Name)
	data.Title = types.StringValue(page.Title)
	data.Description = types.StringValue(page.Description)
	data.SearchEngineIndexed = types.BoolValue(page.SearchEngineIndexed)
	data.WebsiteURL = types.StringValue(page.WebsiteURL)

	data.Subdomain = types.StringPointerValue(page.Subdomain)
	data.Domain = types.StringPointerValue(page.Domain)

	if page.Timeframe != nil {
		data.Timeframe = types.Int64Value(int64(*page.Timeframe))
	} else {
		data.Timeframe = types.Int64Null()
	}

	data.Logo = types.StringPointerValue(page.Logo)
	data.Favicon = types.StringPointerValue(page.Favicon)

	if page.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*page.CreatedAt)
	}
	if page.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*page.UpdatedAt)
	}

	// Convert colors
	colorsObj, diagObj := types.ObjectValue(
		map[string]attr.Type{
			"operational":          types.StringType,
			"degraded_performance": types.StringType,
			"partial_outage":       types.StringType,
			"major_outage":         types.StringType,
			"maintenance":          types.StringType,
			"empty":                types.StringType,
		},
		map[string]attr.Value{
			"operational":          types.StringValue(page.Colors.Operational),
			"degraded_performance": types.StringValue(page.Colors.DegradedPerformance),
			"partial_outage":       types.StringValue(page.Colors.PartialOutage),
			"major_outage":         types.StringValue(page.Colors.MajorOutage),
			"maintenance":          types.StringValue(page.Colors.Maintenance),
			"empty":                types.StringValue(page.Colors.Empty),
		},
	)
	diags.Append(diagObj...)
	data.Colors = colorsObj

	// Convert components
	componentElements := make([]attr.Value, len(page.Components))
	for i, c := range page.Components {
		componentObj, diagComp := types.ObjectValue(
			map[string]attr.Type{
				"componentable_type": types.StringType,
				"componentable_id":   types.Int64Type,
			},
			map[string]attr.Value{
				"componentable_type": types.StringValue(c.ComponentableType),
				"componentable_id":   types.Int64Value(int64(c.ComponentableID)),
			},
		)
		diags.Append(diagComp...)
		componentElements[i] = componentObj
	}

	componentList, diagList := types.ListValue(
		types.ObjectType{AttrTypes: map[string]attr.Type{
			"componentable_type": types.StringType,
			"componentable_id":   types.Int64Type,
		}},
		componentElements,
	)
	diags.Append(diagList...)
	data.Components = componentList

	return diags
}
