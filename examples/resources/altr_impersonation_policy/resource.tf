# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_repo" "example" {
  name     = "example"
  type     = "Oracle"
  hostname = "example.com"
  port     = 1521
}

resource "altr_repo_user" "example" {
  repo_name = altr_repo.example.name
  username  = "example"

  aws_secrets_manager = {
    secrets_path = "arn:aws:secretsmanager:us-east-1:000000000000:secret:example-O3d19H"
  }
}

resource "altr_impersonation_policy" "test" {
  name        = "example"
  description = "Example impersonation policy"
  repo_name   = altr_repo.example.name

  rules = [
    {
      actors = [
        {
          type        = "idp_user"
          identifiers = ["user1@example.com"]
          condition   = "equals"
        }
      ]
      targets = [
        {
          type        = "repo_user"
          identifiers = [altr_repo_user.example.username]
          condition   = "equals"
        }
      ]
    }
  ]
}
