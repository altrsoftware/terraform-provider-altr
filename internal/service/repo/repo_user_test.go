// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/altrsoftware/terraform-provider-altr/internal/acctest"
	"github.com/altrsoftware/terraform-provider-altr/internal/client"
)

func TestAccRepoUserResource_basicAWSSecretsManager(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	secretsPath := "/test/secrets/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_basicAWSSecretsManager(repoName, username, secretsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoExists(repoResourceName),
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(fmt.Sprintf("%s:%s", repoName, username))),
				),
			},
		},
	})
}

func TestAccRepoUserResource_AWSSecretsManagerWithIAMRole(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	secretsPath := "/test/secrets/path"
	iamRole := "arn:aws:iam::123456789012:role/test-role"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_AWSSecretsManagerWithIAMRole(repoName, username, secretsPath, iamRole),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.iam_role", iamRole),
				),
			},
		},
	})
}

func TestAccRepoUserResource_basicAzureKeyVault(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	keyVaultURI := "https://test-vault.vault.azure.net/"
	secretName := "test-secret"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_basicAzureKeyVault(repoName, username, keyVaultURI, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.key_vault_uri", keyVaultURI),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.secret_name", secretName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccRepoUserResource_multipleUsers(t *testing.T) {
	resourceName1 := "altr_repo_user.test1"
	resourceName2 := "altr_repo_user.test2"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username1 := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	username2 := acctest.RandomWithPrefixUnderscoreMaxLength("repouser2", 32)
	secretsPath1 := "/test/secrets/path1"
	keyVaultURI := "https://test-vault.vault.azure.net/"
	secretName := "test-secret"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_multipleUsers(repoName, username1, username2, secretsPath1, keyVaultURI, secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName1),
					testAccCheckRepoUserExists(resourceName2),
					resource.TestCheckResourceAttrPair(resourceName1, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName2, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName1, "username", username1),
					resource.TestCheckResourceAttr(resourceName2, "username", username2),
					resource.TestCheckResourceAttr(resourceName1, "aws_secrets_manager.secrets_path", secretsPath1),
					resource.TestCheckResourceAttr(resourceName2, "azure_key_vault.key_vault_uri", keyVaultURI),
					resource.TestCheckResourceAttr(resourceName2, "azure_key_vault.secret_name", secretName),
				),
			},
		},
	})
}

func TestAccRepoUserResource_credentialStoreValidation(t *testing.T) {
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRepoUserResourceConfig_noCredentialStore(repoName, username),
				ExpectError: regexp.MustCompile(`exactly one credential store must be specified`),
			},
			{
				Config:      testAccRepoUserResourceConfig_bothCredentialStores(repoName, username),
				ExpectError: regexp.MustCompile(`only one credential store can be specified at a time`),
			},
		},
	})
}

func TestAccRepoUserResource_requiredFieldsValidation(t *testing.T) {
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	secretsPath := "/test/secrets/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRepoUserResourceConfig_emptyUsername(repoName, secretsPath),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Length`), // Caught by the framework
			},
			{
				Config:      testAccRepoUserResourceConfig_emptySecretsPath(repoName, "testuser"),
				ExpectError: regexp.MustCompile(`Error: Invalid Configuration`), // Caught by the framework
			},
		},
	})
}

func TestAccRepoUserResource_disappears(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	secretsPath := "/test/secrets/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_basicAWSSecretsManager(repoName, username, secretsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					testAccCheckRepoUserDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRepoUserResource_repoDisappears(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	secretsPath := "/test/secrets/path"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_basicAWSSecretsManager(repoName, username, secretsPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					testAccCheckRepoUserDisappears(resourceName),
					testAccCheckRepoDisappears(repoResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRepoUserResource_updateRepoUserAWSSecretsManager(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	secretsPath1 := "/test/secrets/path1"
	secretsPath2 := "/test/secrets/path2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_basicAWSSecretsManager(repoName, username, secretsPath1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath1),
				),
			},
			{
				Config: testAccRepoUserResourceConfig_basicAWSSecretsManager(repoName, username, secretsPath2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "aws_secrets_manager.secrets_path", secretsPath2),
				),
			},
		},
	})
}

func TestAccRepoUserResource_updateRepoUserAzureKeyVault(t *testing.T) {
	resourceName := "altr_repo_user.test"
	repoResourceName := "altr_repo.test"
	repoName := acctest.RandomWithPrefixUnderscoreMaxLength("repo_user_test", 32)
	username := acctest.RandomWithPrefixUnderscoreMaxLength("repouser", 32)
	keyVaultURI1 := "https://test-vault.vault.azure.net/"
	secretName1 := "test-secret"
	keyVaultURI2 := "https://test-vault2.vault.azure.net/"
	secretName2 := "test-secret-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRepoUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoUserResourceConfig_basicAzureKeyVault(repoName, username, keyVaultURI1, secretName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.key_vault_uri", keyVaultURI1),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.secret_name", secretName1),
				),
			},
			{
				Config: testAccRepoUserResourceConfig_basicAzureKeyVault(repoName, username, keyVaultURI2, secretName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepoUserExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.key_vault_uri", keyVaultURI2),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault.secret_name", secretName2)),
			},
		},
	})
}

func testAccCheckRepoUserExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("remote config %s", s.RootModule().String())
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {

			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Repo User ID is set")
		}

		// Parse the ID to get repo_name and username
		repoName := rs.Primary.Attributes["repo_name"]
		username := rs.Primary.Attributes["username"]

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

		repoUser, err := conn.GetRepoUser(repoName, username)
		if err != nil {
			return err
		}

		if repoUser == nil {
			return fmt.Errorf("Repo User not found")
		}

		return nil
	}
}

func testAccCheckRepoUserDestroy(s *terraform.State) error {
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
		if rs.Type != "altr_repo_user" {
			continue
		}

		// Parse the ID to get repo_name and username
		repoName := rs.Primary.Attributes["repo_name"]
		username := rs.Primary.Attributes["username"]

		repoUser, err := conn.GetRepoUser(repoName, username)
		if err != nil {
			return err
		}

		if repoUser != nil {
			return fmt.Errorf("Repo User %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckRepoUserDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		// Parse the ID to get repo_name and username
		repoName := rs.Primary.Attributes["repo_name"]
		username := rs.Primary.Attributes["username"]

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

		return conn.DeleteRepoUser(repoName, username)
	}
}

func testAccRepoUserResourceConfig_basicAWSSecretsManager(repoName, username, secretsPath string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name      = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  
  aws_secrets_manager = {
    secrets_path = %[3]q
  }
}
`, repoName, username, secretsPath)
}

func testAccRepoUserResourceConfig_AWSSecretsManagerWithIAMRole(repoName, username, secretsPath, iamRole string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name             = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  
  aws_secrets_manager = {
    secrets_path = %[3]q
    iam_role     = %[4]q
  }
}
`, repoName, username, secretsPath, iamRole)
}

func testAccRepoUserResourceConfig_basicAzureKeyVault(repoName, username, keyVaultURI, secretName string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name      = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  
  azure_key_vault  = {
    key_vault_uri = %[3]q
    secret_name   = %[4]q
  }
}
`, repoName, username, keyVaultURI, secretName)
}

func testAccRepoUserResourceConfig_multipleUsers(repoName, username1, username2, secretsPath, keyVaultURI, secretName string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name      = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test1" {
  repo_name = altr_repo.test.name
  username  = %[2]q
  
  aws_secrets_manager = {
    secrets_path = %[4]q
  }
}

resource "altr_repo_user" "test2" {
  repo_name = altr_repo.test.name
  username  = %[3]q
  
  azure_key_vault = {
    key_vault_uri = %[5]q
    secret_name   = %[6]q
  }
}
`, repoName, username1, username2, secretsPath, keyVaultURI, secretName)
}

func testAccRepoUserResourceConfig_noCredentialStore(repoName, username string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test" {
  repo_name = %[1]q
  username  = %[2]q
}
`, repoName, username)
}

func testAccRepoUserResourceConfig_bothCredentialStores(repoName, username string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test" {
  repo_name = %[1]q
  username  = %[2]q
  
  aws_secrets_manager = {
    secrets_path = "/test/secrets/path"
  }
  
  azure_key_vault  ={
    key_vault_uri = "https://test-vault.vault.azure.net/"
    secret_name   = "test-secret"
  }
}
`, repoName, username)
}

func testAccRepoUserResourceConfig_emptyUsername(repoName, secretsPath string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name = %[1]q

  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}
resource "altr_repo_user" "test" {
  repo_name = %[1]q
  username  = ""
  
  aws_secrets_manager = {
    secrets_path = %[2]q
  }
}
`, repoName, secretsPath)
}

func testAccRepoUserResourceConfig_emptySecretsPath(repoName, username string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name = %[1]q
  hostname  = "test-host"
  port      = 5432
  type      = "Oracle"
}

resource "altr_repo_user" "test" {
  repo_name = %[1]q
  username  = %[2]q
  
  aws_secrets_manager = {
    secrets_path = ""
  }
}
`, repoName, username)
}
