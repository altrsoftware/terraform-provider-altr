// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/altrsoftware/terraform-provider-altr/internal/client"
)

func TestAccServiceUserResource_basicAWSSecretsManager(t *testing.T) {
	resourceName := "altr_service_user.test"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)
	secretsPath := "/test/secrets/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_awsSecretsManager(repoName, username, secretsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "resource", "ORCL"),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(fmt.Sprintf("%s:%s", repoName, username))),
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

func TestAccServiceUserResource_azureKeyVault(t *testing.T) {
	resourceName := "altr_service_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)
	keyVaultURI := "https://test-vault.vault.azure.net/"
	secretName := "test-secret"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_azureKeyVault(repoName, username, keyVaultURI, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.key_vault_uri", keyVaultURI),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.secret_name", secretName),
				),
			},
		},
	})
}

func TestAccServiceUserResource_environmentVariable(t *testing.T) {
	resourceName := "altr_service_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)
	variableName := "ALTR_DB_SECRET"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_environmentVariable(repoName, username, variableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "environment_variable.variable_name", variableName),
				),
			},
		},
	})
}

func TestAccServiceUserResource_secretFile(t *testing.T) {
	resourceName := "altr_service_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)
	path := "db-secret.json"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_secretFile(repoName, username, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "secret_file.path", path),
				),
			},
		},
	})
}

func TestAccServiceUserResource_credentialProviderValidation(t *testing.T) {
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceUserResourceConfig_noCredentialProvider(repoName, username),
				ExpectError: regexp.MustCompile(`exactly one credential provider must be specified`),
			},
			{
				Config:      testAccServiceUserResourceConfig_twoCredentialProviders(repoName, username),
				ExpectError: regexp.MustCompile(`only one credential provider can be specified at a time`),
			},
		},
	})
}

func TestAccServiceUserResource_update(t *testing.T) {
	resourceName := "altr_service_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)
	secretsPath1 := "/test/secrets/path1"
	secretsPath2 := "/test/secrets/path2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_awsSecretsManager(repoName, username, secretsPath1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath1),
				),
			},
			{
				Config: testAccServiceUserResourceConfig_awsSecretsManager(repoName, username, secretsPath2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath2),
				),
			},
		},
	})
}

func TestAccServiceUserResource_disappears(t *testing.T) {
	resourceName := "altr_service_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("su_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("svcuser", 32)
	secretsPath := "/test/secrets/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserResourceConfig_awsSecretsManager(repoName, username, secretsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceUserExists(resourceName),
					testAccCheckServiceUserDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccServiceUserClient builds an API client from the standard acceptance
// test environment variables.
func testAccServiceUserClient() (*client.Client, error) {
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

func testAccCheckServiceUserExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service User ID is set")
		}

		conn, err := testAccServiceUserClient()
		if err != nil {
			return err
		}

		repoName := rs.Primary.Attributes["repo_name"]
		username := rs.Primary.Attributes["username"]

		serviceUser, err := conn.GetServiceUser(repoName, username)
		if err != nil {
			return err
		}

		if serviceUser == nil {
			return fmt.Errorf("Service User not found")
		}

		return nil
	}
}

func testAccCheckServiceUserDestroy(s *terraform.State) error {
	conn, err := testAccServiceUserClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_service_user" {
			continue
		}

		repoName := rs.Primary.Attributes["repo_name"]
		username := rs.Primary.Attributes["username"]

		serviceUser, err := conn.GetServiceUser(repoName, username)
		if err != nil {
			return err
		}

		if serviceUser != nil {
			return fmt.Errorf("Service User %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckServiceUserDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn, err := testAccServiceUserClient()
		if err != nil {
			return err
		}

		repoName := rs.Primary.Attributes["repo_name"]
		username := rs.Primary.Attributes["username"]

		return conn.DeleteServiceUser(repoName, username)
	}
}

func testAccServiceUserResourceConfig_awsSecretsManager(repoName, username, secretsPath string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  resource  = "ORCL"

  aws_secrets_manager = {
    secrets_path = %[3]q
  }
}
`, repoName, username, secretsPath)
}

func testAccServiceUserResourceConfig_azureKeyVault(repoName, username, keyVaultURI, secretName string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  resource  = "ORCL"

  azure_key_vault = {
    key_vault_uri = %[3]q
    secret_name   = %[4]q
  }
}
`, repoName, username, keyVaultURI, secretName)
}

func testAccServiceUserResourceConfig_environmentVariable(repoName, username, variableName string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  resource  = "ORCL"

  environment_variable = {
    variable_name = %[3]q
  }
}
`, repoName, username, variableName)
}

func testAccServiceUserResourceConfig_secretFile(repoName, username, path string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  resource  = "ORCL"

  secret_file = {
    path = %[3]q
  }
}
`, repoName, username, path)
}

func testAccServiceUserResourceConfig_noCredentialProvider(repoName, username string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  resource  = "ORCL"
}
`, repoName, username)
}

func testAccServiceUserResourceConfig_twoCredentialProviders(repoName, username string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = %[1]q
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  resource  = "ORCL"

  aws_secrets_manager = {
    secrets_path = "/test/secrets/path"
  }

  environment_variable = {
    variable_name = "ALTR_DB_SECRET"
  }
}
`, repoName, username)
}
