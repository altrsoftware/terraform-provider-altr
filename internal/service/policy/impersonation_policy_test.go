package policy_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"terraform-provider-altr/internal/acctest"
	"terraform-provider-altr/internal/client"
)

func TestAccImpersonationPolicyResource_basic(t *testing.T) {
	resourceName := "altr_impersonation_policy.test"
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("impersonation_policy", 32)
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo", 32)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImpersonationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImpersonationPolicyResourceConfig_basic(policyName, repoName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImpersonationPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "repo_name", repoName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
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

func TestAccImpersonationPolicyResource_invalidRules(t *testing.T) {
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("impersonation_policy", 32)
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo", 32)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccImpersonationPolicyResourceConfig_invalidRules(policyName, repoName),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccImpersonationPolicyResource_disappears(t *testing.T) {
	resourceName := "altr_impersonation_policy.test"
	policyName := acctest.RandomWithPrefixUnderscoreMaxLength("impersonation_policy", 32)
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo", 32)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckImpersonationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImpersonationPolicyResourceConfig_basic(policyName, repoName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImpersonationPolicyExists(resourceName),
					testAccCheckImpersonationPolicyDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckImpersonationPolicyExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Impersonation Policy ID is set")
		}

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

		policy, err := conn.GetImpersonationPolicy(rs.Primary.ID)
		if err != nil {
			return err
		}

		if policy == nil {
			return fmt.Errorf("Impersonation Policy not found")
		}

		return nil
	}
}

func testAccCheckImpersonationPolicyDestroy(s *terraform.State) error {
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
		if rs.Type != "altr_impersonation_policy" {
			continue
		}

		policy, err := conn.GetImpersonationPolicy(rs.Primary.ID)
		if err != nil {
			return err
		}

		if policy != nil {
			return fmt.Errorf("Impersonation Policy %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckImpersonationPolicyDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

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

		return conn.DeleteImpersonationPolicy(rs.Primary.ID)
	}
}

func testAccImpersonationPolicyResourceConfig_basic(policyName, repoName string) string {
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
`, policyName, repoName)
}

func testAccImpersonationPolicyResourceConfig_invalidRules(policyName, repoName string) string {
	return fmt.Sprintf(`
resource "altr_impersonation_policy" "test" {
  name        = %[1]q
  description = "Test impersonation policy"
  repo_name   = %[2]q

  rules = [
    {
      actors = [
        {
          type        = "invalid_type"
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
`, policyName, repoName)
}
