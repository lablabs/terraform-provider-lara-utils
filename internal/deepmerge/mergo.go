// Copyright (c) Labyrinth Labs s.r.o.
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

	if !opts.NullOverride || opts.UnionLists {
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

		// Unwrap the interfaces of srcElem and dstElem
		if srcElem.Kind() == reflect.Interface {
			srcElem = srcElem.Elem()
		}

		if dstElem.Kind() == reflect.Interface {
			dstElem = dstElem.Elem()
		}

		if srcElem.Kind() == reflect.Map && dstElem.Kind() == reflect.Map {
			newValue := t.mergeMaps(dstElem, srcElem) // recursive call
			dst.SetMapIndex(key, newValue)
		} else if !srcElem.IsValid() && !t.NullOverride { // skip override of nil values only if nullOverride is false
			continue
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && t.UnionLists { // handle union
			dst.SetMapIndex(key, unionSlices(dstElem, srcElem))
		} else if srcElem.Kind() == reflect.Slice && dstElem.Kind() == reflect.Slice && t.AppendList { // handle append
			dst.SetMapIndex(key, reflect.AppendSlice(dstElem, srcElem))
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
