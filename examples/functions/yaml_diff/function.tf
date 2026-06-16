# Compare the values currently applied to a helm_release against the values
# that are about to replace them, and surface a compact unified diff at plan time.

locals {
  current_values = helm_release.this.metadata[0].values # YAML currently applied
  new_values     = yamlencode(var.helm_values)          # YAML about to be applied
}

output "helm_values_diff" {
  value = provider::lara-utils::yaml_diff(local.current_values, local.new_values)
}
