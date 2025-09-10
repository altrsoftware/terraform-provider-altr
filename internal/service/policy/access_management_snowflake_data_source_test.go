package policy_test

import (
	"fmt"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccessManagementSnowflakePolicyDataSource_basic(t *testing.T) {
	resourceName := "altr_access_management_snowflake_policy.test"
	dataSourceName := "data.altr_access_management_snowflake_policy.test"

	// Test data
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("snowflake_policy", 32)
	description := "Test Snowflake policy"
	connectionID := 19 // Replace with a valid connection ID for testing
	ruleType := "read"
	actorType := "role"
	actorIdentifier := "ACCOUNTADMIN"
	actorCondition := "equals"
	objectType := "database"
	objectIdentifier := "MY_DB"
	objectCondition := "equals"
	accessType := "read"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementSnowflakePolicyResourceAndDataSourceConfig(
					policyName, description, connectionID, ruleType, actorType, actorIdentifier, actorCondition,
					objectType, objectIdentifier, objectCondition, accessType,
				),
				Check: resource.ComposeTestCheckFunc(
					// Check resource attributes
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "connection_ids.0", fmt.Sprintf("%d", connectionID)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.access.#", "1"),

					// Check data source attributes
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_ids.0", dataSourceName, "connection_ids.0"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.actors.0.type", dataSourceName, "rules.0.actors.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.actors.0.identifiers.0", dataSourceName, "rules.0.actors.0.identifiers.0"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.actors.0.condition", dataSourceName, "rules.0.actors.0.condition"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.objects.0.type", dataSourceName, "rules.0.objects.0.type"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.objects.0.identifiers.0", dataSourceName, "rules.0.objects.0.identifiers.0"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.objects.0.condition", dataSourceName, "rules.0.objects.0.condition"),
					resource.TestCheckResourceAttrPair(resourceName, "rules.0.access.0.name", dataSourceName, "rules.0.access.0.name"),
				),
			},
		},
	})
}

func testAccAccessManagementSnowflakePolicyResourceAndDataSourceConfig(
	policyName, description string, connectionID int, ruleType, actorType, actorIdentifier, actorCondition,
	objectType, objectIdentifier, objectCondition, accessType string,
) string {
	return fmt.Sprintf(`
resource "altr_access_management_snowflake_policy" "test" {
  name        = %[1]q
  description = %[2]q
  connection_ids = [%[3]d]

  rules = [
    {
      actors = [{
        type        = %[5]q
        identifiers = [%[6]q]
        condition   = %[7]q
      }],
      objects = [{
        type        = %[8]q
        identifiers = [%[9]q]
        condition   = %[10]q
      }],
      access = [{
        name = %[11]q
      }]
    }
  ]
}

data "altr_access_management_snowflake_policy" "test" {
  id = altr_access_management_snowflake_policy.test.id
}
`, policyName, description, connectionID, ruleType, actorType, actorIdentifier, actorCondition,
		objectType, objectIdentifier, objectCondition, accessType)
}
