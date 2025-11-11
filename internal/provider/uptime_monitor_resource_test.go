// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccUptimeMonitorResource_HTTP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUptimeMonitorResourceConfig_HTTP("https://immich.app", 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_uptime_monitor.test",
						tfjsonpath.New("protocol"),
						knownvalue.StringExact("http"),
					),
					statecheck.ExpectKnownValue(
						"phare_uptime_monitor.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "phare_uptime_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccUptimeMonitorResourceConfig_HTTP("https://immich.app", 120),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_uptime_monitor.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
				},
			},
			// Delete testing automatically happens at the end
		},
	})
}

func TestAccUptimeMonitorResource_TCP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUptimeMonitorResourceConfig_TCP("8.8.8.8", "53"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_uptime_monitor.test",
						tfjsonpath.New("protocol"),
						knownvalue.StringExact("tcp"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "phare_uptime_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUptimeMonitorResourceConfig_HTTP(url string, interval int) string {
	timestamp := time.Now().Unix() % 10000 // Last 4 digits
	return fmt.Sprintf(`
resource "phare_uptime_monitor" "test" {
  name     = "TF HTTP Test %[3]d"
  protocol = "http"

  http_request = {
    method = "GET"
    url    = %[1]q
  }

  interval                = %[2]d
  timeout                 = 5000
  incident_confirmations  = 1
  recovery_confirmations  = 1
  regions                 = ["na-usa-iad", "eu-deu-fra"]

  success_assertions = [
    {
      type     = "status_code"
      operator = "in"
      value    = "2xx"
    }
  ]
}
`, url, interval, timestamp)
}

func testAccUptimeMonitorResourceConfig_TCP(host, port string) string {
	return fmt.Sprintf(`
resource "phare_uptime_monitor" "test" {
  name     = "Test TCP Monitor"
  protocol = "tcp"

  tcp_request = {
    host       = %[1]q
    port       = %[2]q
    connection = "plain"
  }

  interval                = 60
  timeout                 = 5000
  incident_confirmations  = 1
  recovery_confirmations  = 1
  regions                 = ["na-usa-iad"]
}
`, host, port)
}
