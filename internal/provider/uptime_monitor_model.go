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
func (r *UptimeMonitorResource) terraformToAPIModel(ctx context.Context, data *UptimeMonitorResourceModel) (*client.Monitor, diag.Diagnostics) {
	var diags diag.Diagnostics

	monitor := &client.Monitor{
		Name:                  data.Name.ValueString(),
		Protocol:              data.Protocol.ValueString(),
		Interval:              int(data.Interval.ValueInt64()),
		Timeout:               int(data.Timeout.ValueInt64()),
		IncidentConfirmations: int(data.IncidentConfirmations.ValueInt64()),
		RecoveryConfirmations: int(data.RecoveryConfirmations.ValueInt64()),
	}

	// Convert regions
	var regions []string
	diags.Append(data.Regions.ElementsAs(ctx, &regions, false)...)
	monitor.Regions = regions

	// Convert protocol-specific request
	if data.Protocol.ValueString() == "http" {
		if data.HTTPRequest.IsNull() {
			diags.AddError("Invalid Configuration", "http_request is required when protocol is 'http'")
			return nil, diags
		}

		var httpReq HTTPRequestModel
		diags.Append(data.HTTPRequest.As(ctx, &httpReq, basetypes.ObjectAsOptions{})...)

		monitor.Request = client.MonitorRequest{
			Method:          stringPtr(httpReq.Method.ValueString()),
			URL:             stringPtr(httpReq.URL.ValueString()),
			TLSSkipVerify:   boolPtr(httpReq.TLSSkipVerify.ValueBool()),
			FollowRedirects: boolPtr(httpReq.FollowRedirects.ValueBool()),
		}

		if !httpReq.Body.IsNull() {
			monitor.Request.Body = stringPtr(httpReq.Body.ValueString())
		}
		if !httpReq.UserAgentSecret.IsNull() {
			monitor.Request.UserAgentSecret = stringPtr(httpReq.UserAgentSecret.ValueString())
		}

		// Convert headers
		if !httpReq.Headers.IsNull() {
			var headers []RequestHeaderModel
			diags.Append(httpReq.Headers.ElementsAs(ctx, &headers, false)...)

			monitor.Request.Headers = make([]client.RequestHeader, len(headers))
			for i, h := range headers {
				monitor.Request.Headers[i] = client.RequestHeader{
					Name:  h.Name.ValueString(),
					Value: h.Value.ValueString(),
				}
			}
		}
	} else if data.Protocol.ValueString() == "tcp" {
		if data.TCPRequest.IsNull() {
			diags.AddError("Invalid Configuration", "tcp_request is required when protocol is 'tcp'")
			return nil, diags
		}

		var tcpReq TCPRequestModel
		diags.Append(data.TCPRequest.As(ctx, &tcpReq, basetypes.ObjectAsOptions{})...)

		monitor.Request = client.MonitorRequest{
			Host:          stringPtr(tcpReq.Host.ValueString()),
			Port:          stringPtr(tcpReq.Port.ValueString()),
			Connection:    stringPtr(tcpReq.Connection.ValueString()),
			TLSSkipVerify: boolPtr(tcpReq.TLSSkipVerify.ValueBool()),
		}
	}

	// Convert success assertions
	if !data.SuccessAssertions.IsNull() {
		var assertions []SuccessAssertionModel
		diags.Append(data.SuccessAssertions.ElementsAs(ctx, &assertions, false)...)

		monitor.SuccessAssertions = make([]client.SuccessAssertion, len(assertions))
		for i, a := range assertions {
			assertion := client.SuccessAssertion{
				Type: a.Type.ValueString(),
			}
			if !a.Operator.IsNull() {
				assertion.Operator = stringPtr(a.Operator.ValueString())
			}
			if !a.Value.IsNull() {
				assertion.Value = stringPtr(a.Value.ValueString())
			}
			if !a.Property.IsNull() {
				assertion.Property = stringPtr(a.Property.ValueString())
			}
			monitor.SuccessAssertions[i] = assertion
		}
	}

	return monitor, diags
}

// apiToTerraformModel converts API client model to Terraform model
func (r *UptimeMonitorResource) apiToTerraformModel(ctx context.Context, monitor *client.Monitor, data *UptimeMonitorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if monitor.ID != nil {
		data.ID = types.StringValue(fmt.Sprintf("%d", *monitor.ID))
	}
	data.Name = types.StringValue(monitor.Name)
	data.Protocol = types.StringValue(monitor.Protocol)
	data.Interval = types.Int64Value(int64(monitor.Interval))
	data.Timeout = types.Int64Value(int64(monitor.Timeout))
	data.IncidentConfirmations = types.Int64Value(int64(monitor.IncidentConfirmations))
	data.RecoveryConfirmations = types.Int64Value(int64(monitor.RecoveryConfirmations))

	if monitor.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*monitor.CreatedAt)
	}
	if monitor.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*monitor.UpdatedAt)
	}
	if monitor.Paused != nil {
		data.Paused = types.BoolValue(*monitor.Paused)
	} else {
		data.Paused = types.BoolValue(false)
	}

	// Convert regions
	regionElements := make([]attr.Value, len(monitor.Regions))
	for i, r := range monitor.Regions {
		regionElements[i] = types.StringValue(r)
	}
	regionList, diagList := types.ListValue(types.StringType, regionElements)
	diags.Append(diagList...)
	data.Regions = regionList

	// Convert protocol-specific request
	if monitor.Protocol == "http" {
		httpReq := HTTPRequestModel{
			Method:          types.StringPointerValue(monitor.Request.Method),
			URL:             types.StringPointerValue(monitor.Request.URL),
			TLSSkipVerify:   types.BoolPointerValue(monitor.Request.TLSSkipVerify),
			FollowRedirects: types.BoolPointerValue(monitor.Request.FollowRedirects),
			Body:            types.StringPointerValue(monitor.Request.Body),
			UserAgentSecret: types.StringPointerValue(monitor.Request.UserAgentSecret),
		}

		// Convert headers
		if len(monitor.Request.Headers) > 0 {
			headerElements := make([]attr.Value, len(monitor.Request.Headers))
			for i, h := range monitor.Request.Headers {
				headerObj, diagObj := types.ObjectValue(
					map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
					map[string]attr.Value{
						"name":  types.StringValue(h.Name),
						"value": types.StringValue(h.Value),
					},
				)
				diags.Append(diagObj...)
				headerElements[i] = headerObj
			}
			headerList, diagList := types.ListValue(
				types.ObjectType{AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				}},
				headerElements,
			)
			diags.Append(diagList...)
			httpReq.Headers = headerList
		} else {
			httpReq.Headers = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			}})
		}

		httpObj, diagObj := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"method":            types.StringType,
			"url":               types.StringType,
			"tls_skip_verify":   types.BoolType,
			"body":              types.StringType,
			"follow_redirects":  types.BoolType,
			"user_agent_secret": types.StringType,
			"headers": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			}}},
		}, httpReq)
		diags.Append(diagObj...)
		data.HTTPRequest = httpObj
		data.TCPRequest = types.ObjectNull(map[string]attr.Type{
			"host":            types.StringType,
			"port":            types.StringType,
			"connection":      types.StringType,
			"tls_skip_verify": types.BoolType,
		})
	} else if monitor.Protocol == "tcp" {
		tcpReq := TCPRequestModel{
			Host:          types.StringPointerValue(monitor.Request.Host),
			Port:          types.StringPointerValue(monitor.Request.Port),
			Connection:    types.StringPointerValue(monitor.Request.Connection),
			TLSSkipVerify: types.BoolPointerValue(monitor.Request.TLSSkipVerify),
		}

		tcpObj, diagObj := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"host":            types.StringType,
			"port":            types.StringType,
			"connection":      types.StringType,
			"tls_skip_verify": types.BoolType,
		}, tcpReq)
		diags.Append(diagObj...)
		data.TCPRequest = tcpObj
		data.HTTPRequest = types.ObjectNull(map[string]attr.Type{
			"method":            types.StringType,
			"url":               types.StringType,
			"tls_skip_verify":   types.BoolType,
			"body":              types.StringType,
			"follow_redirects":  types.BoolType,
			"user_agent_secret": types.StringType,
			"headers": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			}}},
		})
	}

	// Convert success assertions
	if len(monitor.SuccessAssertions) > 0 {
		assertionElements := make([]attr.Value, len(monitor.SuccessAssertions))
		for i, a := range monitor.SuccessAssertions {
			assertionObj, diagObj := types.ObjectValue(
				map[string]attr.Type{
					"type":     types.StringType,
					"operator": types.StringType,
					"value":    types.StringType,
					"property": types.StringType,
				},
				map[string]attr.Value{
					"type":     types.StringValue(a.Type),
					"operator": types.StringPointerValue(a.Operator),
					"value":    types.StringPointerValue(a.Value),
					"property": types.StringPointerValue(a.Property),
				},
			)
			diags.Append(diagObj...)
			assertionElements[i] = assertionObj
		}
		assertionList, diagList := types.ListValue(
			types.ObjectType{AttrTypes: map[string]attr.Type{
				"type":     types.StringType,
				"operator": types.StringType,
				"value":    types.StringType,
				"property": types.StringType,
			}},
			assertionElements,
		)
		diags.Append(diagList...)
		data.SuccessAssertions = assertionList
	} else {
		data.SuccessAssertions = types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":     types.StringType,
			"operator": types.StringType,
			"value":    types.StringType,
			"property": types.StringType,
		}})
	}

	return diags
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
