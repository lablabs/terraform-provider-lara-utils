schema_version = 1

project {
  license          = "MPL-2.0"
  copyright_holder = "Labyrinth Labs s.r.o."
  copyright_year   = 2025
  upstream         = "hashicorp/terraform-provider-scaffolding-framework"

  header_ignore = [
    "examples/**",
    ".github/ISSUE_TEMPLATE/*.yml",
    ".golangci.yml",
    ".goreleaser.yml",
  ]
}
