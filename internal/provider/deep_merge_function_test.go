// Copyright Labyrinth Labs s.r.o. 2025, 2026
// SPDX-License-Identifier: MPL-2.0

// Originally copied from https://github.com/isometry/terraform-provider-deepmerge/blob/main/internal/provider/mergo_function_test.go

package provider

import (
	"regexp"

	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/lablabs/terraform-provider-lara-utils/internal/provider/testdata"
)

func TestDeepMergeFunction_Default(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_Default(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_NoOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_NoOverride(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_NoNullOverride(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_NoNullOverride(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_NullRemove(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_NullRemove(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_OptionComposition(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_OptionComposition(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_AppendList(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_AppendList(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_DeepCopyList(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_DeepCopyList(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_UnionLists(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    testdata.TestDeepMergeFunction_UnionLists(testdata.NewDeepMergeTestOptions()),
	})
}

func TestDeepMergeFunction_Null(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge(null)
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: argument must not be null.`),
			},
			{
				Config: `
					variable "null_list" {
						type    = list(any)
						default = null
					}
					output "test" {
						value = provider::lara-utils::deep_merge(var.null_list)
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: argument must not be null.`),
			},
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge([], null)
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "options" parameter: argument must not be null.`),
			},
			{
				Config: `
					variable "null_object" {
						type    = object({
							append_list = bool
						})
						default = null
					}
					output "test" {
						value = provider::lara-utils::deep_merge([], var.null_object)
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "options" parameter: argument must not be null.`),
			},
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge([])
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test",
						knownvalue.MapSizeExact(0),
					),
				},
			},
		},
	})
}

func TestDeepMergeFunction_InvalidType(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge(true)
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: list of objects required.`),
			},
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge(99.9)
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: list of objects required.`),
			},
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge(["a", "b", "c"])
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: merging argument \d+ must be object`),
			},
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge(tolist(["a", "b", "c"]))
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: merging argument \d+ must be object`),
			},
			{
				Config: `
					output "test" {
						value = provider::lara-utils::deep_merge(toset(["a", "b", "c"]))
					}
				`,
				ExpectError: regexp.MustCompile(`Invalid value for "objects" parameter: merging argument \d+ must be object`),
			},
		},
	})
}
