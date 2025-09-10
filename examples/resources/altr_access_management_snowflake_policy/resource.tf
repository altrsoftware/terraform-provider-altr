# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_access_management_snowflake_policy" "example" {
  name           = "example"
  description    = "Example access management policy"
  connection_ids = [1]

  rules = [
    {
      actors = [{
        type        = "role"
        identifiers = ["ACCOUNTADMIN"]
        condition   = "equals"
      }],
      objects = [{
        type        = "database"
        identifiers = ["MY_DB"]
        condition   = "equals"
      }],
      access = [{
        name = "read"
      }]
    }
  ]
}
