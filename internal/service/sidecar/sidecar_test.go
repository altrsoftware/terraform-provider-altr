package sidecar_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"terraform-provider-altr/internal/acctest"
	"terraform-provider-altr/internal/client"
)

var pubKeyExample1 = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp/TpUtMCmMLzQ+vTzuel
XudSzqzsgjEj7dWrpNgY+fwo8r6oVx19pDbeNATlCMrmQM942aGmL2kdBhhPrZuC
0ImfaLjQxqHgXrNEqis7C+mlm9B0NK3LXp8x+FvIuE92z0L9fw/kH9bsicDCh4QQ
W0Amk6rR8Gc2qWwJfwSz+6H/fCfPd9fPsAQ+oeaB4yf8lqeQcbdbCTmGWkpWQ1I6
5yzMizbmwrj7/G44drPSpKahrY3OpUNSiweVCrQ9bxd3Eu/UIgr2CIvG1bYtt1b9
RF5rhiJNZR6dwdftOeHzoQiLHL4r/VsJ7PpMycjdtaVgPnv30JMnkAQTP6m9c+/H
+wIDAQAB
-----END PUBLIC KEY-----`

var pubKeyExample2 = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzEyEtWdjYF8AnUnbrczr
FZBarvdu/urMLvGQ6x9D6+AN8RrdHj5GV5+Bp17MHaceZTLlvra1x2LSaXlQhnro
CV6aP3/T35jIcZzzGWCvuaJan/N4IKy+cLlgmFr+8NQltBM6v/tRDVfQgGDNa4GA
GSHMGuBrYuLjwDdq4cai7BDkNveI6eHDol+uMbrrIjT4Ku1DaqTsgACk0xLJw44u
PQEwv5YtJRuXT6cFuGiSPRJcjJFFiJ0qtWWCoFT6o7ov6cuM4tGqfNYW6z/xFUpE
AdJ3tVlGq0almLUC3+dpY4plkMbI4Dphv/6RaSVYPFklAnnjQzI4eby9cTNUBdeM
WQIDAQAB
-----END PUBLIC KEY-----`

func TestAccSidecarResource_basic(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_basic(rName, rHostname, pubKeyExample1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "listener_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "listener_repo_binding_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "unsupported_query_bypass", "false"),
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

func TestAccSidecarResource_withDescription(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	rDescription := "Test sidecar description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_withDescription(rName, rHostname, pubKeyExample1, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
				),
			},
		},
	})
}

func TestAccSidecarResource_withPublicKeys(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	publicKey1 := pubKeyExample1
	publicKey2 := pubKeyExample2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_withPublicKeys(rName, rHostname, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttr(resourceName, "public_key_1", publicKey1),
					resource.TestCheckResourceAttr(resourceName, "public_key_2", publicKey2),
				),
			},
		},
	})
}

func TestAccSidecarResource_withQueryBypass(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_withQueryBypass(rName, rHostname, pubKeyExample1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttr(resourceName, "unsupported_query_bypass", "true"),
				),
			},
		},
	})
}

func TestAccSidecarResource_update(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	rDescription1 := "Initial description"
	rDescription2 := "Updated description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_withDescription(rName, rHostname, pubKeyExample1, rDescription1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription1),
				),
			},
			{
				Config: testAccSidecarResourceConfig_withDescription(rName, rHostname, pubKeyExample1, rDescription2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription2),
				),
			},
		},
	})
}

func TestAccSidecarResource_complete(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	rDescription := "Complete sidecar configuration"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_complete(rName, rHostname, rDescription, pubKeyExample1, pubKeyExample2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
					resource.TestCheckResourceAttr(resourceName, "public_key_1", pubKeyExample1),
					resource.TestCheckResourceAttr(resourceName, "public_key_2", pubKeyExample2),
					resource.TestCheckResourceAttr(resourceName, "unsupported_query_bypass", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "data_plane_url"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccSidecarResource_nameValidation(t *testing.T) {
	rHostname := "example.altr.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSidecarResourceConfig_basic("", rHostname, pubKeyExample1),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccSidecarResource_hostnameValidation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSidecarResourceConfig_basic(rName, "", pubKeyExample1),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccSidecarResource_disappears(t *testing.T) {
	resourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarResourceConfig_basic(rName, rHostname, pubKeyExample1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarExists(resourceName),
					testAccCheckSidecarDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSidecarExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Sidecar ID is set")
		}

		// Create a new client for testing
		// You'll need to set up test credentials or use environment variables
		conn, err := client.NewClient(
			testGetEnv("ALTR_ORG_ID", "test-org"),
			testGetEnv("ALTR_API_KEY", "test-key"),
			testGetEnv("ALTR_SECRET", "test-secret"),
			testGetEnv("ALTR_BASE_URL", ""),
		)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		sidecar, err := conn.GetSidecar(rs.Primary.ID)
		if err != nil {
			return err
		}

		if sidecar == nil {
			return fmt.Errorf("Sidecar not found")
		}

		return nil
	}
}

func testAccCheckSidecarDestroy(s *terraform.State) error {
	// Create a new client for testing
	conn, err := client.NewClient(
		testGetEnv("ALTR_ORG_ID", "test-org"),
		testGetEnv("ALTR_API_KEY", "test-key"),
		testGetEnv("ALTR_SECRET", "test-secret"),
		testGetEnv("ALTR_BASE_URL", ""),
	)
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_sidecar" {
			continue
		}

		sidecar, err := conn.GetSidecar(rs.Primary.ID)
		if err != nil {
			return err
		}

		if sidecar != nil {
			return fmt.Errorf("Sidecar %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckSidecarDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Create a new client for testing
		conn, err := client.NewClient(
			testGetEnv("ALTR_ORG_ID", "test-org"),
			testGetEnv("ALTR_API_KEY", "test-key"),
			testGetEnv("ALTR_SECRET", "test-secret"),
			testGetEnv("ALTR_BASE_URL", ""),
		)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		return conn.DeleteSidecar(rs.Primary.ID)
	}
}

func testAccSidecarResourceConfig_basic(name, hostname, publicKey string) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}
`, name, hostname, publicKey)
}

func testAccSidecarResourceConfig_withDescription(name, hostname, publicKey, description string) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
  description  = %[4]q
}
`, name, hostname, publicKey, description)
}

func testAccSidecarResourceConfig_withPublicKeys(name, hostname, publicKey1, publicKey2 string) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
  public_key_2 = %[4]q
}
`, name, hostname, publicKey1, publicKey2)
}

func testAccSidecarResourceConfig_withQueryBypass(name, hostname, publicKey string, queryBypass bool) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name                       = %[1]q
  hostname                   = %[2]q
  public_key_1               = %[3]q
  unsupported_query_bypass   = %[4]t
}
`, name, hostname, publicKey, queryBypass)
}

func testAccSidecarResourceConfig_complete(name, hostname, description, publicKey1, publicKey2 string, queryBypass bool) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name                       = %[1]q
  hostname                   = %[2]q
  description                = %[3]q
  public_key_1               = %[4]q
  public_key_2               = %[5]q
  unsupported_query_bypass   = %[6]t
}
`, name, hostname, description, publicKey1, publicKey2, queryBypass)
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if testGetEnv("ALTR_ORG_ID", "") == "" && testGetEnv("TF_ACC", "") == "1" {
		t.Fatal("ALTR_ORG_ID must be set for acceptance tests")
	}
	if testGetEnv("ALTR_API_KEY", "") == "" && testGetEnv("TF_ACC", "") == "1" {
		t.Fatal("ALTR_API_KEY must be set for acceptance tests")
	}
	if testGetEnv("ALTR_SECRET", "") == "" && testGetEnv("TF_ACC", "") == "1" {
		t.Fatal("ALTR_SECRET must be set for acceptance tests")
	}
}

// Helper function to get environment variables with defaults
func testGetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
