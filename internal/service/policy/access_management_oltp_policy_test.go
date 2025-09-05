package policy_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	// sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"terraform-provider-altr/internal/acctest"
	"terraform-provider-altr/internal/client"
)

func TestAccAccessManagementOLTPPolicyResource_basic(t *testing.T) {
	resourceName := "altr_access_management_oltp_policy.test"

	// Access Management OLTP Policy
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("access_management_oltp_policy", 32)
	description := "Test policy"
	caseSensitivity := "case_sensitive"
	databaseType := "4"
	databaseTypeName := "oracle"
	repoName := fmt.Sprintf("repo_%d", rand.Int())
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
		CheckDestroy:             testAccAccessManagementOLTPPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementOLTPPolicyResourceConfig_basic(
					policyName, description, caseSensitivity, databaseTypeName,
					repoName, ruleType, actorType, actorIdentifier, actorCondition,
					objectType, dbName, databaseType, dbWildcard, schemaName, schemaWildcard,
					tableName, tableWildcard, columnName, columnWildcard,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccAccessManagementOLTPPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test policy"),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "case_sensitive"),
					resource.TestCheckResourceAttr(resourceName, "database_type", "4"),
					resource.TestCheckResourceAttr(resourceName, "database_type_name", "oracle"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "repo_name", repoName),
					resource.TestCheckResourceAttr(resourceName, "rules.0.type", "read"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actors.0.type", "idp_user"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actors.0.identifiers.0", "test@altr.com"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.actors.0.condition", "equals"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.type", "column"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.database.name", "testdb"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.database.wildcard", "false"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.schema.name", "public"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.schema.wildcard", "false"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.table.name", "employees"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.table.wildcard", "false"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.column.name", "salary"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.objects.0.identifiers.0.column.wildcard", "false"),
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

func TestAccAccessManagementOLTPPolicyResource_invalidRules(t *testing.T) {
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("access_management_oltp_policy", 32)
	repoName := fmt.Sprintf("repo_%d", rand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementOLTPPolicyResourceConfig_invalidRules(policyName, repoName),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccAccessManagementOLTPPolicyResource_disappears(t *testing.T) {
	resourceName := "altr_access_management_oltp_policy.test"

	// Access Management OLTP Policy
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("access_management_oltp_policy", 32)
	description := "Test policy"
	caseSensitivity := "case_sensitive"
	databaseType := "4"
	databaseTypeName := "oracle"
	repoName := fmt.Sprintf("repo_%d", rand.Int())
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
		CheckDestroy:             testAccAccessManagementOLTPPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessManagementOLTPPolicyResourceConfig_basic(
					policyName, description, caseSensitivity, databaseTypeName,
					repoName, ruleType, actorType, actorIdentifier, actorCondition,
					objectType, dbName, databaseType, dbWildcard, schemaName, schemaWildcard,
					tableName, tableWildcard, columnName, columnWildcard,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccAccessManagementOLTPPolicyExists(resourceName),
					testAccAccessManagementOLTPPolicyDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAccessManagementOLTPPolicyResource_duplicateIdentifiers(t *testing.T) {
	// Access Management OLTP Policy
	policyName := "policy_duplicate_identifiers"
	description := "Test policy with duplicate identifiers"
	caseSensitivity := "case_sensitive"
	databaseType := "4"
	databaseTypeName := "oracle"
	repoName := fmt.Sprintf("repo_%d", rand.Int())
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
				Config: testAccAccessManagementOLTPPolicyResourceConfig_duplicateIdentifiers(
					policyName, description, caseSensitivity, databaseTypeName,
					repoName, ruleType, actorType, actorIdentifier, actorCondition,
					objectType, dbName, databaseType, dbWildcard, schemaName, schemaWildcard,
					tableName, tableWildcard, columnName, columnWildcard,
				),
				ExpectError: regexp.MustCompile(`All values must be unique`),
			},
		},
	})
}

func testAccAccessManagementOLTPPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Access Management OLTP Policy ID is set")
		}

		policyID := rs.Primary.Attributes["id"]

		// Create a new client for testing
		conn, err := client.NewClient(
			acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
			acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
			acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
			acctest.TestGetEnv("ALTR_BASE_URL", ""),
		)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		binding, err := conn.GetAccessManagementOLTPPolicy(policyID)
		if err != nil {
			return err
		}

		if binding == nil {
			return fmt.Errorf("Access Management OLTP Policy ID not found")
		}

		return nil
	}
}

func testAccAccessManagementOLTPPolicyDestroy(s *terraform.State) error {
	// Create a new client for testing
	conn, err := client.NewClient(
		acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
		acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
		acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
		acctest.TestGetEnv("ALTR_BASE_URL", ""),
	)
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_access_management_oltp_policy" {
			continue
		}

		policyID := rs.Primary.Attributes["id"]

		policy, err := conn.GetAccessManagementOLTPPolicy(policyID)
		if err != nil {
			return err
		}

		if policy != nil {
			return fmt.Errorf("Access Management OLTP Policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAccessManagementOLTPPolicyDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		policyID := rs.Primary.Attributes["id"]

		// Create a new client for testing
		conn, err := client.NewClient(
			acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
			acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
			acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
			acctest.TestGetEnv("ALTR_BASE_URL", ""),
		)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		return conn.DeleteAccessManagementOLTPPolicy(policyID)
	}
}

func testAccAccessManagementOLTPPolicyResourceConfig_basic(
	policyName, description, caseSensitivity, databaseTypeName,
	repoName, ruleType, actorType, actorIdentifier, actorCondition,
	objectType, dbName, databaseType string, dbWildcard bool,
	schemaName string, schemaWildcard bool,
	tableName string, tableWildcard bool,
	columnName string, columnWildcard bool,
) string {
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
        identifiers = [%[9]q],
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
`,
		policyName, description, caseSensitivity, databaseType, databaseTypeName,
		repoName, ruleType, actorType, actorIdentifier, actorCondition, objectType,
		dbName, dbWildcard, schemaName, schemaWildcard, tableName, tableWildcard,
		columnName, columnWildcard,
	)
}

func testAccAccessManagementOLTPPolicyResourceConfig_invalidRules(policyName, repoName string) string {
	return fmt.Sprintf(`
resource "altr_access_management_oltp_policy" "test" {
  name = %[1]q
  description = "Test policy"
  case_sensitivity = "case_sensitive"
  database_type = "4"
  database_type_name = "oracle"
  repo_name = %[2]q
  rules = [{
    type = "fake_rule_type"
    actors = [{
        type = "idp_user"
        identifiers = ["test@altr.com"]
        condition = "equals"
    }]
    objects = [{
        type = "column"
        identifiers = [{
            database = {
                name = "testdb"
                wildcard = false
            }
            schema = {
                name = "public"
                wildcard = false
            }
            table = {
                name = "employees"
                wildcard = false
            }
            column = {
                name = "salary"
                wildcard = false
            }
        }]
    }]
  }]
}
`, policyName, repoName)
}

func testAccAccessManagementOLTPPolicyResourceConfig_duplicateIdentifiers(
	policyName, description, caseSensitivity, databaseTypeName,
	repoName, ruleType, actorType, actorIdentifier, actorCondition,
	objectType, dbName, databaseType string, dbWildcard bool,
	schemaName string, schemaWildcard bool,
	tableName string, tableWildcard bool,
	columnName string, columnWildcard bool,
) string {
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

`,
		policyName, description, caseSensitivity, databaseType, databaseTypeName,
		repoName, ruleType, actorType, actorIdentifier, actorCondition, objectType,
		dbName, dbWildcard, schemaName, schemaWildcard, tableName, tableWildcard,
		columnName, columnWildcard,
	)
}
