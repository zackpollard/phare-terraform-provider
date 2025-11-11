# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for Phare (https://docs.phare.io), a platform monitoring solution. The provider is built using the Terraform Plugin Framework and follows HashiCorp's provider scaffolding structure.

**Key Documentation:**
- Phare API: https://docs.phare.io/api-reference/introduction
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/plugin/framework

## Build and Test Commands

**Building:**
```bash
make build          # Build the provider
make install        # Build and install to $GOPATH/bin
go install          # Alternative: direct install
```

**Testing:**
```bash
make test           # Run unit tests (120s timeout, parallel=10)
make testacc        # Run acceptance tests (requires TF_ACC=1, 120m timeout)
go test -v -cover -timeout=120s -parallel=10 ./...  # Direct unit test
TF_ACC=1 go test -v -cover -timeout 120m ./...      # Direct acceptance test
```

**Code Quality:**
```bash
make fmt            # Format code with gofmt
make lint           # Run golangci-lint
make generate       # Generate documentation and copyright headers
make                # Run default: fmt, lint, install, generate
```

**Documentation Generation:**
```bash
cd tools && go generate ./...   # Generates provider docs, formats examples, adds copyright headers
```

## Architecture

### Provider Structure

**Entry Point:** `main.go`
- Sets up the provider server with address `registry.terraform.io/phare/phare`
- Passes version string to provider (set by goreleaser, defaults to "dev")
- Supports debug mode via `-debug` flag for debugger attachment

**Provider Core:** `internal/provider/provider.go`
- `PhareProvider` struct implements `provider.Provider`, `ProviderWithFunctions`, and `ProviderWithEphemeralResources`
- `Configure()` method sets up shared Phare API client from `internal/client` package, passed to resources/datasources via `ProviderData`
- Provider schema defines configuration attributes (`api_token`, `base_url`)
- Registers resources via `Resources()`, data sources via `DataSources()`, ephemeral resources via `EphemeralResources()`, and functions via `Functions()`

### Resource Implementation Pattern

Resources follow a standard pattern (see `internal/provider/uptime_monitor_resource.go` or `internal/provider/alert_rule_resource.go`):

1. **Struct Definition:**
   - Resource struct holds shared client from provider
   - Model struct with `tfsdk` tags maps to Terraform schema

2. **Required Methods:**
   - `Metadata()` - sets resource type name (provider name + suffix)
   - `Schema()` - defines attributes, descriptions, validators, plan modifiers
   - `Configure()` - receives client from provider's `Configure()`
   - `Create()`, `Read()`, `Update()`, `Delete()` - CRUD operations
   - `ImportState()` - enables `terraform import` (typically uses `resource.ImportStatePassthroughID`)

3. **Key Patterns:**
   - Use `types.String`, `types.Int64`, etc. for nullable Terraform values
   - Plan modifiers (e.g., `stringplanmodifier.UseStateForUnknown()`) control attribute behavior
   - Defaults via `stringdefault.StaticString()`, etc.
   - Access provider client via type assertion on `req.ProviderData.(*http.Client)`
   - Log with `tflog` package
   - Return errors via `resp.Diagnostics.AddError()`

### Testing Pattern

**Unit Tests:** `internal/provider/*_test.go`
- Use `resource.Test()` from `terraform-plugin-testing`
- Define test steps with HCL configs
- Use `statecheck.ExpectKnownValue()` with `knownvalue` matchers
- `testAccProtoV6ProviderFactories` provides test provider factory
- `testAccPreCheck()` validates test environment

**Acceptance Tests:**
- Set `TF_ACC=1` environment variable
- Actually create/destroy resources (may cost money)
- Test import with `ImportState: true, ImportStateVerify: true`

## Module Structure

```
internal/provider/   # All provider code (resources, datasources, functions, ephemeral resources)
examples/            # Example Terraform configurations
docs/                # Generated documentation (do not edit manually)
tools/               # Build tools (copywrite, tfplugindocs)
```

## Key Dependencies

- `github.com/hashicorp/terraform-plugin-framework` - Core framework
- `github.com/hashicorp/terraform-plugin-testing` - Testing utilities
- `github.com/hashicorp/terraform-plugin-log` - Logging (tflog)
- `github.com/hashicorp/terraform-plugin-docs` - Doc generation

## Development Workflow

1. **Adding a Resource:**
   - Create `internal/provider/[resource_name]_resource.go`
   - Implement required interfaces and CRUD methods
   - Register in provider's `Resources()` method
   - Create test file `internal/provider/[resource_name]_resource_test.go`
   - Add example in `examples/resources/[provider]_[resource]/`
   - Run `make generate` to create docs

2. **Adding a Data Source:**
   - Similar pattern to resources but implement `datasource.DataSource`
   - Only needs `Read()` method
   - Register in provider's `DataSources()` method

3. **Updating Provider Configuration:**
   - Modify `ScaffoldingProviderModel` struct
   - Update schema in provider's `Schema()` method
   - Use values in provider's `Configure()` method

## Phare Provider Specific Notes

**Implementation Status:** ✅ Complete

The provider has been fully migrated from the terraform-provider-scaffolding-framework template. All scaffolding references have been replaced with Phare-specific implementations.

**Completed:**
- ✅ Implemented resources from Phare API (https://docs.phare.io/api-reference/introduction)
  - phare_uptime_monitor (HTTP and TCP)
  - phare_alert_rule
  - phare_status_page
- ✅ Implemented data sources
  - phare_uptime_incident
- ✅ Created comprehensive unit tests
- ✅ Created acceptance tests that call actual Phare API
- ✅ Updated provider configuration to accept Phare API credentials
- ✅ Replaced all "scaffolding" references with "phare"
- ✅ Updated registry address to `registry.terraform.io/phare/phare`
- ✅ Created proper Phare API client in `internal/client/`
- ✅ All tests passing (5/5 acceptance tests)