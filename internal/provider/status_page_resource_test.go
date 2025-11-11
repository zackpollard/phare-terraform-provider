// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccStatusPageResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccStatusPageResourceConfig("Test Status Page", "Test Status"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_status_page.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("Test Status Page"),
					),
					statecheck.ExpectKnownValue(
						"phare_status_page.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Status"),
					),
					statecheck.ExpectKnownValue(
						"phare_status_page.test",
						tfjsonpath.New("search_engine_indexed"),
						knownvalue.Bool(false),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "phare_status_page.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccStatusPageResourceConfig("Test Status Page", "Updated Status"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_status_page.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Updated Status"),
					),
				},
			},
		},
	})
}

func testAccStatusPageResourceConfig(name, title string) string {
	return fmt.Sprintf(`
resource "phare_uptime_monitor" "status_test" {
  name     = "Monitor for Status Page Test"
  protocol = "http"

  http_request = {
    method = "GET"
    url    = "https://immich.app"
  }

  interval                = 60
  timeout                 = 5000
  incident_confirmations  = 1
  recovery_confirmations  = 1
  regions                 = ["na-usa-iad"]

  success_assertions = [
    {
      type     = "status_code"
      operator = "in"
      value    = "2xx"
    }
  ]
}

resource "phare_status_page" "test" {
  name                  = %[1]q
  title                 = %[2]q
  description           = "Test status page description"
  search_engine_indexed = false
  website_url           = "https://example.com"
  subdomain             = "tf-test-status"
  timeframe             = 90

  colors = {
    operational          = "#16a34a"
    degraded_performance = "#fbbf24"
    partial_outage       = "#f59e0b"
    major_outage         = "#ef4444"
    maintenance          = "#6366f1"
    empty                = "#d3d3d3"
  }

  components = [
    {
      componentable_type = "uptime/monitor"
      componentable_id   = tonumber(phare_uptime_monitor.status_test.id)
    }
  ]
}
`, name, title)
}
