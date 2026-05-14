// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

// Originally copied from https://github.com/isometry/terraform-provider-deepmerge/blob/main/internal/provider/mergo_function.go

package deepmerge

import (
	"fmt"
	"reflect"
	"slices"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type DeepMergeTransformer struct {
	DeepMergeOptions
	// pathAware is true when path_overrides drove the config. In that mode the
	// transformer is the sole authority for map/list merge semantics — mergo's
	// global flags (WithOverride, WithAppendSlice) are not registered, so we
	// must reproduce their behavior ourselves.
	pathAware bool
	rules     []pathRule
	curPath   []string
}

func (opts DeepMergeOptions) newMergoConfig() ([]func(*mergo.Config), error) {
	cfg := []func(*mergo.Config){}

	hasPathOverrides := len(opts.PathOverrides) > 0

	// When path_overrides is set, the transformer fully owns map and list
	// merging so it can apply per-path effective options. Mergo's global
	// list/override flags would pre-empt the transformer, so we don't add them.
	if !hasPathOverrides {
		if opts.Override {
			cfg = append(cfg, mergo.WithOverride)
		}
		if opts.AppendList {
			cfg = append(cfg, mergo.WithAppendSlice)
		}
	}

	// deep_copy_list stays global-only in v1 and is always honored via mergo.
	if opts.DeepCopyList {
		cfg = append(cfg, mergo.WithSliceDeepCopy)
	}

	if hasPathOverrides || !opts.NullOverride || opts.UnionLists {
		rules, err := compilePathRules(opts.PathOverrides)
		if err != nil {
			return nil, err
		}
		cfg = append(cfg, mergo.WithTransformers(DeepMergeTransformer{
			DeepMergeOptions: opts,
			pathAware:        hasPathOverrides,
			rules:            rules,
		}))
	}

	return cfg, nil
}

func merge(objs []map[string]any, opts DeepMergeOptions) (merged map[string]any, diags diag.Diagnostics) {
	cfg, err := opts.newMergoConfig()
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("invalid path_overrides", err.Error()))
		return
	}

	dst := make(map[string]any)
	for i, m := range objs {
		if err := mergo.Merge(&dst, m, cfg...); err != nil {
			diags.Append(diag.NewErrorDiagnostic(fmt.Sprintf("error merging argument %d", i+1), err.Error()))
			return
		}
	}

	return dst, nil
}

func (t DeepMergeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Map {
		return func(dst, src reflect.Value) error {
			t.mergeMaps(dst, src)
			return nil
		}
	}
	return nil
}

// effectiveOptionsAt returns the global options overlaid by any matching path rule.
// When multiple rules match the same path, the last one in sorted order wins
// (compilePathRules sorts the rules for deterministic precedence).
func (t DeepMergeTransformer) effectiveOptionsAt(path []string) DeepMergeOptions {
	eff := t.DeepMergeOptions
	for _, r := range t.rules {
		if r.matches(path) {
			eff = r.opts.Apply(eff)
		}
	}
	return eff
}

func (t DeepMergeTransformer) mergeMaps(dst, src reflect.Value) reflect.Value {
	for _, key := range src.MapKeys() {
		srcElem := src.MapIndex(key)
		dstElem := dst.MapIndex(key)

		// Unwrap the interfaces of srcElem and dstElem
		if srcElem.Kind() == reflect.Interface {
			srcElem = srcElem.Elem()
		}

		if dstElem.Kind() == reflect.Interface {
			dstElem = dstElem.Elem()
		}

		// effectiveOptionsAt re-runs the path-rule overlay at each step so
		// per-path overrides apply only at the matching level, not below.
		childPath := slices.Concat(t.curPath, []string{fmt.Sprintf("%v", key.Interface())})
		eff := t.effectiveOptionsAt(childPath)

		if srcElem.Kind() == reflect.Map && dstElem.Kind() == reflect.Map {
			child := t
			child.curPath = childPath
			dst.SetMapIndex(key, child.mergeMaps(dstElem, srcElem))
		} else if !srcElem.IsValid() && !eff.NullOverride { // skip override of nil values only if nullOverride is false
			continue
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && eff.UnionLists {
			dst.SetMapIndex(key, unionSlices(dstElem, srcElem))
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && eff.AppendList {
			dst.SetMapIndex(key, reflect.AppendSlice(dstElem, srcElem))
		} else if t.pathAware && !eff.Override && dstElem.IsValid() {
			// In path-aware mode the transformer owns override semantics
			// (mergo.WithOverride isn't registered). Outside path-aware mode
			// mergo handles override and this branch is skipped.
			continue
		} else {
			dst.SetMapIndex(key, srcElem)
		}
	}

	return dst
}

func unionSlices(dst, src reflect.Value) reflect.Value {
	result := reflect.MakeSlice(dst.Type(), 0, dst.Len()+src.Len())

	// Add elements from dst (preserving order)
	for i := 0; i < dst.Len(); i++ {
		if !containsElement(result, dst.Index(i)) {
			result = reflect.Append(result, dst.Index(i))
		}
	}

	// Add new elements from src
	for i := 0; i < src.Len(); i++ {
		if !containsElement(result, src.Index(i)) {
			result = reflect.Append(result, src.Index(i))
		}
	}

	return result
}

// containsElement checks if a slice contains a specific element using reflect.DeepEqual.
func containsElement(slice, elem reflect.Value) bool {
	elemInterface := elem.Interface()
	for i := 0; i < slice.Len(); i++ {
		if reflect.DeepEqual(slice.Index(i).Interface(), elemInterface) {
			return true
		}
	}
	return false
}
