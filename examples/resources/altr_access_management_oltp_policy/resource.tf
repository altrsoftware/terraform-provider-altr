# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_repo" "example" {
  name     = "example"
  type     = "Oracle"
  hostname = "example.com"
  port     = 1521
}


resource "altr_access_management_oltp_policy" "example" {
  name               = "example"
  description        = "Example policy"
  case_sensitivity   = "case_sensitive"
  database_type      = "4"
  database_type_name = "oracle"
  repo_name          = altr_repo.example.name
  rules = [{
    type = "read"
    actors = [{
      type        = "idp_user",
      identifiers = ["example@example.com"],
      condition   = "equals",
    }],
    objects = [{
      type = "column",
      identifiers = [{
        database = {
          name     = "exampledb"
          wildcard = false
        }
        schema = {
          name     = public
          wildcard = false
        }
        table = {
          name     = "employees"
          wildcard = false
        }
        column = {
          name     = "salary"
          wildcard = false
        }
      }]
    }],
  }]
}
