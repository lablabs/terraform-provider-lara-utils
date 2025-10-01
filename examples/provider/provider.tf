# Copyright (c) Labyrinth Labs s.r.o.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    lara-utils = {
      source = "lablabs/lara-utils"
    }
  }
}

locals {
  map1 = {
    a = {
      x = [1, 2, 3]
      y = false
    }
    b = {
      s = "hello, world"
      n = 17
    }
  }
  map2 = {
    a = { x = [4, 5, 6] }
    b = { n = 42 }
  }

  merged = provider::lara-utils::deep_merge(local.map1, local.map2)
}
