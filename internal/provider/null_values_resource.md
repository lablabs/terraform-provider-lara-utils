Stores a structured values object in Terraform state so that Terraform's own plan engine can diff it natively — showing individual field changes like `~ cpu = "20m" -> "30m"` instead of a raw text blob.

Pass any HCL object via `yamldecode()` and Terraform handles the rest. No custom diff logic, no text parsing — the diff is produced by the same engine that diffs every other Terraform resource.

## Example

```hcl
resource "lara-utils_null_values" "values" {
  values = yamldecode(nonsensitive(var.values))
}
```

`terraform plan` will then show field-level changes, e.g.:

```
  ~ resource "lara-utils_null_values" "values" {
      ~ values = {
          ~ resources = {
              ~ requests = {
                  ~ cpu = "20m" -> "30m"
                }
            }
        }
    }
```

## Sensitive values

The `values` attribute is **not** marked sensitive. If your input is a Terraform sensitive value (e.g. a `variable` with `sensitive = true`), wrap it with `nonsensitive()` before passing to `yamldecode()`:

```hcl
resource "lara-utils_null_values" "values" {
  values = yamldecode(nonsensitive(var.values))  # intentional: no secrets in these helm values
}
```

Without `nonsensitive()`, Terraform propagates sensitivity to the `values` attribute and the diff will be redacted as `(sensitive value)` in the plan, defeating the purpose of the resource.
