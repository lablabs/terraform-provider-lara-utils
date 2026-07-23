// Copyright Labyrinth Labs s.r.o. 2025, 2026
// SPDX-License-Identifier: MPL-2.0

// Originally copied from https://github.com/isometry/terraform-provider-deepmerge/blob/main/internal/provider/mergo_function.go

package deepmerge

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type DeepMergeTransformer struct {
	DeepMergeOptions
}

func (opts DeepMergeOptions) newMergoConfig() []func(*mergo.Config) {
	cfg := []func(*mergo.Config){}

	if opts.Override {
		cfg = append(cfg, mergo.WithOverride)
	}

	if opts.AppendList {
		cfg = append(cfg, mergo.WithAppendSlice)
	}

	if opts.DeepCopyList {
		cfg = append(cfg, mergo.WithSliceDeepCopy)
	}

	if !opts.NullOverride || opts.NullRemove || opts.UnionLists {
		cfg = append(cfg, mergo.WithTransformers(DeepMergeTransformer{
			DeepMergeOptions: opts,
		}))
	}

	return cfg
}

func merge(objs []map[string]any, opts DeepMergeOptions) (merged map[string]any, diags diag.Diagnostics) {
	cfg := opts.newMergoConfig()

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

func (t DeepMergeTransformer) mergeMaps(dst, src reflect.Value) reflect.Value {
	for _, key := range src.MapKeys() {
		srcElem := src.MapIndex(key)
		dstElem := dst.MapIndex(key)

		// Unwrap the interfaces of srcElem and dstElem.
		if srcElem.Kind() == reflect.Interface {
			srcElem = srcElem.Elem()
		}

		if dstElem.Kind() == reflect.Interface {
			dstElem = dstElem.Elem()
		}

		switch {
		case !srcElem.IsValid(): // source value is null.
			switch {
			case t.NullRemove: // remove the key entirely; takes precedence over nullOverride.
				dst.SetMapIndex(key, reflect.Value{})
			case !t.NullOverride: // keep the existing value; null does not override.
				continue
			default: // nullOverride: drop the key (matches prior transformer behavior).
				dst.SetMapIndex(key, reflect.Value{})
			}
		case srcElem.Kind() == reflect.Map && dstElem.Kind() == reflect.Map: // deep-merge nested maps (recursion ignores override).
			dst.SetMapIndex(key, t.mergeMaps(dstElem, srcElem))
		case srcElem.Kind() == reflect.Map: // source map replacing a non-map or absent destination.
			switch {
			case dstElem.IsValid() && !t.Override: // keep the existing value when override is disabled.
				continue
			case t.NullRemove: // strip nested nulls from the newly-introduced subtree.
				dst.SetMapIndex(key, t.mergeMaps(reflect.MakeMap(srcElem.Type()), srcElem))
			default:
				dst.SetMapIndex(key, srcElem)
			}
		case srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && t.UnionLists: // merge lists as sets.
			dst.SetMapIndex(key, unionSlices(dstElem, srcElem))
		case srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && t.AppendList: // concatenate lists.
			dst.SetMapIndex(key, reflect.AppendSlice(dstElem, srcElem))
		case srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && t.DeepCopyList: // deep-merge lists element by element.
			dst.SetMapIndex(key, t.deepCopySlices(dstElem, srcElem))
		case dstElem.IsValid() && !t.Override: // keep the existing value when override is disabled.
			continue
		default:
			dst.SetMapIndex(key, srcElem)
		}
	}

	return dst
}

// deepCopySlices merges two slices element by element, mirroring mergo's
// WithSliceDeepCopy: the result keeps the destination's length, overlapping map
// elements are deep-merged, and every other element keeps the destination value.
// Non-map elements are never overridden and elements in src beyond the
// destination's length are dropped, matching mergo.
func (t DeepMergeTransformer) deepCopySlices(dst, src reflect.Value) reflect.Value {
	result := reflect.MakeSlice(dst.Type(), dst.Len(), dst.Len())
	for i := 0; i < dst.Len(); i++ {
		dstItem := dst.Index(i)
		result.Index(i).Set(dstItem) // default: keep the destination element.

		if i >= src.Len() {
			continue
		}

		dstElem := dstItem
		srcElem := src.Index(i)
		if dstElem.Kind() == reflect.Interface {
			dstElem = dstElem.Elem()
		}
		if srcElem.Kind() == reflect.Interface {
			srcElem = srcElem.Elem()
		}

		// Only overlapping map elements are deep-merged; everything else keeps dst.
		if dstElem.Kind() == reflect.Map && srcElem.Kind() == reflect.Map {
			result.Index(i).Set(t.mergeMaps(dstElem, srcElem))
		}
	}

	return result
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
