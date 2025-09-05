package repo_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"

	"github.com/google/uuid"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"terraform-provider-altr/internal/acctest"
	"terraform-provider-altr/internal/client"
)

func TestAccRepoResource_basic(t *testing.T) {
	resourceName := "altr_repo.test"
	rName := fmt.Sprintf("repo_%d", rand.Int())
	rHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	port := sdkacctest.RandIntRange(1, 65535)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_basic(rName, "Oracle", rHostname, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateId:                        rName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccRepoResource_withDescription(t *testing.T) {
	resourceName := "altr_repo.test"
	rName := fmt.Sprintf("repo_%d", rand.Int())
	rDescription := "Test repository description"
	rHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	port := sdkacctest.RandIntRange(1, 65535)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_withDescription(rName, "Oracle", rHostname, rDescription, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
				),
			},
		},
	})
}

func TestAccRepoResource_update(t *testing.T) {
	resourceName := "altr_repo.test"
	rName := fmt.Sprintf("repo_%d", rand.Int())
	rDescription1 := "Initial description"
	rDescription2 := "Updated description"
	rHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	port := sdkacctest.RandIntRange(1, 65535)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_withDescription(rName, "Oracle", rHostname, rDescription1, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription1),
				),
			},
			{
				Config: testAccRepoResourceConfig_withDescription(rName, "Oracle", rHostname, rDescription2, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription2),
				),
			},
		},
	})
}

func TestAccRepoResource_nameValidation(t *testing.T) {
	rHostname := fmt.Sprintf("%s.example.altr.com", "abc")
	port := sdkacctest.RandIntRange(1, 65535)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRepoResourceConfig_basic("", "Oracle", rHostname, port),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccRepoResource_disappears(t *testing.T) {
	resourceName := "altr_repo.test"
	rName := fmt.Sprintf("repo_%d", rand.Int())
	rHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	port := sdkacctest.RandIntRange(1, 65535)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoResourceConfig_basic(rName, "Oracle", rHostname, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoExists(resourceName),
					testAccCheckRepoDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepoExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("No Repo Name is set")
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

		repo, err := conn.GetRepo(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if repo == nil {
			return fmt.Errorf("Repo not found")
		}

		return nil
	}
}

func testAccCheckRepoDestroy(s *terraform.State) error {
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
		if rs.Type != "altr_repo" {
			continue
		}

		repo, err := conn.GetRepo(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if repo != nil {
			return fmt.Errorf("Repo %s still exists", rs.Primary.Attributes["name"])
		}
	}

	return nil
}

func testAccCheckRepoDisappears(resourceName string) resource.TestCheckFunc {
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

		return conn.DeleteRepo(rs.Primary.Attributes["name"])
	}
}

func testAccRepoResourceConfig_basic(name, repoType, hostname string, port int) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  type     = %[2]q
  hostname = %[3]q
  port     = %[4]d
}
`, name, repoType, hostname, port)
}

func testAccRepoResourceConfig_withDescription(name, repoType, hostname, description string, port int) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name        = %[1]q
  type        = %[2]q
  hostname    = %[3]q
  port        = %[4]d
  description = %[5]q
}
`, name, repoType, hostname, port, description)
}
