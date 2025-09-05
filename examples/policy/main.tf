terraform {
    required_providers {
        altr = {
        source = "altr"
        }
    }
}

provider "altr" {}

resource "altr_sidecar" "test" {
  name  = "test-sidecar"
  description     = "test description"
  hostname     = "localhost"
  unsupported_query_bypass = false
  public_key_1 = "-----BEGIN PUBLIC KEY-----\nMIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQB2VO3t0vpIEKFHbC8c3Ljf\n8b4dSthL7LHXsYtqEvnHWPVbDDJXyp8nVLEdcH6dM7lzOZncfhHHVzREgYgV8LRV\nrMEX4TgkmNFhHk/HasY56BwEU6Vy75L5i34xwkIIGFVNyO/DCJHqFKvCVfTv4tOP\n1Qk7T205Wa53f7lINKQp21BMncl1PjjN1TIqL4IKCZ4AWQB0tybZ6rojODKABzSU\n3ljF4GXkvJUPjfXEfylcarSjukQRurJEhd8vcxZlaVvMWXV2vOy85K/YS2s/BIMa\n9WC6qELXwaMAXVN2er7/U9e12emp/QYpZxjNpTemQCjSa+hFKb1CSK8Ezriypt43\nAgMBAAE=\n-----END PUBLIC KEY-----"
}

resource "altr_repo" "example" {
  name        = "example_repo"
  description = "An example repository"
  type        = "Oracle"
  hostname    = "example-host"
  port        = 5432
}

resource "altr_repo_user" "test_repo_user" {
    repo_name = altr_repo.example.name
    username = "testuser"
    aws_secrets_manager = {
        iam_role = "some_role"
        secrets_path = "some_path"
    }
}

resource "altr_impersonation_policy" "test" {
  name = "test-impersonation-policy-2"
  repo_name = altr_repo.example.name
  description = "Test impersonation policy"
  rules = [{
    actors = [{
        type = "idp_user",
        identifiers = ["test@altr.com"]
        condition = "equals"
    }],
    targets = [{
        type = "repo_user"
        identifiers = [altr_repo_user.test_repo_user.username]
        condition = "equals"
    }],
  }]
}

resource "altr_access_management_oltp_policy" "test" {
  name = "test-access-management-policy 3"
  description = "Test access management policy"
  case_sensitivity = "case_sensitive"
  database_type = 4
  database_type_name = "oracle"
  repo_name = altr_repo.example.name
  rules = [{
    type = "read"
    actors = [{
        type = "idp_user",
        identifiers = ["test@altr.com"]
        condition = "equals"
    }],
    objects = [{
        type = "column"
        identifiers = [{
            database = {
                name = "some-database"
                wildcard = false
            }
            schema = {
                name = "some-database"
                wildcard = false
            }
            table = {
                name = "some-database"
                wildcard = false
            }
            column = {
                name = "some-database"
                wildcard = false
            }
        }]
    }],
  }]
}

resource "altr_access_management_snowflake_policy" "test" {
  name = "test-access-management-policy 4"
  description = "Test access management policy day"
  connection_ids = [19] // This is the id of your snowflake data source, data sources are not currenly in the provider
   policy_maintenance = {
     rate = "day"
     value = 1
   }
  rules = [
    {
        actors = [{
            type = "role",
            identifiers = ["ACCOUNTADMIN"]
            condition = "equals"
        }],
        objects = [{
            type = "database"
            identifiers = ["MY_DB"]
            condition = "equals"
        }],
        access = [{
            name = "read"
        }]
    },
    {
        actors = [{
            type = "role",
            identifiers = ["ACCOUNTADMIN"]
            condition = "equals"
        }],
        objects = [{
            type = "table"
            fully_qualified_identifiers = [{
                database = "MY_DB"
                schema = "PUBLIC"
                table = "MY_TABLE"
            }]
            condition = "fully_qualified"
        }],
        access = [{
            name = "read"
        }]
    },
    {
        actors = [{
            type = "role",
            identifiers = ["ACCOUNTADMIN"]
            condition = "equals"
        }],
        access = [{
            name = "read"
        }],
        tagged_objects = [{
            check_against = ["databases"]
            tag_condition = "or"
            tagged_with = [{
                database = "MY_DB"
                schema = "PUBLIC"
                name = "sensitivity"
                value = "high"
            }]
        }]
    }
  ]
}
