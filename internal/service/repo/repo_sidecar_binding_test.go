// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/google/uuid"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRepoSidecarBindingResource_basic(t *testing.T) {
	resourceName := "altr_repo_sidecar_binding.test"
	repoResourceName := "altr_repo.test"
	sidecarResourceName := "altr_sidecar.test"
	listenerResourceName := "altr_sidecar_listener.test"

	sidecarName := sdkacctest.RandomWithPrefix("tf-acc-test")
	sidecarHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoName := fmt.Sprintf("repo_%d", rand.Int())
	dbType := "Oracle"
	repoHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoPort := sdkacctest.RandIntRange(1, 65535)
	listenerPort := sdkacctest.RandIntRange(1, 65535)
	listenerAdvertisedVersion := "19.0.0.0"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoSidecarBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoSidecarBindingResourceConfig_basic(sidecarName, sidecarHostname, pubKeyExample1, repoName, dbType, repoHostname, dbType, listenerAdvertisedVersion, repoPort, listenerPort),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoSidecarBindingExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "port", listenerResourceName, "port"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(listenerPort)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(fmt.Sprintf(".*:%d:%s", listenerPort, repoName))),
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

func TestAccRepoSidecarBindingResource_multipleRepos(t *testing.T) {
	resourceName1 := "altr_repo_sidecar_binding.test1"
	resourceName2 := "altr_repo_sidecar_binding.test2"
	repoResourceName1 := "altr_repo.test1"
	repoResourceName2 := "altr_repo.test2"
	sidecarResourceName := "altr_sidecar.test"
	listenerResourceName := "altr_sidecar_listener.test"
	sidecarName := sdkacctest.RandomWithPrefix("tf-acc-test")
	sidecarHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repo1Name := fmt.Sprintf("repo_%d", rand.Int())
	repo2Name := fmt.Sprintf("repo_%d", rand.Int())
	dbType := "Oracle"
	repo1Hostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repo2Hostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repo1Port := sdkacctest.RandIntRange(1, 65535)
	repo2Port := sdkacctest.RandIntRange(1, 65535)
	listenerPort := sdkacctest.RandIntRange(1, 65535)
	listenerAdvertisedVersion := "19.0.0.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoSidecarBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoSidecarBindingResourceConfig_multipleRepos(sidecarName, sidecarHostname, pubKeyExample1, repo1Name, repo2Name, dbType, repo1Hostname, repo2Hostname, dbType, listenerAdvertisedVersion, repo1Port, repo2Port, listenerPort),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoSidecarBindingExists(resourceName1),
					testAccCheckRepoSidecarBindingExists(resourceName2),
					resource.TestCheckResourceAttrPair(resourceName1, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName2, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName1, "repo_name", repoResourceName1, "name"),
					resource.TestCheckResourceAttrPair(resourceName2, "repo_name", repoResourceName2, "name"),
					resource.TestCheckResourceAttrPair(resourceName1, "port", listenerResourceName, "port"),
					resource.TestCheckResourceAttrPair(resourceName2, "port", listenerResourceName, "port"),
				),
			},
		},
	})
}

func TestAccRepoSidecarBindingResource_multiplePorts(t *testing.T) {
	resourceName1 := "altr_repo_sidecar_binding.test1"
	resourceName2 := "altr_repo_sidecar_binding.test2"
	repoResourceName := "altr_repo.test"
	sidecarResourceName := "altr_sidecar.test"
	listenerResourceName1 := "altr_sidecar_listener.test1"
	listenerResourceName2 := "altr_sidecar_listener.test2"
	sidecarName := sdkacctest.RandomWithPrefix("tf-acc-test")
	sidecarHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoName := fmt.Sprintf("repo_%d", rand.Int())
	dbType := "Oracle"
	repoHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoPort := sdkacctest.RandIntRange(1, 65535)
	listener1Port := sdkacctest.RandIntRange(1, 65535)
	listener2Port := sdkacctest.RandIntRange(1, 65535)
	listenerAdvertisedVersion := "19.0.0.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoSidecarBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoSidecarBindingResourceConfig_multiplePorts(sidecarName, sidecarHostname, pubKeyExample1, repoName, dbType, repoHostname, dbType, listenerAdvertisedVersion, repoPort, listener1Port, listener2Port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoSidecarBindingExists(resourceName1),
					testAccCheckRepoSidecarBindingExists(resourceName2),
					resource.TestCheckResourceAttrPair(resourceName1, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName2, "sidecar_id", sidecarResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName1, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName2, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName1, "port", listenerResourceName1, "port"),
					resource.TestCheckResourceAttrPair(resourceName2, "port", listenerResourceName2, "port"),
					resource.TestCheckResourceAttr(resourceName1, "port", strconv.Itoa(listener1Port)),
					resource.TestCheckResourceAttr(resourceName2, "port", strconv.Itoa(listener2Port)),
				),
			},
		},
	})
}

func TestAccRepoSidecarBindingResource_requiredFieldsValidation(t *testing.T) {
	sidecarName := sdkacctest.RandomWithPrefix("tf-acc-test")
	sidecarHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoName := fmt.Sprintf("repo_%d", rand.Int())
	dbType := "Oracle"
	repoHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoPort := sdkacctest.RandIntRange(1, 65535)
	listenerPort := sdkacctest.RandIntRange(1, 65535)
	listenerAdvertisedVersion := "19.0.0.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRepoSidecarBindingResourceConfig_emptySidecarID(sidecarName, sidecarHostname, pubKeyExample1, repoName, dbType, repoHostname, dbType, listenerAdvertisedVersion, repoPort, listenerPort),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
			{
				Config:      testAccRepoSidecarBindingResourceConfig_emptyRepoName(sidecarName, sidecarHostname, pubKeyExample1, repoName, dbType, repoHostname, dbType, listenerAdvertisedVersion, repoPort, listenerPort),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value`),
			},
		},
	})
}

func TestAccRepoSidecarBindingResource_nonExistentListener(t *testing.T) {
	sidecarName := sdkacctest.RandomWithPrefix("tf-acc-test")
	sidecarHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoName := fmt.Sprintf("repo_%d", rand.Int())
	dbType := "Oracle"
	repoHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoPort := sdkacctest.RandIntRange(2, 65535)
	listenerPort := sdkacctest.RandIntRange(2, 65535)
	listenerAdvertisedVersion := "19.0.0.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRepoSidecarBindingResourceConfig_nonExistentListener(sidecarName, sidecarHostname, pubKeyExample1, repoName, dbType, repoHostname, dbType, listenerAdvertisedVersion, repoPort, listenerPort),
				ExpectError: regexp.MustCompile(`Error creating repo sidecar binding`),
			},
		},
	})
}

func TestAccRepoSidecarBindingResource_disappears(t *testing.T) {
	resourceName := "altr_repo_sidecar_binding.test"
	sidecarName := sdkacctest.RandomWithPrefix("tf-acc-test")
	sidecarHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoName := fmt.Sprintf("repo_%d", rand.Int())
	dbType := "Oracle"
	repoHostname := fmt.Sprintf("%s.example.altr.com", uuid.New().String())
	repoPort := sdkacctest.RandIntRange(1, 65535)
	listenerPort := sdkacctest.RandIntRange(1, 65535)
	listenerAdvertisedVersion := "19.0.0.0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoSidecarBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoSidecarBindingResourceConfig_basic(sidecarName, sidecarHostname, pubKeyExample1, repoName, dbType, repoHostname, dbType, listenerAdvertisedVersion, repoPort, listenerPort),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoSidecarBindingExists(resourceName),
					testAccCheckRepoSidecarBindingDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepoSidecarBindingExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Repo Sidecar Binding ID is set")
		}

		// Parse the ID to get sidecar_id, port, and repo_name
		sidecarID := rs.Primary.Attributes["sidecar_id"]
		portStr := rs.Primary.Attributes["port"]
		repoName := rs.Primary.Attributes["repo_name"]

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("Invalid port in state: %s", portStr)
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

		binding, err := conn.GetRepoSidecarBinding(sidecarID, repoName, port)
		if err != nil {
			return err
		}

		if binding == nil {
			return fmt.Errorf("Repo Sidecar Binding not found")
		}

		return nil
	}
}

func testAccCheckRepoSidecarBindingDestroy(s *terraform.State) error {
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
		if rs.Type != "altr_repo_sidecar_binding" {
			continue
		}

		// Parse the ID to get sidecar_id, port, and repo_name
		sidecarID := rs.Primary.Attributes["sidecar_id"]
		portStr := rs.Primary.Attributes["port"]
		repoName := rs.Primary.Attributes["repo_name"]

		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue // Skip if we can't parse the port
		}

		binding, err := conn.GetRepoSidecarBinding(sidecarID, repoName, port)
		if err != nil {
			return err
		}

		if binding != nil {
			return fmt.Errorf("Repo Sidecar Binding %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRepoSidecarBindingDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Parse the ID to get sidecar_id, port, and repo_name
		sidecarID := rs.Primary.Attributes["sidecar_id"]
		portStr := rs.Primary.Attributes["port"]
		repoName := rs.Primary.Attributes["repo_name"]

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("Invalid port in state: %s", portStr)
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

		return conn.DeleteRepoSidecarBinding(sidecarID, repoName, port)
	}
}

func testAccRepoSidecarBindingResourceConfig_basic(sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, listenerDatabaseType, listenerAdvertisedVersion string, repoPort, listenerPort int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_repo" "test" {
  name     = %[4]q
  type     = %[5]q
  hostname = %[6]q
  port     = %[7]d
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[8]d
  database_type      = %[9]q
  advertised_version = %[10]q
}

resource "altr_repo_sidecar_binding" "test" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = altr_repo.test.name
  port       = altr_sidecar_listener.test.port
}

`, sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, repoPort, listenerPort, listenerDatabaseType, listenerAdvertisedVersion)
}

func testAccRepoSidecarBindingResourceConfig_multipleRepos(sidecarName, sidecarHostname, sidecarPubKey1, repo1Name, repo2Name, repoType,
	repo1Hostname, repo2Hostname, listenerDatabaseType, listenerAdvertisedVersion string, repo1Port, repo2Port, listenerPort int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_repo" "test1" {
  name     = %[4]q
  type     = %[5]q
  hostname = %[6]q
  port     = %[7]d
}

resource "altr_repo" "test2" {
  name     = %[8]q
  type     = %[9]q
  hostname = %[10]q
  port     = %[11]d
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[12]d
  database_type      = %[13]q
  advertised_version = %[14]q
}

resource "altr_repo_sidecar_binding" "test1" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = altr_repo.test1.name
  port       = altr_sidecar_listener.test.port
}

resource "altr_repo_sidecar_binding" "test2" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = altr_repo.test2.name
  port       = altr_sidecar_listener.test.port
}

`, sidecarName, sidecarHostname, sidecarPubKey1, repo1Name, repoType, repo1Hostname, repo1Port, repo2Name, repoType, repo2Hostname, repo2Port, listenerPort, listenerDatabaseType, listenerAdvertisedVersion)
}

func testAccRepoSidecarBindingResourceConfig_multiplePorts(sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, listenerDatabaseType, listenerAdvertisedVersion string, repoPort, listener1Port, listener2Port int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_repo" "test" {
  name     = %[4]q
  type     = %[5]q
  hostname = %[6]q
  port     = %[7]d
}

resource "altr_sidecar_listener" "test1" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[8]d
  database_type      = %[10]q
  advertised_version = %[11]q
}

resource "altr_sidecar_listener" "test2" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[9]d
  database_type      = %[10]q
  advertised_version = %[11]q
}

resource "altr_repo_sidecar_binding" "test1" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = altr_repo.test.name
  port       = altr_sidecar_listener.test1.port
}

resource "altr_repo_sidecar_binding" "test2" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = altr_repo.test.name
  port       = altr_sidecar_listener.test2.port
}

`, sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, repoPort, listener1Port, listener2Port, listenerDatabaseType, listenerAdvertisedVersion)
}

func testAccRepoSidecarBindingResourceConfig_emptySidecarID(sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, listenerDatabaseType, listenerAdvertisedVersion string, repoPort, listenerPort int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_repo" "test" {
  name     = %[4]q
  type     = %[5]q
  hostname = %[6]q
  port     = %[7]d
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[8]d
  database_type      = %[9]q
  advertised_version = %[10]q
}

resource "altr_repo_sidecar_binding" "test" {
  sidecar_id = ""
  repo_name  = altr_repo.test.name
  port       = altr_sidecar_listener.test.port
}

`, sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, repoPort, listenerPort, listenerDatabaseType, listenerAdvertisedVersion)
}

func testAccRepoSidecarBindingResourceConfig_emptyRepoName(sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, listenerDatabaseType, listenerAdvertisedVersion string, repoPort, listenerPort int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_repo" "test" {
  name     = %[4]q
  type     = %[5]q
  hostname = %[6]q
  port     = %[7]d
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[8]d
  database_type      = %[9]q
  advertised_version = %[10]q
}

resource "altr_repo_sidecar_binding" "test" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = ""
  port       = altr_sidecar_listener.test.port
}

`, sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, repoPort, listenerPort, listenerDatabaseType, listenerAdvertisedVersion)
}

func testAccRepoSidecarBindingResourceConfig_nonExistentListener(sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, listenerDatabaseType, listenerAdvertisedVersion string, repoPort, listenerPort int) string {
	return fmt.Sprintf(`
resource "altr_sidecar" "test" {
  name         = %[1]q
  hostname     = %[2]q
  public_key_1 = %[3]q
}

resource "altr_repo" "test" {
  name     = %[4]q
  type     = %[5]q
  hostname = %[6]q
  port     = %[7]d
}

resource "altr_sidecar_listener" "test" {
  sidecar_id         = altr_sidecar.test.id
  port               = %[8]d
  database_type      = %[9]q
  advertised_version = %[10]q
}

resource "altr_repo_sidecar_binding" "test" {
  sidecar_id = altr_sidecar.test.id
  repo_name  = altr_repo.test.name
  port       = 1
}

`, sidecarName, sidecarHostname, sidecarPubKey1, repoName, repoType, repoHostname, repoPort, listenerPort, listenerDatabaseType, listenerAdvertisedVersion)
}

func testAccCheckSidecarDisappears(resourceName string) resource.TestCheckFunc {
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

		return conn.DeleteSidecar(rs.Primary.ID)
	}
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
			acctest.TestGetEnv("ALTR_ORG_ID", "test-org"),
			acctest.TestGetEnv("ALTR_API_KEY", "test-key"),
			acctest.TestGetEnv("ALTR_SECRET", "test-secret"),
			acctest.TestGetEnv("ALTR_BASE_URL", ""),
		)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		return conn.DeregisterSidecarListener(sidecarID, port)
	}
}

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
