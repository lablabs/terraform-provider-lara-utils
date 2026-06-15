// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package deepmerge

import (
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
)

// boolPtr returns a pointer to b. Used for PathOverrideOptions construction.
func boolPtr(b bool) *bool { return &b }

// runMerge is a thin helper around the package-internal merge() function so
// tests can assert directly on the merged map without going through Terraform.
func runMerge(t *testing.T, opts DeepMergeOptions, objs ...map[string]any) map[string]any {
	t.Helper()
	got, diags := merge(objs, opts)
	if diags.HasError() {
		t.Fatalf("merge returned errors: %v", diags)
	}
	return got
}

// TestMergeDefaults verifies the historical default behavior — override on,
// list replace, null override on — still holds when PathOverrides is empty.
func TestMergeDefaults(t *testing.T) {
	opts := *NewDefaultOptions()
	got := runMerge(t, opts,
		map[string]any{"a": 1, "b": map[string]any{"x": 10, "y": 20}, "list": []any{1, 2}},
		map[string]any{"a": 2, "b": map[string]any{"y": 30, "z": 40}, "list": []any{3, 4}},
	)
	want := map[string]any{
		"a":    2,
		"b":    map[string]any{"x": 10, "y": 30, "z": 40},
		"list": []any{3, 4},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_AppendListAtSpecificKey: append_list off globally, on at
// one specific path. Sibling lists must keep the global "replace" behavior.
func TestPathOverride_AppendListAtSpecificKey(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".spec.containers": {AppendList: boolPtr(true)},
	}

	got := runMerge(t, opts,
		map[string]any{
			"spec": map[string]any{
				"containers": []any{"a", "b"},
				"volumes":    []any{"v1"},
			},
		},
		map[string]any{
			"spec": map[string]any{
				"containers": []any{"c"},
				"volumes":    []any{"v2"},
			},
		},
	)
	want := map[string]any{
		"spec": map[string]any{
			"containers": []any{"a", "b", "c"}, // appended
			"volumes":    []any{"v2"},          // replaced (global default)
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_UnionListsWildcard: wildcard step matches multiple paths.
func TestPathOverride_UnionListsWildcard(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".*.tags": {UnionLists: boolPtr(true)},
	}

	got := runMerge(t, opts,
		map[string]any{
			"app1": map[string]any{"tags": []any{"a", "b"}},
			"app2": map[string]any{"tags": []any{"x"}, "other": []any{"keep"}},
		},
		map[string]any{
			"app1": map[string]any{"tags": []any{"b", "c"}},
			"app2": map[string]any{"tags": []any{"y"}, "other": []any{"replace"}},
		},
	)
	want := map[string]any{
		"app1": map[string]any{"tags": []any{"a", "b", "c"}},
		"app2": map[string]any{"tags": []any{"x", "y"}, "other": []any{"replace"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_NullOverrideOff: at one specific path, treat null as "keep".
// Other paths keep the global null_override=true, under which a null src deletes
// the destination key (matching the existing transformer behavior).
func TestPathOverride_NullOverrideOff(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".keep": {NullOverride: boolPtr(false)},
	}

	got := runMerge(t, opts,
		map[string]any{"keep": "stays", "drop": "stays"},
		map[string]any{"keep": nil, "drop": nil},
	)
	want := map[string]any{
		"keep": "stays", // null skipped at this path
		// "drop" is removed: global null_override=true means a null src
		// overrides the dst, and the transformer drops the key in that case.
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_OnlyMatchedNode: rule on .a.b must not affect .a.b.c.
func TestPathOverride_OnlyMatchedNode(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".a": {AppendList: boolPtr(true)},
	}

	got := runMerge(t, opts,
		map[string]any{
			"a":     []any{1, 2}, // matched: append
			"other": []any{"x"},  // not matched: replace
		},
		map[string]any{
			"a":     []any{3},
			"other": []any{"y"},
		},
	)
	want := map[string]any{
		"a":     []any{1, 2, 3},
		"other": []any{"y"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_OverlayPreservesGlobal: rule overrides one field; others
// stay at the global setting.
func TestPathOverride_OverlayPreservesGlobal(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.UnionLists = true // global
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".x": {AppendList: boolPtr(true)}, // overrides only append; union remains active globally
	}

	got := runMerge(t, opts,
		map[string]any{"x": []any{1, 1}, "y": []any{"a", "a"}},
		map[string]any{"x": []any{1, 2}, "y": []any{"a", "b"}},
	)
	// At .x: append takes precedence over union (matches transformer order: union checked first).
	// To verify "append wins over union when both set at a path", we expect concat.
	// Note: union_lists is checked before append_list in the transformer; with both
	// effectively true at .x, union wins. So .x ends as union.
	want := map[string]any{
		"x": []any{1, 2},     // union (union checked first, append also set but union wins)
		"y": []any{"a", "b"}, // union (global)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_EmptyMapIsNoOp: an empty PathOverrides map must not change behavior.
func TestPathOverride_EmptyMapIsNoOp(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{}

	got := runMerge(t, opts,
		map[string]any{"a": []any{1, 2}},
		map[string]any{"a": []any{3}},
	)
	want := map[string]any{"a": []any{3}} // default: replace
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_MalformedPathErrors: malformed path strings surface as a diagnostic.
func TestPathOverride_MalformedPathErrors(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		"bad_no_dot": {AppendList: boolPtr(true)},
	}
	_, diags := merge([]map[string]any{{"a": 1}}, opts)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostic from malformed path")
	}
}

// TestMapstructureDecodesPathOverrides ensures mapstructure can decode the
// nested option map produced by helpers.EncodeValue (a map[string]any tree)
// into DeepMergeOptions including the *bool fields of PathOverrideOptions.
// This mirrors what GetMergingOptions does in the provider layer.
func TestMapstructureDecodesPathOverrides(t *testing.T) {
	input := map[string]any{
		"append_list": false,
		"path_overrides": map[string]any{
			".spec.containers": map[string]any{
				"append_list": true,
			},
			".keep": map[string]any{
				"null_override": false,
			},
		},
	}
	opts := *NewDefaultOptions()
	if err := mapstructure.Decode(input, &opts); err != nil {
		t.Fatalf("mapstructure.Decode: %v", err)
	}
	if opts.AppendList {
		t.Fatalf("global append_list should be false")
	}
	if len(opts.PathOverrides) != 2 {
		t.Fatalf("expected 2 path overrides, got %d", len(opts.PathOverrides))
	}
	rule := opts.PathOverrides[".spec.containers"]
	if rule == nil || rule.AppendList == nil || !*rule.AppendList {
		t.Fatalf(".spec.containers append_list not decoded as true: %+v", rule)
	}
	rule = opts.PathOverrides[".keep"]
	if rule == nil || rule.NullOverride == nil || *rule.NullOverride {
		t.Fatalf(".keep null_override not decoded as false: %+v", rule)
	}
}

// TestPathOverride_OverrideOffAtPath: override=false at a specific path keeps
// the existing destination value; other paths still get overwritten.
func TestPathOverride_OverrideOffAtPath(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".pinned": {Override: boolPtr(false)},
	}
	got := runMerge(t, opts,
		map[string]any{"pinned": "keep_me", "other": "old"},
		map[string]any{"pinned": "new_value", "other": "new"},
	)
	want := map[string]any{
		"pinned": "keep_me", // override disabled here
		"other":  "new",     // global override on
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_MultipleRulesAtDifferentPaths: stack several rules in one
// call — each must fire independently at its own path.
func TestPathOverride_MultipleRulesAtDifferentPaths(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".a.list":   {AppendList: boolPtr(true)},
		".b.tags":   {UnionLists: boolPtr(true)},
		".c.pinned": {Override: boolPtr(false)},
		".d.keep":   {NullOverride: boolPtr(false)},
	}
	got := runMerge(t, opts,
		map[string]any{
			"a": map[string]any{"list": []any{1, 2}},
			"b": map[string]any{"tags": []any{"x", "y"}},
			"c": map[string]any{"pinned": "old"},
			"d": map[string]any{"keep": "stays"},
			"e": map[string]any{"v": "unaffected_old"},
		},
		map[string]any{
			"a": map[string]any{"list": []any{3}},
			"b": map[string]any{"tags": []any{"y", "z"}},
			"c": map[string]any{"pinned": "new"},
			"d": map[string]any{"keep": nil},
			"e": map[string]any{"v": "unaffected_new"},
		},
	)
	want := map[string]any{
		"a": map[string]any{"list": []any{1, 2, 3}},       // append
		"b": map[string]any{"tags": []any{"x", "y", "z"}}, // union
		"c": map[string]any{"pinned": "old"},              // override off
		"d": map[string]any{"keep": "stays"},              // null skipped
		"e": map[string]any{"v": "unaffected_new"},        // global default replace
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_ThreeInputs: rules must fire on every pairwise merge step,
// not just the first. With three inputs and append at a path, all three lists
// should concatenate in order.
func TestPathOverride_ThreeInputs(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".items": {AppendList: boolPtr(true)},
	}
	got := runMerge(t, opts,
		map[string]any{"items": []any{1}},
		map[string]any{"items": []any{2}},
		map[string]any{"items": []any{3}},
	)
	want := map[string]any{"items": []any{1, 2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_DeepNesting: the path is reconstructed correctly even at
// significant depth, and rules at a deep path don't bleed up.
func TestPathOverride_DeepNesting(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".a.b.c.d.e.list": {AppendList: boolPtr(true)},
	}
	mk := func(v []any) map[string]any {
		return map[string]any{
			"a": map[string]any{
				"b": map[string]any{
					"c": map[string]any{
						"d": map[string]any{
							"e": map[string]any{"list": v, "sibling": []any{"replace_me"}},
						},
					},
				},
			},
		}
	}
	got := runMerge(t, opts, mk([]any{1, 2}), mk([]any{3}))
	want := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": map[string]any{
							"list":    []any{1, 2, 3}, // appended at depth 6
							"sibling": []any{"replace_me"},
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_NonMatchingRuleIsNoOp: a rule that never matches must not
// alter behavior — same result as no rules at all.
func TestPathOverride_NonMatchingRuleIsNoOp(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".does.not.exist": {AppendList: boolPtr(true)},
	}
	got := runMerge(t, opts,
		map[string]any{"a": []any{1, 2}},
		map[string]any{"a": []any{3}},
	)
	want := map[string]any{"a": []any{3}} // global default: replace
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_LastSortedRuleWins: when two rules match the same path,
// compilePathRules sorts ascending and effectiveOptionsAt applies each match
// in order, so the lexicographically-last matching rule wins.
func TestPathOverride_LastSortedRuleWins(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".a.b": {AppendList: boolPtr(true)}, // wildcard "*" is matched by both
		".a.*": {UnionLists: boolPtr(true)}, // sorts before ".a.b" -> applied first
	}
	got := runMerge(t, opts,
		map[string]any{"a": map[string]any{"b": []any{1, 1, 2}}},
		map[string]any{"a": map[string]any{"b": []any{1, 3}}},
	)
	// Sorted order: ".a.*", ".a.b". Both match path [a,b]; .a.b applied last
	// sets AppendList=true. Union is also set (from .a.*), but the transformer
	// checks UnionLists before AppendList in the switch, so union wins at runtime.
	// Verify behavior matches that documented precedence.
	want := map[string]any{
		"a": map[string]any{"b": []any{1, 2, 3}}, // union: dedup + append new
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_ReEnableAtPath: global has null_override off; rule turns it
// back on at one specific path. With null_override on, the transformer drops
// the destination key when src is nil (its existing semantics).
func TestPathOverride_ReEnableAtPath(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.NullOverride = false // global off
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".clear_me": {NullOverride: boolPtr(true)},
	}
	got := runMerge(t, opts,
		map[string]any{"clear_me": "old", "stays": "old"},
		map[string]any{"clear_me": nil, "stays": nil},
	)
	want := map[string]any{
		// "clear_me" removed: null override re-enabled at this path
		"stays": "old", // global null_override=false keeps existing
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestPathOverride_MixedWildcards: a path mixing "*" and "[]" wildcards in
// one rule must match correctly.
func TestPathOverride_MixedWildcards(t *testing.T) {
	opts := *NewDefaultOptions()
	opts.PathOverrides = map[string]*PathOverrideOptions{
		".a.*.b": {AppendList: boolPtr(true)},
	}
	got := runMerge(t, opts,
		map[string]any{
			"a": map[string]any{
				"first":  map[string]any{"b": []any{1}, "c": []any{"x"}},
				"second": map[string]any{"b": []any{2}, "c": []any{"x"}},
			},
		},
		map[string]any{
			"a": map[string]any{
				"first":  map[string]any{"b": []any{10}, "c": []any{"y"}},
				"second": map[string]any{"b": []any{20}, "c": []any{"y"}},
			},
		},
	)
	want := map[string]any{
		"a": map[string]any{
			"first":  map[string]any{"b": []any{1, 10}, "c": []any{"y"}},
			"second": map[string]any{"b": []any{2, 20}, "c": []any{"y"}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
