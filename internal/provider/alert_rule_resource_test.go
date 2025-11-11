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

func TestAccAlertRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAlertRuleResourceConfig(64493, 0),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_alert_rule.test",
						tfjsonpath.New("event"),
						knownvalue.StringExact("uptime.incident.created"),
					),
					statecheck.ExpectKnownValue(
						"phare_alert_rule.test",
						tfjsonpath.New("integration_id"),
						knownvalue.Int64Exact(64493),
					),
					statecheck.ExpectKnownValue(
						"phare_alert_rule.test",
						tfjsonpath.New("rate_limit"),
						knownvalue.Int64Exact(0),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "phare_alert_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				// event_settings is not returned by the API, so we ignore it during import
				ImportStateVerifyIgnore: []string{"event_settings"},
			},
			// Update and Read testing
			{
				Config: testAccAlertRuleResourceConfig(64493, 5),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"phare_alert_rule.test",
						tfjsonpath.New("rate_limit"),
						knownvalue.Int64Exact(5),
					),
				},
			},
		},
	})
}

func testAccAlertRuleResourceConfig(integrationID, rateLimit int) string {
	return fmt.Sprintf(`
resource "phare_alert_rule" "test" {
  event          = "uptime.incident.created"
  integration_id = %[1]d
  rate_limit     = %[2]d

  event_settings = {
    type = "all"
  }
}
`, integrationID, rateLimit)
}
