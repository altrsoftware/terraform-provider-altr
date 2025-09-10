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

resource "altr_repo_user" "other" {
  repo_name = altr_repo.example.name
  username  = "other"

  aws_secrets_manager = {
    secrets_path = "arn:aws:secretsmanager:us-east-1:000000000000:secret:other-12c26f"
  }
}
