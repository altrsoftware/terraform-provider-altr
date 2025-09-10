# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    altr = {
      source  = "altrsoftware/altr"
      version = "~> 1.0"
    }
  }
}

# Configure the ALTR Provider
provider "altr" {
  api_key  = "api-key"
  base_url = "https://org-id.altrnet.live.altr.com"
  org_id   = "org-id"
  secret   = "api-secret"
}
