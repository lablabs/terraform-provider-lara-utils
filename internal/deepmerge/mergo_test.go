// Copyright Labyrinth Labs s.r.o. 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package deepmerge

import (
	"reflect"
	"testing"

	"dario.cat/mergo"
)

// nativeDeepCopy runs a plain mergo merge using WithOverride + WithSliceDeepCopy,
// i.e. exactly what deep_copy_list relies on when no custom transformer is
// installed. It is the reference implementation deepCopySlices must match.
func nativeDeepCopy(t *testing.T, objs []map[string]any) map[string]any {
	t.Helper()

	dst := map[string]any{}
	for _, m := range objs {
		if err := mergo.Merge(&dst, m, mergo.WithOverride, mergo.WithSliceDeepCopy); err != nil {
			t.Fatalf("native mergo merge failed: %v", err)
		}
	}

	return dst
}

// TestDeepCopySlicesParity asserts that our hand-rolled deepCopySlices produces
// results identical to mergo's native WithSliceDeepCopy for null-free inputs.
// The transformer is installed via null_override=false (which does not affect
// null-free inputs) so that our slice handling, rather than mergo's, runs.
func TestDeepCopySlicesParity(t *testing.T) {
	tests := []struct {
		name string
		objs []map[string]any
	}{
		{
			name: "overlapping map elements merge, extra dst element preserved",
			objs: []map[string]any{
				{"items": []any{map[string]any{"id": 1, "name": "a"}, map[string]any{"id": 2}}},
				{"items": []any{map[string]any{"id": 1, "status": "on"}}},
			},
		},
		{
			name: "nested maps inside list elements merge",
			objs: []map[string]any{
				{"a": []any{map[string]any{"x": map[string]any{"p": 1}}}},
				{"a": []any{map[string]any{"x": map[string]any{"q": 2}}}},
			},
		},
		{
			name: "src longer than dst: extra src element dropped",
			objs: []map[string]any{
				{"l": []any{map[string]any{"a": 1}}},
				{"l": []any{map[string]any{"b": 2}, map[string]any{"c": 3}}},
			},
		},
		{
			name: "dst longer than src: extra dst element preserved",
			objs: []map[string]any{
				{"l": []any{map[string]any{"a": 1}, map[string]any{"z": 9}}},
				{"l": []any{map[string]any{"b": 2}}},
			},
		},
		{
			name: "scalar elements: dst preserved, src ignored",
			objs: []map[string]any{
				{"l": []any{1, 2, 3}},
				{"l": []any{9}},
			},
		},
		{
			name: "zero-value scalar element still keeps dst",
			objs: []map[string]any{
				{"l": []any{0, "", 3}},
				{"l": []any{9, "x"}},
			},
		},
		{
			name: "three layers merged element by element",
			objs: []map[string]any{
				{"d": []any{map[string]any{"a": 1}}},
				{"d": []any{map[string]any{"b": 2}}},
				{"d": []any{map[string]any{"c": 3}}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := nativeDeepCopy(t, tc.objs)

			opts := *NewDefaultOptions()
			opts.DeepCopyList = true
			opts.NullOverride = false // installs the transformer without changing null-free behavior.

			got, diags := merge(tc.objs, opts)
			if diags.HasError() {
				t.Fatalf("merge returned diags: %v", diags)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("deepCopySlices diverged from native mergo\n got:  %#v\n want: %#v", got, want)
			}
		})
	}
}

// TestDeepCopyListNullRemoveInElements documents the deliberate difference from
// native mergo: null_remove is applied inside deep-copied list elements too, so
// a null nested in a list element removes that key rather than leaking through.
func TestDeepCopyListNullRemoveInElements(t *testing.T) {
	objs := []map[string]any{
		{"items": []any{map[string]any{"a": 1, "b": 2}}},
		{"items": []any{map[string]any{"b": nil, "c": 3}}},
	}

	opts := *NewDefaultOptions()
	opts.DeepCopyList = true
	opts.NullRemove = true

	got, diags := merge(objs, opts)
	if diags.HasError() {
		t.Fatalf("merge returned diags: %v", diags)
	}

	want := map[string]any{"items": []any{map[string]any{"a": 1, "c": 3}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("null_remove not applied inside list element\n got:  %#v\n want: %#v", got, want)
	}
}

// mustMerge runs merge and fails the test on any diagnostics, returning the
// merged result. It keeps the option-focused tests below free of diag plumbing.
func mustMerge(t *testing.T, objs []map[string]any, opts DeepMergeOptions) map[string]any {
	t.Helper()

	got, diags := merge(objs, opts)
	if diags.HasError() {
		t.Fatalf("merge returned diags: %v", diags)
	}

	return got
}

// TestMergeUnionLists covers the union_lists branch (mergo.go:103-104) plus
// unionSlices and containsElement, asserting set semantics: order is dst-first
// then new src elements, duplicates are dropped (including duplicates already
// present within dst), and equality spans mixed scalar types.
func TestMergeUnionLists(t *testing.T) {
	tests := []struct {
		name string
		objs []map[string]any
		want map[string]any
	}{
		{
			name: "string lists dedup, dst order then new src",
			objs: []map[string]any{
				{"tags": []any{"a", "b", "c"}},
				{"tags": []any{"b", "c", "d"}},
			},
			want: map[string]any{"tags": []any{"a", "b", "c", "d"}},
		},
		{
			name: "number lists dedup",
			objs: []map[string]any{
				{"ports": []any{1, 2, 3}},
				{"ports": []any{3, 4, 5}},
			},
			want: map[string]any{"ports": []any{1, 2, 3, 4, 5}},
		},
		{
			name: "mixed scalar types deduped by value",
			objs: []map[string]any{
				{"mixed": []any{1, "x", true}},
				{"mixed": []any{true, "y", 1}},
			},
			want: map[string]any{"mixed": []any{1, "x", true, "y"}},
		},
		{
			name: "duplicates within dst are collapsed",
			objs: []map[string]any{
				{"l": []any{"a", "a", "b"}},
				{"l": []any{"a"}},
			},
			want: map[string]any{"l": []any{"a", "b"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := *NewDefaultOptions()
			opts.UnionLists = true

			got := mustMerge(t, tc.objs, opts)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("union_lists mismatch\n got:  %#v\n want: %#v", got, tc.want)
			}
		})
	}
}

// TestMergeNullHandling covers the null-source branches of mergeMaps. With
// null_override disabled the existing value is kept (mergo.go:87-88); with the
// transformer installed via union_lists and null_override left at its default,
// a null drops the key (mergo.go:89-90).
func TestMergeNullHandling(t *testing.T) {
	t.Run("null_override=false keeps existing value", func(t *testing.T) {
		objs := []map[string]any{
			{"a": "x", "keep": "y"},
			{"a": nil},
		}

		opts := *NewDefaultOptions()
		opts.NullOverride = false

		got := mustMerge(t, objs, opts)
		want := map[string]any{"a": "x", "keep": "y"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("null did not preserve existing value\n got:  %#v\n want: %#v", got, want)
		}
	})

	t.Run("null_override default drops key", func(t *testing.T) {
		objs := []map[string]any{
			{"a": "x", "keep": "y"},
			{"a": nil},
		}

		// union_lists installs the transformer while null_override stays true,
		// exercising the default (drop-key) null branch.
		opts := *NewDefaultOptions()
		opts.UnionLists = true

		got := mustMerge(t, objs, opts)
		want := map[string]any{"keep": "y"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("null did not drop key\n got:  %#v\n want: %#v", got, want)
		}
	})
}

// TestMergeSrcMapReplacesScalar covers the "source map, non-map destination"
// switch (mergo.go:94-101): override disabled keeps the scalar, null_remove
// prunes nulls from the introduced subtree, and the default replaces the scalar.
func TestMergeSrcMapReplacesScalar(t *testing.T) {
	t.Run("override=false keeps scalar", func(t *testing.T) {
		objs := []map[string]any{
			{"x": "scalar"},
			{"x": map[string]any{"a": 1}},
		}

		opts := *NewDefaultOptions()
		opts.Override = false
		opts.NullOverride = false // install the transformer.

		got := mustMerge(t, objs, opts)
		want := map[string]any{"x": "scalar"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("scalar was replaced despite override=false\n got:  %#v\n want: %#v", got, want)
		}
	})

	t.Run("null_remove prunes nulls from introduced subtree", func(t *testing.T) {
		objs := []map[string]any{
			{"x": "scalar"},
			{"x": map[string]any{"a": 1, "b": nil}},
		}

		opts := *NewDefaultOptions()
		opts.NullRemove = true

		got := mustMerge(t, objs, opts)
		want := map[string]any{"x": map[string]any{"a": 1}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("null not pruned from replacing subtree\n got:  %#v\n want: %#v", got, want)
		}
	})

	t.Run("default replaces scalar with map", func(t *testing.T) {
		objs := []map[string]any{
			{"x": "scalar"},
			{"x": map[string]any{"a": 1}},
		}

		opts := *NewDefaultOptions()
		opts.UnionLists = true // install the transformer, keep override on.

		got := mustMerge(t, objs, opts)
		want := map[string]any{"x": map[string]any{"a": 1}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("scalar not replaced by map\n got:  %#v\n want: %#v", got, want)
		}
	})
}

// TestMergeAppendList covers both append paths: the transformer branch
// (mergo.go:105-106), reached when append_list combines with a transformer-
// installing flag, and the native mergo.WithAppendSlice config (mergo.go:27-29),
// used when append_list runs with otherwise-default options.
func TestMergeAppendList(t *testing.T) {
	objs := []map[string]any{
		{"l": []any{1, 2}},
		{"l": []any{3, 4}},
	}
	want := map[string]any{"l": []any{1, 2, 3, 4}}

	t.Run("transformer append branch", func(t *testing.T) {
		opts := *NewDefaultOptions()
		opts.AppendList = true
		opts.NullOverride = false // install the transformer.

		got := mustMerge(t, objs, opts)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("transformer append mismatch\n got:  %#v\n want: %#v", got, want)
		}
	})

	t.Run("native mergo append config", func(t *testing.T) {
		opts := *NewDefaultOptions()
		opts.AppendList = true // default options: no transformer, native WithAppendSlice.

		got := mustMerge(t, objs, opts)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("native append mismatch\n got:  %#v\n want: %#v", got, want)
		}
	})
}

// TestMergeOverrideDisabled covers the scalar override-disabled branch
// (mergo.go:109-110): with override off and the transformer installed, a shared
// scalar key keeps the earlier value while new keys are still added.
func TestMergeOverrideDisabled(t *testing.T) {
	objs := []map[string]any{
		{"a": "first", "b": "keep"},
		{"a": "second", "c": "add"},
	}

	opts := *NewDefaultOptions()
	opts.Override = false
	opts.NullOverride = false // install the transformer.

	got := mustMerge(t, objs, opts)
	want := map[string]any{"a": "first", "b": "keep", "c": "add"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("override=false did not preserve scalar\n got:  %#v\n want: %#v", got, want)
	}
}

// The merge() error path (mergo.go:49-52) is intentionally not covered: with the
// custom transformer installed, map merges never return an error, and native
// mergo does not error on the map[string]any inputs this package produces, so
// there is no deterministic input that reaches it.
