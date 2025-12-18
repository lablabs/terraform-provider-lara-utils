// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package testdata

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type DeepMergeTestConfig struct {
	Yaml bool
}

type DeepMergeTestOption func(*DeepMergeTestConfig)

func WithYaml() DeepMergeTestOption {
	return func(c *DeepMergeTestConfig) {
		c.Yaml = true
	}
}

func NewDeepMergeTestOptions(opts ...DeepMergeTestOption) DeepMergeTestConfig {
	cfg := DeepMergeTestConfig{
		Yaml: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func providerFunctionCall(cfg DeepMergeTestConfig, variables []string, options string) string {
	yamldecode := ""
	function := "deep_merge"

	if cfg.Yaml {
		yamldecode = "yamldecode"
		function = "yaml_deep_merge"

		for i, v := range variables {
			variables[i] = "yamlencode(" + v + ")"
		}
	}

	return fmt.Sprintf("%s(provider::lara-utils::%s([%s], %s))", yamldecode, function, strings.Join(variables, ","), options)
}

func TestDeepMergeFunction_Default(cfg DeepMergeTestConfig) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: `
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{}") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownOutputValue("test", knownvalue.MapExact(map[string]knownvalue.Check{
				"x1": knownvalue.MapExact(map[string]knownvalue.Check{
					"y1": knownvalue.Bool(false),
					"y2": knownvalue.Int64Exact(2),
					"y3": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.Int64Exact(4),
						knownvalue.Int64Exact(5),
						knownvalue.Int64Exact(6),
					}),
				}),
				"x2": knownvalue.MapExact(map[string]knownvalue.Check{
					"y4": knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("hello"),
						"b": knownvalue.StringExact("mergo"),
						"c": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("a"),
							knownvalue.Int64Exact(2),
							knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("b"),
							}),
						}),
					}),
				}),
			}))},
		},
		{
			Config: `
				locals {
					map1 = {
						a = null
						b = "foo"
					}
					map2 = {
						a = "bar"
						b = "baz"
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{}") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("bar"),
						"b": knownvalue.StringExact("baz"),
					}),
				),
			},
		},
		{
			Config: `
				locals {
					map1 = {
						a = null
						b = "foo"
					}
					map2 = {
						a = "bar"
						b = "baz"
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{}") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("bar"),
						"b": knownvalue.StringExact("baz"),
					}),
				),
			},
		},
		{
			Config: `
				variable "obj" {
					type = object({
						x1 = any
						x2 = any
					})
					default = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}

					}
				}
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "var.obj"}, "{}") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.Bool(false),
							"y2": knownvalue.Int64Exact(2),
							"y3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(4),
								knownvalue.Int64Exact(5),
								knownvalue.Int64Exact(6),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"y4": knownvalue.MapExact(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("hello"),
								"b": knownvalue.StringExact("mergo"),
								"c": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.Int64Exact(2),
									knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("b"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
	}
}

func TestDeepMergeFunction_NoOverride(cfg DeepMergeTestConfig) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: `
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ override = false }") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.Bool(true),
							"y2": knownvalue.Int64Exact(1),
							"y3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"y4": knownvalue.MapExact(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("hello"),
								"b": knownvalue.StringExact("world"),
								"c": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.Int64Exact(2),
									knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("b"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
	}
}

func TestDeepMergeFunction_NoNullOverride(cfg DeepMergeTestConfig) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: `
					locals {
						map1 = {
							a = "foo"
							b = "bar"
						}
						map2 = {
							b = "bam"
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("foo"),
						"b": knownvalue.StringExact("bam"),
					}),
				),
			},
		},
		{
			Config: `
				locals {
					map1 = {
						a = "foo"
						b = "bar"
					}
					map2 = {}
				}
				output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ null_override = false }") + `
				}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("foo"),
						"b": knownvalue.StringExact("bar"),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							x1 = {
								y1 = false
								y2 = 1
							}
						}
						map2 = {
							x1 = {
								y2 = 1
								y3 = [4, 5, 6]
							}
							x2 = {
								y4 = {
									a = "hello"
									b = "world"
								}
							}
						}
						map3 = {
							x1 = {
								y1 = true
								y3 = [1, 2, 3]
							}
							x2 = {
								y4 = {
									b = "mergo"
									c = ["a", 2, ["b"]]
								}
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.Bool(true),
							"y2": knownvalue.Int64Exact(1),
							"y3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"y4": knownvalue.MapExact(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("hello"),
								"b": knownvalue.StringExact("mergo"),
								"c": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.Int64Exact(2),
									knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("b"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							a = "foo"
							b = "bar"
						}
						map2 = {
							a = null
							b = "bam"
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("foo"),
						"b": knownvalue.StringExact("bam"),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							a = "foo"
							b = "bar"
						}
						map2 = {
							a = null
							b = null
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"a": knownvalue.StringExact("foo"),
						"b": knownvalue.StringExact("bar"),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							x1 = {
								y1 = false
								y2 = 1
							}
						}
						map2 = {
							x1 = {
								y2 = 1
								y3 = [4, 5, 6]
							}
							x2 = {
								y4 = {
									a = "hello"
									b = "world"
								}
							}
						}
						map3 = {
							x1 = {
								y1 = true
								y3 = [1, 2, 3]
							}
							x2 = {
								y4 = {
									b = "mergo"
									c = ["a", 2, ["b"]]
								}
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.Bool(true),
							"y2": knownvalue.Int64Exact(1),
							"y3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"y4": knownvalue.MapExact(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("hello"),
								"b": knownvalue.StringExact("mergo"),
								"c": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.Int64Exact(2),
									knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("b"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							x1 = {
								y2 = {
									z1 = 1
									z2 = 2
								}
							}
							x2 = {
								s1 = "hello"
								s2 = "world"
								s3 = null
							}
							x3 = {
								t1 = "foo"
							}
						}
						map2 = {
							x1 = {
								y2 = {
									z2 = null
									z3 = 3
								}
								y2 = "bar"
							}
							x3 = {
								t1 = null
							}
						}
						map3 = {
							x1 = {
								y2 = null
							}
						}
						map4 = {
							x1 = {
								y1 = 4
							}
						}
						map5 = {
							x1 = {
								y2 = {
									z1 = {
										n1 = 1
										n2 = {
											m1 = 1
										}
									}
									z2 = null
									z3 = null
									z4 = 4
								}
							}
							x2 = {
								s2 = "mergo"
								s4 = "today"
							}
							x3 = {
								t2 = "foz"
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3", "local.map4", "local.map5"}, "{ null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.Int64Exact(4),
							"y2": knownvalue.MapExact(map[string]knownvalue.Check{
								"z1": knownvalue.MapExact(map[string]knownvalue.Check{
									"n1": knownvalue.Int64Exact(1),
									"n2": knownvalue.MapExact(map[string]knownvalue.Check{
										"m1": knownvalue.Int64Exact(1),
									}),
								}),
								"z2": knownvalue.Null(),
								"z3": knownvalue.Null(),
								"z4": knownvalue.Int64Exact(4),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"s1": knownvalue.StringExact("hello"),
							"s2": knownvalue.StringExact("mergo"),
							"s3": knownvalue.Null(),
							"s4": knownvalue.StringExact("today"),
						}),
						"x3": knownvalue.MapExact(map[string]knownvalue.Check{
							"t1": knownvalue.StringExact("foo"),
							"t2": knownvalue.StringExact("foz"),
						}),
					}),
				),
			},
		},
	}
}

func TestDeepMergeFunction_AppendList(cfg DeepMergeTestConfig) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: `
				locals {
					map1 = {
						x1 = {
							y1 = true
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = false
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ append_list = true }") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.Bool(false),
							"y2": knownvalue.Int64Exact(2),
							"y3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
								knownvalue.Int64Exact(4),
								knownvalue.Int64Exact(5),
								knownvalue.Int64Exact(6),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"y4": knownvalue.MapExact(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("hello"),
								"b": knownvalue.StringExact("mergo"),
								"c": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.Int64Exact(2),
									knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("b"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
				locals {
					map1 = {
						x1 = {
							y1 = "foo"
							y2 = 1
						}
					}
					map2 = {
						x1 = {
							y2 = 2
							y3 = [1, 2, 3]
						}
						x2 = {
							y4 = {
								a = "hello"
								b = "world"
							}
						}
					}
					map3 = {
						x1 = {
							y1 = null
							y3 = [4, 5, 6]
						}
						x2 = {
							y4 = {
								b = "mergo"
								c = ["a", 2, ["b"]]
							}
						}
					}
				}
				output "test" {
					value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ append_list = true, null_override = false }") + `
				}
			`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.StringExact("foo"),
							"y2": knownvalue.Int64Exact(2),
							"y3": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.Int64Exact(1),
								knownvalue.Int64Exact(2),
								knownvalue.Int64Exact(3),
								knownvalue.Int64Exact(4),
								knownvalue.Int64Exact(5),
								knownvalue.Int64Exact(6),
							}),
						}),
						"x2": knownvalue.MapExact(map[string]knownvalue.Check{
							"y4": knownvalue.MapExact(map[string]knownvalue.Check{
								"a": knownvalue.StringExact("hello"),
								"b": knownvalue.StringExact("mergo"),
								"c": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.Int64Exact(2),
									knownvalue.ListExact([]knownvalue.Check{
										knownvalue.StringExact("b"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
	}
}

func TestDeepMergeFunction_DeepCopyList(cfg DeepMergeTestConfig) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: `
					locals {
						map1 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y1 = "foo"
											y2 = "foo"
										}
									}
								]
							}
						}
						map2 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y1 = "bar"
										}
									}
								]
							}
						}
						map3 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y3 = "baz"
										}
									}
								]
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ deep_copy_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"x2": knownvalue.MapExact(map[string]knownvalue.Check{
										"y1": knownvalue.StringExact("bar"),
										"y2": knownvalue.StringExact("foo"),
										"y3": knownvalue.StringExact("baz"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y1 = "foo"
											y2 = ["foo"]
										}
									}
								]
							}
						}
						map2 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y1 = "bar"
											y2 = ["bar"]
										}
									}
								]
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ deep_copy_list = true, append_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"x2": knownvalue.MapExact(map[string]knownvalue.Check{
										"y1": knownvalue.StringExact("foo"),
										"y2": knownvalue.ListExact([]knownvalue.Check{
											knownvalue.StringExact("foo"),
										}),
									}),
								}),
								knownvalue.MapExact(map[string]knownvalue.Check{
									"x2": knownvalue.MapExact(map[string]knownvalue.Check{
										"y1": knownvalue.StringExact("bar"),
										"y2": knownvalue.ListExact([]knownvalue.Check{
											knownvalue.StringExact("bar"),
										}),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y1 = "foo"
											y2 = "bar"
										}
									}
								]
							}
						}
						map2 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y1 = "baz"
										}
									}
								]
							}
						}
						map3 = {
							x1 = {
								y1 = [
									{
										x2 = {
											y3 = "quux"
										}
									}
								]
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ deep_copy_list = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"x1": knownvalue.MapExact(map[string]knownvalue.Check{
							"y1": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.MapExact(map[string]knownvalue.Check{
									"x2": knownvalue.MapExact(map[string]knownvalue.Check{
										"y3": knownvalue.StringExact("quux"),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							items = [
								{ id = 1, name = "item1", tags = ["a", "b"] },
								{ id = 2, name = "item2", tags = ["c"] }
							]
						}
						map2 = {
							items = [
								{ id = 1, status = "active" },
								{ id = 2, status = "inactive" }
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ deep_copy_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"items": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.MapExact(map[string]knownvalue.Check{
								"id":   knownvalue.Int64Exact(1),
								"name": knownvalue.StringExact("item1"),
								"tags": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("a"),
									knownvalue.StringExact("b"),
								}),
								"status": knownvalue.StringExact("active"),
							}),
							knownvalue.MapExact(map[string]knownvalue.Check{
								"id":   knownvalue.Int64Exact(2),
								"name": knownvalue.StringExact("item2"),
								"tags": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("c"),
								}),
								"status": knownvalue.StringExact("inactive"),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							nested = [
								{
									level1 = {
										level2 = [
											{ value = "a" }
										]
									}
								}
							]
						}
						map2 = {
							nested = [
								{
									level1 = {
										level2 = [
											{ value = "b", extra = "field" }
										]
									}
								}
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ deep_copy_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"nested": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.MapExact(map[string]knownvalue.Check{
								"level1": knownvalue.MapExact(map[string]knownvalue.Check{
									"level2": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.MapExact(map[string]knownvalue.Check{
											"value": knownvalue.StringExact("b"),
											"extra": knownvalue.StringExact("field"),
										}),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							data = [
								{ key = "x", nested = { a = 1, b = 2 } }
							]
						}
						map2 = {
							data = [
								{ key = "x", nested = { b = 3, c = 4 } }
							]
						}
						map3 = {
							data = [
								{ key = "x", nested = { d = 5 }, extra = "value" }
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2", "local.map3"}, "{ deep_copy_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"data": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.MapExact(map[string]knownvalue.Check{
								"key": knownvalue.StringExact("x"),
								"nested": knownvalue.MapExact(map[string]knownvalue.Check{
									"a": knownvalue.Int64Exact(1),
									"b": knownvalue.Int64Exact(3),
									"c": knownvalue.Int64Exact(4),
									"d": knownvalue.Int64Exact(5),
								}),
								"extra": knownvalue.StringExact("value"),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							config = [
								{
									settings = {
										feature1 = { enabled = true, value = 10 }
										feature2 = { enabled = false }
									}
								}
							]
						}
						map2 = {
							config = [
								{
									settings = {
										feature1 = { value = 20, priority = "high" }
										feature3 = { enabled = true }
									}
								}
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ deep_copy_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"config": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.MapExact(map[string]knownvalue.Check{
								"settings": knownvalue.MapExact(map[string]knownvalue.Check{
									"feature1": knownvalue.MapExact(map[string]knownvalue.Check{
										"enabled":  knownvalue.Bool(true),
										"value":    knownvalue.Int64Exact(20),
										"priority": knownvalue.StringExact("high"),
									}),
									"feature2": knownvalue.MapExact(map[string]knownvalue.Check{
										"enabled": knownvalue.Bool(false),
									}),
									"feature3": knownvalue.MapExact(map[string]knownvalue.Check{
										"enabled": knownvalue.Bool(true),
									}),
								}),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							arr = [
								{ type = "string", values = ["a", "b"] }
							]
						}
						map2 = {
							arr = [
								{ type = "string", values = ["c", "d"], count = 4 }
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ deep_copy_list = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"arr": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.MapExact(map[string]knownvalue.Check{
								"type": knownvalue.StringExact("string"),
								"values": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("c"),
									knownvalue.StringExact("d"),
								}),
								"count": knownvalue.Int64Exact(4),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							resources = [
								{ name = "res1", props = { size = "small", region = "us-east" } },
								{ name = "res2", props = { size = "large" } }
							]
						}
						map2 = {
							resources = [
								{ name = "res1", props = { region = "us-west", tier = "premium" } }
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ deep_copy_list = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"resources": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.MapExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("res1"),
								"props": knownvalue.MapExact(map[string]knownvalue.Check{
									"size":   knownvalue.StringExact("small"),
									"region": knownvalue.StringExact("us-west"),
									"tier":   knownvalue.StringExact("premium"),
								}),
							}),
							knownvalue.MapExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("res2"),
								"props": knownvalue.MapExact(map[string]knownvalue.Check{
									"size": knownvalue.StringExact("large"),
								}),
							}),
						}),
					}),
				),
			},
		},
	}
}

func TestDeepMergeFunction_UnionLists(cfg DeepMergeTestConfig) []resource.TestStep {
	return []resource.TestStep{
		{
			Config: `
					locals {
						map1 = {
							tags = ["app", "terraform", "prod"]
							ports = [80, 443, 22]
							features = {
								security = ["ssl", "firewall"]
								monitoring = ["logs", "metrics"]
							}
						}
						map2 = {
							tags = ["monitoring", "prod", "security"]
							ports = [8080, 80, 9090]
							features = {
								security = ["encryption", "ssl"]
								monitoring = ["alerts", "metrics"]
								caching = ["redis"]
							}
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ union_lists = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"tags": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("app"),
							knownvalue.StringExact("terraform"),
							knownvalue.StringExact("prod"),
							knownvalue.StringExact("monitoring"),
							knownvalue.StringExact("security"),
						}),
						"ports": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(80),
							knownvalue.Int64Exact(443),
							knownvalue.Int64Exact(22),
							knownvalue.Int64Exact(8080),
							knownvalue.Int64Exact(9090),
						}),
						"features": knownvalue.MapExact(map[string]knownvalue.Check{
							"security": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("ssl"),
								knownvalue.StringExact("firewall"),
								knownvalue.StringExact("encryption"),
							}),
							"monitoring": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("logs"),
								knownvalue.StringExact("metrics"),
								knownvalue.StringExact("alerts"),
							}),
							"caching": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("redis"),
							}),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						base = {
							allowed_cidrs = ["10.0.0.0/8", "192.168.1.0/24"]
							environments = ["dev", "staging"]
						}
						overrides = {
							allowed_cidrs = ["172.16.0.0/12", "10.0.0.0/8"]
							environments = ["prod", "staging"]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.base", "local.overrides"}, "{ union_lists = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"allowed_cidrs": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("10.0.0.0/8"),
							knownvalue.StringExact("192.168.1.0/24"),
							knownvalue.StringExact("172.16.0.0/12"),
						}),
						"environments": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("dev"),
							knownvalue.StringExact("staging"),
							knownvalue.StringExact("prod"),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						map1 = {
							mixed_types = [1, "string", true]
							numbers = [1, 2, 3]
						}
						map2 = {
							mixed_types = [true, "another", 1]
							numbers = [3, 4, 5]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.map1", "local.map2"}, "{ union_lists = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"mixed_types": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.StringExact("string"),
							knownvalue.Bool(true),
							knownvalue.StringExact("another"),
						}),
						"numbers": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(2),
							knownvalue.Int64Exact(3),
							knownvalue.Int64Exact(4),
							knownvalue.Int64Exact(5),
						}),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						base = {
							tags = ["app", "base"]
							important_setting = "replace_this"
						}
						overrides = {
							tags = ["override", "app"]
							important_setting = null
							new_field = "added"
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.base", "local.overrides"}, "{ union_lists = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"tags": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("app"),
							knownvalue.StringExact("base"),
							knownvalue.StringExact("override"),
						}),
						"new_field": knownvalue.StringExact("added"),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						base = {
							tags = ["app", "base"]
							important_setting = "keep_this"
						}
						overrides = {
							tags = ["override", "app"]
							important_setting = null
							new_field = "added"
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.base", "local.overrides"}, "{ union_lists = true, null_override = false }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"tags": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("app"),
							knownvalue.StringExact("base"),
							knownvalue.StringExact("override"),
						}),
						"important_setting": knownvalue.StringExact("keep_this"),
						"new_field":         knownvalue.StringExact("added"),
					}),
				),
			},
		},
		{
			Config: `
					locals {
						base = {
							nested_lists = [
								["a", "b"],
								["c", "d"]
							]
						}
						additional = {
							nested_lists = [
								["a", "b"],
								["e", "f"]
							]
						}
					}
					output "test" {
						value = ` + providerFunctionCall(cfg, []string{"local.base", "local.additional"}, "{ union_lists = true }") + `
					}
				`,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownOutputValue("test",
					knownvalue.MapExact(map[string]knownvalue.Check{
						"nested_lists": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
							}),
							knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("c"),
								knownvalue.StringExact("d"),
							}),
							knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("e"),
								knownvalue.StringExact("f"),
							}),
						}),
					}),
				),
			},
		},
	}
}
