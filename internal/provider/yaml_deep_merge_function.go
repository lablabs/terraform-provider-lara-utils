// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/lablabs/terraform-provider-lara-utils/internal/deepmerge"
	"github.com/lablabs/terraform-provider-lara-utils/internal/helpers"
	"sigs.k8s.io/yaml"
)

var (
	_ function.Function = YamlDeepMergeFunction{}
	//go:embed yaml_deep_merge_function.md
	yamlDeepMergeFunctionDescription string
)

type YamlDeepMergeFunction struct {
	DeepMergeFunction
}

func NewYamlDeepMergeFunction() function.Function {
	return YamlDeepMergeFunction{}
}

func (fn YamlDeepMergeFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "yaml_deep_merge"
}

func (fn YamlDeepMergeFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = deepmerge.NewFunctionDefinition(fn)
}

func (fn YamlDeepMergeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	deepmerge.Run(ctx, req, resp, fn)
}

func (fn YamlDeepMergeFunction) FunctionSummary() string {
	return "Deep merge YAML-encoded objects"
}

func (fn YamlDeepMergeFunction) FunctionDescription() string {
	return yamlDeepMergeFunctionDescription
}

func (fn YamlDeepMergeFunction) FunctionObjectsParameter() function.Parameter {
	return function.ListParameter{
		Name:                "objects",
		MarkdownDescription: "List of YAML strings to merge",
		ElementType:         basetypes.StringType{},
		AllowNullValue:      false,
		AllowUnknownValues:  false,
	}
}

func (r YamlDeepMergeFunction) GetMergingObjects(ctx context.Context, args function.ArgumentsData) ([]map[string]any, *function.FuncError) {
	arg := basetypes.ListValue{}
	if err := args.GetArgument(ctx, 0, &arg); err != nil {
		return nil, err
	}

	objs := []map[string]any{}
	for idx, elem := range arg.Elements() {
		val, err := helpers.EncodeValue(elem)
		if err != nil {
			return nil, function.NewArgumentFuncError(int64(0), err.Error())
		}
		if _, ok := val.(string); !ok {
			return nil, function.NewArgumentFuncError(int64(0), fmt.Sprintf("merging argument %d must be string, got: %s", idx+1, reflect.TypeOf(val)))
		}

		var obj map[string]any
		if err := yaml.Unmarshal([]byte(val.(string)), &obj); err != nil { //nolint:forcetypeassert
			return nil, function.NewArgumentFuncError(int64(0), strings.ReplaceAll(err.Error(), "JSON", "YAML")) // sigs.k8s.io/yaml.Unmarshal returns JSON-related error messages
		}

		objs = append(objs, obj)
	}

	return objs, nil
}

func (fn YamlDeepMergeFunction) FunctionResult(ctx context.Context, merged map[string]any) (basetypes.DynamicValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if len(merged) == 0 {
		return types.DynamicValue(types.StringValue("")), diags
	}

	value, err := yaml.Marshal(merged)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("error marshaling merged result to YAML", err.Error()))
	}

	return types.DynamicValue(types.StringValue(string(value))), diags
}
