# Terraform Provider for Phare

The official [Terraform](https://www.terraform.io) provider for [Phare](https://phare.io), a platform monitoring solution.

This provider is built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and allows you to manage Phare resources including:

- **Uptime Monitors** - HTTP and TCP monitors for tracking service availability
- **Alert Rules** - Configure notifications for platform events
- **Status Pages** - Manage public incident communication
- **Incidents** - Query incident data (data source)

Documentation for the Phare API can be found at [https://docs.phare.io/api-reference/introduction](https://docs.phare.io/api-reference/introduction).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Installation

This provider will be available on the [Terraform Registry](https://registry.terraform.io/). To use it, add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    phare = {
      source = "phare/phare"
      version = "~> 1.0"
    }
  }
}

provider "phare" {
  api_token = var.phare_api_token # or set PHARE_API_TOKEN environment variable
}
```

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the Provider

### Authentication

The provider requires a Phare API token for authentication. You can provide it in two ways:

1. Via provider configuration:
```hcl
provider "phare" {
  api_token = "phare_your_api_token_here"
}
```

2. Via environment variable (recommended):
```shell
export PHARE_API_TOKEN="phare_your_api_token_here"
terraform plan
```

### Example Usage

Create an uptime monitor:

```hcl
resource "phare_uptime_monitor" "example" {
  name        = "My Website"
  type        = "http"
  url         = "https://example.com"
  interval    = 60
  timeout     = 30
  num_retries = 3

  http_settings = {
    method = "GET"
  }
}
```

Create an alert rule:

```hcl
resource "phare_alert_rule" "example" {
  event          = "uptime.incident.created"
  integration_id = 12345
  rate_limit     = 5

  event_settings = {
    type = "all"
  }
}
```

For more examples, see the `examples/` directory.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

### Building

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```shell
make build
```

### Testing

Run unit tests:

```shell
make test
```

Run acceptance tests (requires `PHARE_API_TOKEN` environment variable):

```shell
export PHARE_API_TOKEN="your_token_here"
make testacc
```

**Note:** Acceptance tests create real resources using your Phare account.

### Documentation

To generate or update documentation, run:

```shell
make generate
```

Documentation is automatically generated from the provider schema and example files in the `examples/` directory.

## Contributing

Contributions are welcome! Please see [TESTING.md](TESTING.md) for information on running tests.
