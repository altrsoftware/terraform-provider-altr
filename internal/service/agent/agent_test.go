// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/altrsoftware/terraform-provider-altr/internal/client"
)

// testAgentPublicKey1 and testAgentPublicKey2 are well-formed RSA public keys
// used across the agent, agent task, and service user acceptance tests.
const (
	testAgentPublicKey1 = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAy6R/os6auy9mvWYp2Bib
s66dnmwoNKrEa40V5bh1ET+T7XPssrl5n2tUDYbjYvQlAmDoh53+vmqt+hkS2AdB
JSAwz9xOjs0S9VeEsEMbhzorfZ8UAOQiZeRoKJDG95dy45SXivcsiGjGWVakgCSt
QfceFIAoxmFHJ3kFQGPe5kXvolOnfNsWblg3Y8j2uWGrowD2kIIbnsApdZE/cbKQ
nKm976Jt5GG6GggLvQ+zr/ix1omSzsOedc5kRYK+XbHq0YKJivDCyumz4jQEczx8
8x8aWCTDFFnr6GEb2b7t6+siRtAl0jZgf0/tCvlu0YqZgDlfZFBPGljY7J7rDH/u
kwIDAQAB
-----END PUBLIC KEY-----`

	testAgentPublicKey2 = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3kPCXOQKcBl3dQsWHfq4
F7oZ6+6qWmONextCAjADMs5JC3iS9M6lIojTn2Mpqo6FvbCUQmIY2XvRYN/etuvN
M8/oX7N/xcNmDLBBvqC0KlZ4p1KksAPX/GkykjiDcLDt8PjAcoXOXYZPFHf07Yzo
ec6JEdRkgAlw3CoEgNy6CTUWxlCF93FvXqTWqkdoPIxPKZN7/j84pzZY6sKlvkA+
LccI33pXo1Um8ZMaHSYJnSk6PZacVTqrZOdpXkSjcKF+27OWSuPV3NpG+TYICSG+
Rlnb+slUD0XiZA4JCpZdpfx47/1EIhZhIXaDlu0AVIMntbFzG9apOKlCEsX19kW1
pwIDAQAB
-----END PUBLIC KEY-----`
)

func TestAccAgentResource_basic(t *testing.T) {
	resourceName := "altr_agent.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("agent_test", 64)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "CLASSIFIER"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key_1"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`^[0-9a-fA-F-]+$`)),
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

func TestAccAgentResource_twoPublicKeys(t *testing.T) {
	resourceName := "altr_agent.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("agent_test", 64)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_twoPublicKeys(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "public_key_1"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key_2"),
				),
			},
		},
	})
}

func TestAccAgentResource_update(t *testing.T) {
	resourceName := "altr_agent.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("agent_test", 64)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_description(name, "initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
				),
			},
			{
				Config: testAccAgentResourceConfig_description(name, "updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
				),
			},
		},
	})
}

func TestAccAgentResource_publicKeyValidation(t *testing.T) {
	name := acctest.RandomWithPrefixUnderscoreMaxLength("agent_test", 64)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgentResourceConfig_noPublicKey(name),
				ExpectError: regexp.MustCompile(`at least one of 'public_key_1' or 'public_key_2' must be specified`),
			},
		},
	})
}

func TestAccAgentResource_typeValidation(t *testing.T) {
	name := acctest.RandomWithPrefixUnderscoreMaxLength("agent_test", 64)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgentResourceConfig_invalidType(name),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

func TestAccAgentResource_disappears(t *testing.T) {
	resourceName := "altr_agent.test"
	name := acctest.RandomWithPrefixUnderscoreMaxLength("agent_test", 64)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentResourceConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentExists(resourceName),
					testAccCheckAgentDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAgentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Agent ID is set")
		}

		conn, err := testAccAgentClient()
		if err != nil {
			return err
		}

		agent, err := conn.GetAgent(rs.Primary.ID)
		if err != nil {
			return err
		}

		if agent == nil {
			return fmt.Errorf("Agent not found")
		}

		return nil
	}
}

func testAccCheckAgentDestroy(s *terraform.State) error {
	conn, err := testAccAgentClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_agent" {
			continue
		}

		agent, err := conn.GetAgent(rs.Primary.ID)
		if err != nil {
			return err
		}

		if agent != nil {
			return fmt.Errorf("Agent %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAgentDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn, err := testAccAgentClient()
		if err != nil {
			return err
		}

		return conn.DeleteAgent(rs.Primary.ID)
	}
}

// testAccAgentClient builds an API client from the standard acceptance test
// environment variables.
func testAccAgentClient() (*client.Client, error) {
	conn, err := client.NewClient(
		acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
		acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
		acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
		acctest.TestGetEnv("ALTR_BASE_URL", ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create test client: %w", err)
	}

	return conn, nil
}

func testAccAgentResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "altr_agent" "test" {
  type         = "CLASSIFIER"
  name         = %[1]q
  public_key_1 = <<-EOT
%[2]s
EOT
}
`, name, testAgentPublicKey1)
}

func testAccAgentResourceConfig_twoPublicKeys(name string) string {
	return fmt.Sprintf(`
resource "altr_agent" "test" {
  type         = "CLASSIFIER"
  name         = %[1]q
  public_key_1 = <<-EOT
%[2]s
EOT
  public_key_2 = <<-EOT
%[3]s
EOT
}
`, name, testAgentPublicKey1, testAgentPublicKey2)
}

func testAccAgentResourceConfig_description(name, description string) string {
	return fmt.Sprintf(`
resource "altr_agent" "test" {
  type         = "CLASSIFIER"
  name         = %[1]q
  description  = %[2]q
  public_key_1 = <<-EOT
%[3]s
EOT
}
`, name, description, testAgentPublicKey1)
}

func testAccAgentResourceConfig_noPublicKey(name string) string {
	return fmt.Sprintf(`
resource "altr_agent" "test" {
  type = "CLASSIFIER"
  name = %[1]q
}
`, name)
}

func testAccAgentResourceConfig_invalidType(name string) string {
	return fmt.Sprintf(`
resource "altr_agent" "test" {
  type         = "INVALID"
  name         = %[1]q
  public_key_1 = <<-EOT
%[2]s
EOT
}
`, name, testAgentPublicKey1)
}
