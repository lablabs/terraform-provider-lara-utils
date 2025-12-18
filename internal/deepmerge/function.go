// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package deepmerge

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DeepMergeFunction interface {
	FunctionSummary() string
	FunctionDescription() string
	FunctionObjectsParameter() function.Parameter
	FunctionResult(context.Context, map[string]any) (basetypes.DynamicValue, diag.Diagnostics)

	GetMergingObjects(context.Context, function.ArgumentsData) ([]map[string]any, *function.FuncError)
	GetMergingOptions(context.Context, function.ArgumentsData) (*DeepMergeOptions, *function.FuncError)
}

type DeepMergeOptions struct {
	Override     bool `mapstructure:"override"`
	NullOverride bool `mapstructure:"null_override"`
	AppendList   bool `mapstructure:"append_list"`
	DeepCopyList bool `mapstructure:"deep_copy_list"`
	UnionLists   bool `mapstructure:"union_lists"`
}

func NewFunctionDefinition(fn DeepMergeFunction) function.Definition {
	return function.Definition{
		Summary:             fn.FunctionSummary(),
		MarkdownDescription: fn.FunctionDescription(),
		Parameters: []function.Parameter{
			fn.FunctionObjectsParameter(),
		},
		VariadicParameter: function.DynamicParameter{
			Name:                "options",
			MarkdownDescription: "Merging options",
			AllowNullValue:      false,
			AllowUnknownValues:  false,
		},
		Return: function.DynamicReturn{},
	}
}

func NewDefaultOptions() *DeepMergeOptions {
	return &DeepMergeOptions{
		Override:     true,
		NullOverride: true,
		AppendList:   false,
		DeepCopyList: false,
		UnionLists:   false,
	}
}

func Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse, fn DeepMergeFunction) {
	objs, err := fn.GetMergingObjects(ctx, req.Arguments)
	if resp.Error = function.ConcatFuncErrors(err); resp.Error != nil {
		return
	}

	opts, err := fn.GetMergingOptions(ctx, req.Arguments)
	if resp.Error = function.ConcatFuncErrors(err); resp.Error != nil {
		return
	}

	merged, diags := merge(objs, *opts)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	result, diags := fn.FunctionResult(ctx, merged)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, &result))
}
