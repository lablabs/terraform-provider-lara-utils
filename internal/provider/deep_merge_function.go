// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/lablabs/terraform-provider-lara-utils/internal/deepmerge"
	"github.com/lablabs/terraform-provider-lara-utils/internal/helpers"
	"github.com/mitchellh/mapstructure"
)

var (
	_ function.Function = DeepMergeFunction{}
	//go:embed deep_merge_function.md
	deepMergeFunctionDescription string
)

type DeepMergeFunction struct{}

func NewDeepMergeFunction() function.Function {
	return DeepMergeFunction{}
}

func (fn DeepMergeFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "deep_merge"
}

func (fn DeepMergeFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = deepmerge.NewFunctionDefinition(fn)
}

func (fn DeepMergeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	deepmerge.Run(ctx, req, resp, fn)
}

func (fn DeepMergeFunction) FunctionSummary() string {
	return "Deep merge objects"
}

func (fn DeepMergeFunction) FunctionDescription() string {
	return deepMergeFunctionDescription
}

func (fn DeepMergeFunction) FunctionObjectsParameter() function.Parameter {
	return function.DynamicParameter{
		Name:                "objects",
		MarkdownDescription: "List of objects to merge",
		AllowNullValue:      false,
		AllowUnknownValues:  false,
	}
}

func (fn DeepMergeFunction) GetMergingObjects(ctx context.Context, args function.ArgumentsData) ([]map[string]any, *function.FuncError) {
	arg := types.Dynamic{}
	if err := args.GetArgument(ctx, 0, &arg); err != nil {
		return nil, err
	}

	var elems []attr.Value
	argVal := arg.UnderlyingValue()
	switch argType := argVal.Type(ctx).(type) {
	case basetypes.SetType:
		elems = argVal.(basetypes.SetValue).Elements() //nolint:forcetypeassert
	case basetypes.ListType:
		elems = argVal.(basetypes.ListValue).Elements() //nolint:forcetypeassert
	case basetypes.TupleType:
		elems = argVal.(basetypes.TupleValue).Elements() //nolint:forcetypeassert
	default:
		return nil, function.NewArgumentFuncError(int64(0), fmt.Sprintf("list of objects required, got: %s", argType))
	}

	objs := []map[string]any{}
	for idx, elem := range elems {
		val, err := helpers.EncodeValue(elem)
		if err != nil {
			return nil, function.NewArgumentFuncError(int64(0), err.Error())
		}
		if _, ok := val.(map[string]any); !ok {
			return nil, function.NewArgumentFuncError(int64(0), fmt.Sprintf("merging argument %d must be object, got: %s", idx+1, reflect.TypeOf(val)))
		}

		obj := val.(map[string]any) //nolint:forcetypeassert
		if len(obj) == 0 {
			continue
		}

		objs = append(objs, obj)
	}

	return objs, nil
}

func (fn DeepMergeFunction) GetMergingOptions(ctx context.Context, args function.ArgumentsData) (*deepmerge.DeepMergeOptions, *function.FuncError) {
	arg := basetypes.TupleValue{}
	if err := args.GetArgument(ctx, 1, &arg); err != nil {
		return nil, err
	}

	opts := deepmerge.NewDefaultOptions()
	if arg.IsNull() {
		return opts, nil
	}

	for idx, elem := range arg.Elements() {
		val, err := helpers.EncodeValue(elem)
		if err != nil {
			return nil, function.NewArgumentFuncError(int64(idx), err.Error())
		}

		if err := mapstructure.Decode(val, &opts); err != nil {
			return nil, function.NewArgumentFuncError(int64(idx), err.Error())
		}
	}

	return opts, nil
}

func (fn DeepMergeFunction) FunctionResult(ctx context.Context, merged map[string]any) (basetypes.DynamicValue, diag.Diagnostics) {
	value, diags := helpers.DecodeScalar(ctx, merged)
	return types.DynamicValue(value), diags
}
