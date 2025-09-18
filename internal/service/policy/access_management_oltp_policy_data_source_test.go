// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccessManagementOLTPPolicyDataSource_basic(t *testing.T) {
	resourceName := "altr_access_management_oltp_policy.test"
	dataSourceName := "data.altr_access_management_oltp_policy.test"

	// Test data
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("access_management_oltp_policy", 32)
	description := "Test policy"
	caseSensitivity := "case_sensitive"
	databaseType := "4"
	databaseTypeName := "oracle"
	repoName := fmt.Sprintf("repo_%d", rand.Intn(1000000))
	//repoName := fmt.Sprintf("repo_%d", rand.Intn(1000000))
	ruleType := "read"
	actorType := "idp_user"
	actorIdentifier := "test@altr.com"
	actorCondition := "equals"
	objectType := "column"
	dbName := "testdb"
	dbWildcard := false
	schemaName := "public"
	schemaWildcard := false
	tableName := "employees"
	tableWildcard := false
	columnName := "salary"
	columnWildcard := false

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementOLTPPolicyDataSourceConfig(policyName, description, caseSensitivity, databaseTypeName,
					repoName, ruleType, actorType, actorIdentifier, actorCondition,
					objectType, dbName, databaseType, dbWildcard, schemaName, schemaWildcard,
					tableName, tableWildcard, columnName, columnWildcard),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", dataSourceName, "repo_name"),
					resource.TestCheckResourceAttrPair(resourceName, "case_sensitivity", dataSourceName, "case_sensitivity"),
					resource.TestCheckResourceAttrPair(resourceName, "database_type", dataSourceName, "database_type"),
					resource.TestCheckResourceAttrPair(resourceName, "database_type_name", dataSourceName, "database_type_name"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "updated_at", dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func testAccAccessManagementOLTPPolicyDataSourceConfig(policyName, description, caseSensitivity, databaseTypeName,
	repoName, ruleType, actorType, actorIdentifier, actorCondition,
	objectType, dbName, databaseType string, dbWildcard bool,
	schemaName string, schemaWildcard bool,
	tableName string, tableWildcard bool,
	columnName string, columnWildcard bool) string {
	return fmt.Sprintf(`
resource "altr_access_management_oltp_policy" "test" {
  name = %[1]q
  description = %[2]q
  case_sensitivity = %[3]q
  database_type = %[4]q
  database_type_name = %[5]q
  repo_name = %[6]q
  rules = [{
    type = %[7]q
    actors = [{
        type = %[8]q,
        identifiers = [%[9]q, %[9]q], // Duplicate identifiers
        condition = %[10]q,
    }],
    objects = [{
        type = %[11]q,
        identifiers = [{
            database = {
                name = %[12]q
                wildcard = %[13]t
            }
            schema = {
                name = %[14]q
                wildcard = %[15]t
            }
            table = {
                name = %[16]q
                wildcard = %[17]t
            }
            column = {
                name = %[18]q
                wildcard = %[19]t
            }
        }]
    }],
  }]
}

data "altr_access_management_oltp_policy" "test" {
  id = altr_access_management_oltp_policy.test.id
}
`, policyName, description, caseSensitivity, databaseType, databaseTypeName,
		repoName, ruleType, actorType, actorIdentifier, actorCondition, objectType,
		dbName, dbWildcard, schemaName, schemaWildcard, tableName, tableWildcard,
		columnName, columnWildcard,
	)
}
