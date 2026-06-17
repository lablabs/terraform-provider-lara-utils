// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// TestTerraformDataResource_create verifies a basic create: id is a UUID
// and input values are stored.
func TestTerraformDataResource_create(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "lara-utils_terraform_data" "test" {
						input = {
							replicas = 3
							image    = "nginx:1.25"
						}
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(uuidRegexp),
					),
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input").AtMapKey("image"),
						knownvalue.StringExact("nginx:1.25"),
					),
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input").AtMapKey("replicas"),
						knownvalue.Int64Exact(3),
					),
					},
			},
		},
	})
}

// TestTerraformDataResource_update verifies that changing input updates the
// resource in-place (no replace) and the id is preserved.
func TestTerraformDataResource_update(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "lara-utils_terraform_data" "test" {
						input = { replicas = 3 }
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("id"),
						knownvalue.StringRegexp(uuidRegexp),
					),
				},
			},
			{
				Config: `
					resource "lara-utils_terraform_data" "test" {
						input = { replicas = 5 }
					}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("lara-utils_terraform_data.test",
							plancheck.ResourceActionUpdate,
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input").AtMapKey("replicas"),
						knownvalue.Int64Exact(5),
					),
				},
			},
		},
	})
}

// TestTerraformDataResource_string_input verifies that input accepts a plain string.
func TestTerraformDataResource_string_input(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "lara-utils_terraform_data" "test" {
						input = "hello world"
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input"),
						knownvalue.StringExact("hello world"),
					),
				},
			},
		},
	})
}

// TestTerraformDataResource_nested_input verifies that deeply nested objects
// are stored and diffed correctly.
func TestTerraformDataResource_nested_input(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "lara-utils_terraform_data" "test" {
						input = {
							resources = {
								requests = { cpu = "100m", memory = "128Mi" }
								limits   = { cpu = "500m", memory = "256Mi" }
							}
						}
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input").AtMapKey("resources").AtMapKey("requests").AtMapKey("cpu"),
						knownvalue.StringExact("100m"),
					),
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input").AtMapKey("resources").AtMapKey("limits").AtMapKey("memory"),
						knownvalue.StringExact("256Mi"),
					),
				},
			},
			{
				Config: `
					resource "lara-utils_terraform_data" "test" {
						input = {
							resources = {
								requests = { cpu = "200m", memory = "128Mi" }
								limits   = { cpu = "500m", memory = "256Mi" }
							}
						}
					}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("lara-utils_terraform_data.test",
							plancheck.ResourceActionUpdate,
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("lara-utils_terraform_data.test",
						tfjsonpath.New("input").AtMapKey("resources").AtMapKey("requests").AtMapKey("cpu"),
						knownvalue.StringExact("200m"),
					),
				},
			},
		},
	})
}
