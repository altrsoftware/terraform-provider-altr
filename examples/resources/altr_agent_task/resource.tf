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
    collection_name = "default"
  }

  schedule = {
    type         = "CRON"
    value        = "0 0 * * *"
    max_duration = "PT30M"
  }
}
