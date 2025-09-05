package sidecar_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"terraform-provider-altr/internal/acctest"
	"terraform-provider-altr/internal/client"
)

func TestAccSidecarListenerResource_basic(t *testing.T) {
	resourceName := "altr_sidecar_listener.test"
	sidecarResourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	port := sdkacctest.RandIntRange(3000, 9000)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarListenerResourceConfig_basic(rName, rHostname, pubKeyExample1, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarListenerExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(port)),
					resource.TestCheckResourceAttr(resourceName, "database_type", "Oracle"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(fmt.Sprintf(".*:%d", port))),
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

func TestAccSidecarListenerResource_withAdvertisedVersion(t *testing.T) {
	resourceName := "altr_sidecar_listener.test"
	sidecarResourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	port := sdkacctest.RandIntRange(3000, 9000)
	advertisedVersion := "14.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarListenerResourceConfig_withAdvertisedVersion(rName, rHostname, pubKeyExample1, port, advertisedVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarListenerExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(port)),
					resource.TestCheckResourceAttr(resourceName, "database_type", "Oracle"),
					resource.TestCheckResourceAttr(resourceName, "advertised_version", advertisedVersion),
				),
			},
		},
	})
}

func TestAccSidecarListenerResource_oracle(t *testing.T) {
	resourceName := "altr_sidecar_listener.test"
	sidecarResourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	port := sdkacctest.RandIntRange(3000, 9000)
	advertisedVersion := "8.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarListenerResourceConfig_oracle(rName, rHostname, pubKeyExample1, port, advertisedVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarListenerExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(port)),
					resource.TestCheckResourceAttr(resourceName, "database_type", "Oracle"),
					resource.TestCheckResourceAttr(resourceName, "advertised_version", advertisedVersion),
				),
			},
		},
	})
}

func TestAccSidecarListenerResource_multiplePorts(t *testing.T) {
	resourceName1 := "altr_sidecar_listener.test1"
	resourceName2 := "altr_sidecar_listener.test2"
	sidecarResourceName := "altr_sidecar.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	port1 := sdkacctest.RandIntRange(3000, 6000)
	port2 := sdkacctest.RandIntRange(6001, 9000)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarListenerResourceConfig_multiplePorts(rName, rHostname, pubKeyExample1, port1, port2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarListenerExists(resourceName1),
					testAccCheckSidecarListenerExists(resourceName2),
					resource.TestCheckResourceAttrPair(resourceName1, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName2, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName1, "port", strconv.Itoa(port1)),
					resource.TestCheckResourceAttr(resourceName2, "port", strconv.Itoa(port2)),
					resource.TestCheckResourceAttr(resourceName1, "database_type", "Oracle"),
					resource.TestCheckResourceAttr(resourceName2, "database_type", "Oracle"),
				),
			},
		},
	})
}

func TestAccSidecarListenerResource_portValidation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSidecarListenerResourceConfig_invalidPort(rName, rHostname, pubKeyExample1, 0),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
			{
				Config:      testAccSidecarListenerResourceConfig_invalidPort(rName, rHostname, pubKeyExample1, 70000),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccSidecarListenerResource_databaseTypeValidation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	port := sdkacctest.RandIntRange(3000, 9000)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSidecarListenerResourceConfig_invalidDatabaseType(rName, rHostname, pubKeyExample1, port, ""),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccSidecarListenerResource_disappears(t *testing.T) {
	resourceName := "altr_sidecar_listener.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rHostname := fmt.Sprintf("%s.example.altr.com", rName)
	port := sdkacctest.RandIntRange(3000, 9000)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSidecarListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSidecarListenerResourceConfig_basic(rName, rHostname, pubKeyExample1, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSidecarListenerExists(resourceName),
					testAccCheckSidecarListenerDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSidecarListenerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Sidecar Listener ID is set")
		}

		// Parse the ID to get sidecar_id and port
		sidecarID := rs.Primary.Attributes["sidecar_id"]
		portStr := rs.Primary.Attributes["port"]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("Invalid port in state: %s", portStr)
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

		listener, err := conn.GetSidecarListener(sidecarID, port)
		if err != nil {
			return err
		}

		if listener == nil {
			return fmt.Errorf("Sidecar Listener not found")
		}

		return nil
	}
}

func testAccCheckSidecarListenerDestroy(s *terraform.State) error {
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
		if rs.Type != "altr_sidecar_listener" {
			continue
		}

		// Parse the ID to get sidecar_id and port
		sidecarID := rs.Primary.Attributes["sidecar_id"]
		portStr := rs.Primary.Attributes["port"]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue // Skip if we can't parse the port
		}

		listener, err := conn.GetSidecarListener(sidecarID, port)
		if err != nil {
			return err
		}

		if listener != nil {
			return fmt.Errorf("Sidecar Listener %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckSidecarListenerDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Parse the ID to get sidecar_id and port
		sidecarID := rs.Primary.Attributes["sidecar_id"]
		portStr := rs.Primary.Attributes["port"]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("Invalid port in state: %s", portStr)
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

		return conn.DeregisterSidecarListener(sidecarID, port)
	}
}

func testAccSidecarListenerResourceConfig_basic(name, hostname, publicKey1 string, port int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_sidecar_listener" "test" {
  sidecar_id    = altr_sidecar.test.id
  port          = %[4]d
  database_type = "Oracle"
  advertised_version = "19.0.0.0"
}
`, name, hostname, publicKey1, port)
}

func testAccSidecarListenerResourceConfig_withAdvertisedVersion(name, hostname, publicKey1 string, port int, advertisedVersion string) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[4]d
  database_type      = "Oracle"
  advertised_version = %[5]q
}
`, name, hostname, publicKey1, port, advertisedVersion)
}

func testAccSidecarListenerResourceConfig_oracle(name, hostname, publicKey1 string, port int, advertisedVersion string) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[4]d
  database_type      = "Oracle"
  advertised_version = %[5]q
}
`, name, hostname, publicKey1, port, advertisedVersion)
}

func testAccSidecarListenerResourceConfig_multiplePorts(name, hostname, publicKey1 string, port1, port2 int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_sidecar_listener" "test1" {
  sidecar_id    = altr_sidecar.test.id
  port          = %[4]d
  database_type = "Oracle"
  advertised_version = "19.0.0.0"
}

resource "altr_sidecar_listener" "test2" {
  sidecar_id    = altr_sidecar.test.id
  port          = %[5]d
  database_type = "Oracle"
  advertised_version = "19.0.0.0"
}
`, name, hostname, publicKey1, port1, port2)
}

func testAccSidecarListenerResourceConfig_invalidPort(name, hostname, publicKey1 string, port int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_sidecar_listener" "test" {
  sidecar_id    = altr_sidecar.test.id
  port          = %[4]d
  database_type = "Oracle"
  advertised_version = "19.0.0.0"
}
`, name, hostname, publicKey1, port)
}

func testAccSidecarListenerResourceConfig_invalidDatabaseType(name, hostname, publicKey1 string, port int, databaseType string) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_sidecar_listener" "test" {
  sidecar_id    = altr_sidecar.test.id
  port          = %[4]d
  database_type = %[5]q
  advertised_version = "19.0.0.0"
}
`, name, hostname, publicKey1, port, databaseType)
}
