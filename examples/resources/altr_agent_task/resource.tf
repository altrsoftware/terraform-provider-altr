# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_repo" "example" {
  name     = "example"
  type     = "Oracle"
  hostname = "example.com"
  port     = 1521
}

resource "altr_agent" "example" {
  type = "CLASSIFIER"
  name = "example"

  public_key_1 = file("${path.module}/agent_public_key.pem")
}

resource "altr_service_user" "example" {
  repo_name = altr_repo.example.name
  username  = "example"
  resource  = "ORCL" # actual Oracle service name the agent connects to

  aws_secrets_manager = {
    secrets_path = "arn:aws:secretsmanager:us-east-1:000000000000:secret:example-O3d19H"
  }
}

resource "altr_agent_task" "example" {
  agent_id     = altr_agent.example.id
  name         = "example"
  repo_name    = altr_repo.example.name
  service_user = altr_service_user.example.username

  configuration = {
    classification_type = 5
    sample_strategy     = "ROWS"
    collection_name     = "default"
  }

  schedule = {
    type         = "CRON"
    value        = "0 0 * * *"
    max_duration = "PT30M"
  }
}

# SIS (Security Intelligence Scout) audit-ingestion task. SIS tasks use the
# audit_* configuration fields instead of classification_type/sample_strategy.
resource "altr_repo" "postgres" {
  name     = "example-postgres"
  type     = "Postgres"
  hostname = "postgres.example.com"
  port     = 5432
}

resource "altr_agent" "sis" {
  type = "SIS"
  name = "example-sis"

  public_key_1 = file("${path.module}/agent_public_key.pem")
}

resource "altr_agent_task" "sis" {
  agent_id  = altr_agent.sis.id
  name      = "example-sis-audit"
  repo_name = altr_repo.postgres.name

  configuration = {
    audit_file_path = "/var/lib/postgresql/audit/*.json"
    audit_file_type = "json"
    log_line_prefix = "%m [%p] %q%u@%d "
  }

  schedule = {
    type  = "CRON"
    value = "*/5 * * * *"
  }
}
