// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"fmt"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccImpersonationPolicyDataSource_basic(t *testing.T) {
	resourceName := "altr_impersonation_policy.test"
	dataSourceName := "data.altr_impersonation_policy.test"

	// Test data
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("impersonation_policy", 32)
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo", 32)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImpersonationPolicyResourceAndDataSourceConfig(policyName, repoName),
				Check: resource.ComposeTestCheckFunc(
					// Check resource attributes
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "repo_name", repoName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),

					// Check data source attributes
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", dataSourceName, "repo_name"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "updated_at", dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccImpersonationPolicyResourceAndDataSourceConfig(policyName, repoName string) string {
	return fmt.Sprintf(`
resource "altr_impersonation_policy" "test" {
  name        = %[1]q
  description = "Test impersonation policy"
  repo_name   = %[2]q

  rules = [
    {
      actors = [
        {
          type        = "idp_user"
          identifiers = ["user1@example.com"]
          condition   = "equals"
        }
      ]
      targets = [
        {
          type        = "repo_user"
          identifiers = ["target_user"]
          condition   = "equals"
        }
      ]
    }
  ]
}

data "altr_impersonation_policy" "test" {
  id = altr_impersonation_policy.test.id
}
`, policyName, repoName)
}
