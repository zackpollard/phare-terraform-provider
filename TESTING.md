# Testing Guide

This document describes how to run tests for the Phare Terraform Provider.

## Prerequisites

- Go 1.24 or later
- Phare API token (for acceptance tests)
- Access to a Phare organization for testing

## Unit Tests

Unit tests can be run without any special configuration:

```bash
make test
# or
go test -v -cover -timeout=120s -parallel=10 ./...
```

These tests verify basic functionality like client initialization and do not make API calls.

## Acceptance Tests

Acceptance tests actually interact with the Phare API to create, update, and delete real resources. **These tests will create and destroy real resources in your Phare account.**

### Setup

1. Copy the `.env.example` file to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your Phare API token:
   ```bash
   PHARE_API_TOKEN=your_actual_api_token_here
   ```

3. Source the environment file:
   ```bash
   source .env
   ```

### Running Acceptance Tests

To run acceptance tests, set `TF_ACC=1`:

```bash
make testacc
# or
TF_ACC=1 go test -v -cover -timeout 120m ./...
```

### Running Specific Tests

To run a specific acceptance test:

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider -run TestAccUptimeMonitorResource_HTTP
```

### Important Notes

- **Acceptance tests create real resources** that may incur costs or count against your account limits
- Tests use the organization and project specified in your `.env` file (defaults to `zack` org and `tf-integration` project)
- Each test cleans up after itself by deleting resources it creates
- If a test fails mid-execution, you may need to manually clean up orphaned resources
- Rate limits apply - the Phare API allows 100 calls per minute per organization

## Test Coverage

Current test coverage:

- **Unit Tests**: Basic client and validation logic
- **Acceptance Tests**: Full CRUD lifecycle for all resources
  - `phare_uptime_monitor` (HTTP and TCP)
  - `phare_alert_rule`
  - `phare_status_page`
  - `phare_uptime_incident` (data source)

## Debugging Tests

To enable verbose logging during tests:

```bash
TF_LOG=DEBUG TF_ACC=1 go test -v ./internal/provider -run TestAccUptimeMonitorResource_HTTP
```

This will show detailed Terraform and provider logs.

## Writing New Tests

When adding new resources or data sources, follow these patterns:

1. Create a test file named `<resource_name>_test.go` in `internal/provider/`
2. Use the `testAccProtoV6ProviderFactories` for provider setup
3. Include `PreCheck: func() { testAccPreCheck(t) }` to validate environment
4. Test the full lifecycle: Create, Read, Update (if applicable), Delete, Import
5. Use state checks to verify attributes are set correctly

Example test structure:

```go
func TestAccMyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccMyResourceConfig("value1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_my_resource.test",
						tfjsonpath.New("attribute"),
						knownvalue.StringExact("value1"),
					),
				},
			},
			// Import
			{
				ResourceName:      "phare_my_resource.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
```
