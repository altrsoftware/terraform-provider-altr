terraform {
  required_providers {
    aws = {
      source  = "altr"
      version = "~> 1.0"
    }
  }
}

# Configure the ALTR Provider
provider "altr" {
    api_key  = "my-api-key"
    base_url = "https://sc-control.live.altr.com"
    org_id   = "my-org-id"
    secret   = "my-secret"
}
