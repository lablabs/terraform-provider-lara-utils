Stores any HCL value in Terraform state so that Terraform's own plan engine diffs it natively — showing individual field changes like `~ cpu = "20m" -> "30m"` instead of a raw text blob.

Equivalent to the built-in `terraform_data` resource but without the redundant `output` attribute that bloats plan output on every update.

## Example

```hcl
resource "lara-utils_terraform_data" "config" {
  input = yamldecode(nonsensitive(var.values))
}
```

`terraform plan` will show field-level changes, e.g.:

```
  ~ resource "lara-utils_terraform_data" "config" {
      ~ input = {
          ~ resources = {
              ~ requests = {
                  ~ cpu = "20m" -> "30m"
                }
            }
        }
    }
```

## Forcing replacement

Use `triggers_replace` to force destroy+recreate when a value outside of `input` changes:

```hcl
resource "lara-utils_terraform_data" "config" {
  input            = yamldecode(nonsensitive(var.values))
  triggers_replace = var.environment
}
```

## Sensitive values

The `input` attribute is **not** marked sensitive. If your input is a Terraform sensitive value, wrap it with `nonsensitive()` before passing it in:

```hcl
resource "lara-utils_terraform_data" "config" {
  input = yamldecode(nonsensitive(var.values))  # intentional: no secrets in these values
}
```

Without `nonsensitive()`, Terraform propagates sensitivity to the `input` attribute and the diff will be redacted as `(sensitive value)` in the plan, defeating the purpose of the resource.
