# Copyright (c) ALTR Solutions, Inc.
# SPDX-License-Identifier: Apache-2.0

resource "altr_agent" "example" {
  type        = "CLASSIFIER"
  name        = "example"
  description = "Example classifier agent"

  public_key_1 = file("${path.module}/agent_public_key.pem")
}
