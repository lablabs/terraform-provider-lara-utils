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

  merged = provider::lara-utils::deep_merge([local.map1, local.map2])

  merged_with_options = provider::lara-utils::deep_merge([local.map1, local.map2], { append_list = true })

  merged_yaml = provider::lara-utils::yaml_deep_merge([yamlencode(local.map1), yamlencode(local.map2)])
}
