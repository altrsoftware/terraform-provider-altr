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
)

func TestAccAgentTaskResource_basic(t *testing.T) {
	resourceName := "altr_agent_task.test"
	agentResourceName := "altr_agent.test"
	repoResourceName := "altr_repo.test"
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("task_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentTaskResourceConfig_basic(prefix, "0 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentTaskExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "agent_id", agentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name", prefix+"_task"),
					resource.TestCheckResourceAttr(resourceName, "configuration.classification_type", "5"),
					resource.TestCheckResourceAttr(resourceName, "configuration.sample_strategy", "ROWS"),
					resource.TestCheckResourceAttr(resourceName, "configuration.collection_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "schedule.type", "CRON"),
					resource.TestCheckResourceAttr(resourceName, "schedule.value", "0 0 * * *"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAgentTaskImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccAgentTaskResource_sis(t *testing.T) {
	resourceName := "altr_agent_task.test"
	agentResourceName := "altr_agent.test"
	repoResourceName := "altr_repo.test"
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("task_sis", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentTaskResourceConfig_sis(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentTaskExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "agent_id", agentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "repo_name", repoResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "configuration.audit_file_path", "/var/lib/postgresql/audit/*.json"),
					resource.TestCheckResourceAttr(resourceName, "configuration.audit_file_type", "json"),
					// SIS tasks don't use the classifier fields; the empty->null mapping
					// must leave them unset (guards the inconsistent-result risk).
					resource.TestCheckNoResourceAttr(resourceName, "configuration.classification_type"),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.sample_strategy"),
					resource.TestCheckResourceAttr(resourceName, "schedule.type", "CRON"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				// Re-applying the same config must produce an empty plan (no
				// perpetual diff / inconsistent-result from the empty->null mapping).
				Config:             testAccAgentTaskResourceConfig_sis(prefix),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAgentTaskImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccAgentTaskResource_update(t *testing.T) {
	resourceName := "altr_agent_task.test"
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("task_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentTaskResourceConfig_basic(prefix, "0 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentTaskExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.value", "0 0 * * *"),
				),
			},
			{
				Config: testAccAgentTaskResourceConfig_basic(prefix, "30 2 * * *"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentTaskExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.value", "30 2 * * *"),
				),
			},
		},
	})
}

func TestAccAgentTaskResource_scheduleTypeValidation(t *testing.T) {
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("task_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAgentTaskResourceConfig_invalidScheduleType(prefix),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

func TestAccAgentTaskResource_disappears(t *testing.T) {
	resourceName := "altr_agent_task.test"
	prefix := acctest.RandomWithPrefixUnderscoreMaxLength("task_test", 24)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAgentTaskResourceConfig_basic(prefix, "0 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgentTaskExists(resourceName),
					testAccCheckAgentTaskDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAgentTaskExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Agent Task ID is set")
		}

		conn, err := testAccAgentClient()
		if err != nil {
			return err
		}

		agentID := rs.Primary.Attributes["agent_id"]

		task, err := conn.GetAgentTask(agentID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if task == nil {
			return fmt.Errorf("Agent Task not found")
		}

		return nil
	}
}

func testAccCheckAgentTaskDestroy(s *terraform.State) error {
	conn, err := testAccAgentClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "altr_agent_task" {
			continue
		}

		agentID := rs.Primary.Attributes["agent_id"]

		task, err := conn.GetAgentTask(agentID, rs.Primary.ID)
		if err != nil {
			return err
		}

		if task != nil {
			return fmt.Errorf("Agent Task %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAgentTaskDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn, err := testAccAgentClient()
		if err != nil {
			return err
		}

		agentID := rs.Primary.Attributes["agent_id"]

		return conn.DeleteAgentTask(agentID, rs.Primary.ID)
	}
}

// testAccAgentTaskImportStateIDFunc builds the "agent_id:task_id" import ID
// expected by the resource's ImportState implementation.
func testAccAgentTaskImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["agent_id"], rs.Primary.ID), nil
	}
}

// testAccAgentTaskResourceConfig_base provisions the repo, agent, and service
// user that a task depends on.
func testAccAgentTaskResourceConfig_base(prefix string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = "%[1]s_repo"
  hostname = "test-host"
  port     = 5432
  type     = "Oracle"
}

resource "altr_agent" "test" {
  type         = "CLASSIFIER"
  name         = "%[1]s_agent"
  public_key_1 = <<-EOT
%[2]s
EOT
}

resource "altr_service_user" "test" {
  repo_name = altr_repo.test.name
  username  = "%[1]s_user"
  resource  = "ORCL"

  aws_secrets_manager = {
    secrets_path = "/test/secrets/path"
  }
}
`, prefix, testAgentPublicKey1)
}

func testAccAgentTaskResourceConfig_basic(prefix, cron string) string {
	return testAccAgentTaskResourceConfig_base(prefix) + fmt.Sprintf(`
resource "altr_agent_task" "test" {
  agent_id     = altr_agent.test.id
  name         = "%[1]s_task"
  repo_name    = altr_repo.test.name
  service_user = altr_service_user.test.username

  configuration = {
    classification_type = 5
    sample_strategy     = "ROWS"
    collection_name     = "default"
  }

  schedule = {
    type         = "CRON"
    value        = %[2]q
    max_duration = "PT30M"
  }
}
`, prefix, cron)
}

func testAccAgentTaskResourceConfig_sis(prefix string) string {
	return fmt.Sprintf(`
resource "altr_repo" "test" {
  name     = "%[1]s_repo"
  hostname = "test-host"
  port     = 5432
  type     = "Postgres"
}

resource "altr_agent" "test" {
  type         = "SIS"
  name         = "%[1]s_agent"
  public_key_1 = <<-EOT
%[2]s
EOT
}

resource "altr_agent_task" "test" {
  agent_id  = altr_agent.test.id
  name      = "%[1]s_task"
  repo_name = altr_repo.test.name

  configuration = {
    audit_file_path = "/var/lib/postgresql/audit/*.json"
    audit_file_type = "json"
  }

  schedule = {
    type  = "CRON"
    value = "*/5 * * * *"
  }
}
`, prefix, testAgentPublicKey1)
}

func testAccAgentTaskResourceConfig_invalidScheduleType(prefix string) string {
	return testAccAgentTaskResourceConfig_base(prefix) + fmt.Sprintf(`
resource "altr_agent_task" "test" {
  agent_id     = altr_agent.test.id
  name         = "%[1]s_task"
  repo_name    = altr_repo.test.name
  service_user = altr_service_user.test.username

  configuration = {
    classification_type = 5
    sample_strategy     = "ROWS"
    collection_name     = "default"
  }

  schedule = {
    type  = "INTERVAL"
    value = "0 0 * * *"
  }
}
`, prefix)
}
