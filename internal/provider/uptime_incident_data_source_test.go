// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUptimeIncidentDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccUptimeIncidentDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source attributes exist
					resource.TestCheckResourceAttrSet("data.phare_uptime_incident.test", "id"),
					resource.TestCheckResourceAttrSet("data.phare_uptime_incident.test", "title"),
					resource.TestCheckResourceAttrSet("data.phare_uptime_incident.test", "status"),
					resource.TestCheckResourceAttrSet("data.phare_uptime_incident.test", "incident_at"),
				),
			},
		},
	})
}

func testAccUptimeIncidentDataSourceConfig() string {
	// Using the test incident ID 221114 created for integration testing
	return `
data "phare_uptime_incident" "test" {
  id = "221114"
}
`
}
