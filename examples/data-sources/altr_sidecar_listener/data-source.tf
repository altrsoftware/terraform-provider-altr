# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

data "altr_sidecar_listener" "example_9000" {
  sidecar_id = "682bfaa1-77e1-40aa-897f-1a0469f4ac64"
  port       = 9000
}
