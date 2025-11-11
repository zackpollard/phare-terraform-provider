// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/phare/terraform-provider-phare/internal/client"
)

// Ensure PhareProvider satisfies various provider interfaces.
var _ provider.Provider = &PhareProvider{}
var _ provider.ProviderWithFunctions = &PhareProvider{}
var _ provider.ProviderWithEphemeralResources = &PhareProvider{}

// PhareProvider defines the provider implementation.
type PhareProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PhareProviderModel describes the provider data model.
type PhareProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
}

func (p *PhareProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "phare"
	resp.Version = p.version
}

func (p *PhareProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for Phare platform monitoring.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Phare API token for authentication. Can also be set via PHARE_API_TOKEN environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Phare API base URL. Defaults to https://api.phare.io. Can also be set via PHARE_BASE_URL environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *PhareProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PhareProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get API token from config or environment variable
	apiToken := os.Getenv("PHARE_API_TOKEN")
	if !data.APIToken.IsNull() {
		apiToken = data.APIToken.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"API token must be configured either via the provider configuration "+
				"or the PHARE_API_TOKEN environment variable.",
		)
		return
	}

	// Get base URL from config or environment variable, default to production
	baseURL := client.DefaultBaseURL
	if envURL := os.Getenv("PHARE_BASE_URL"); envURL != "" {
		baseURL = envURL
	}
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	tflog.Debug(ctx, "Configuring Phare API client", map[string]any{
		"base_url": baseURL,
	})

	// Create the Phare API client
	phareClient, err := client.NewClient(apiToken, baseURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Phare API Client",
			"An unexpected error occurred when creating the Phare API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = phareClient
	resp.ResourceData = phareClient
}

func (p *PhareProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUptimeMonitorResource,
		NewAlertRuleResource,
		NewStatusPageResource,
	}
}

func (p *PhareProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		// Phare doesn't use ephemeral resources
	}
}

func (p *PhareProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUptimeIncidentDataSource,
	}
}

func (p *PhareProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// Phare doesn't use functions
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PhareProvider{
			version: version,
		}
	}
}
