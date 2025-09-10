// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"fmt"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAccessManagementSnowflakePolicy_basic(t *testing.T) {
	resourceName := "altr_access_management_snowflake_policy.test"
	connectionID := 19 // Replace with a valid connection ID for testing

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessManagementSnowflakePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementSnowflakePolicyConfigBasic(connectionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-access-management-policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test access management policy"),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.0", fmt.Sprintf("%d", connectionID)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.access.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAccessManagementSnowflakePolicy_update(t *testing.T) {
	resourceName := "altr_access_management_snowflake_policy.test"
	connectionID := 19 // Replace with a valid connection ID for testing

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessManagementSnowflakePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementSnowflakePolicyConfigBasic(connectionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-access-management-policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test access management policy"),
				),
			},
			{
				Config: testAccAccessManagementSnowflakePolicyConfigUpdated(connectionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-access-management-policy-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated access management policy"),
				),
			},
		},
	})
}

func testAccCheckAccessManagementSnowflakePolicyDestroy(s *terraform.State) error {
	// Implement logic to verify the resource has been destroyed
	return nil
}

func testAccAccessManagementSnowflakePolicyConfigBasic(connectionID int) string {
	return fmt.Sprintf(`
resource "altr_access_management_snowflake_policy" "test" {
  name        = "test-access-management-policy"
  description = "Test access management policy"
  connection_ids = [%d]

  rules = [
    {
      actors = [{
        type        = "role"
        identifiers = ["ACCOUNTADMIN"]
        condition   = "equals"
      }],
      objects = [{
        type        = "database"
        identifiers = ["MY_DB"]
        condition   = "equals"
      }],
      access = [{
        name = "read"
      }]
    }
  ]
}
`, connectionID)
}

func testAccAccessManagementSnowflakePolicyConfigUpdated(connectionID int) string {
	return fmt.Sprintf(`
resource "altr_access_management_snowflake_policy" "test" {
  name        = "test-access-management-policy-updated"
  description = "Updated access management policy"
  connection_ids = [%d]

  rules = [
    {
      actors = [{
        type        = "role"
        identifiers = ["ACCOUNTADMIN"]
        condition   = "equals"
      }],
      objects = [{
        type        = "table"
        fully_qualified_identifiers = [{
          database = "MY_DB"
          schema   = "PUBLIC"
          table    = "MY_TABLE"
        }]
        condition = "fully_qualified"
      }],
      access = [{
        name = "write"
      }]
    }
  ]
}
`, connectionID)
}
