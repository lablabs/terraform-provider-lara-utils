// Copyright (c) Labyrinth Labs s.r.o.
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
