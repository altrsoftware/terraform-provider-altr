# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_sidecar" "example" {
  name     = "example"
  hostname = "example.com"
}

resource "altr_sidecar_listener" "example_8080" {
  sidecar_id         = altr_sidecar.example.id
  port               = 8080
  database_type      = "Oracle"
  advertised_version = "19.0.0.0"
}

resource "altr_sidecar_listener" "example_9000" {
  sidecar_id         = altr_sidecar.example.id
  port               = 9000
  database_type      = "Oracle"
  advertised_version = "19.0.0.0"
}
