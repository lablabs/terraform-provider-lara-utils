// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &NullValuesResource{}
	//go:embed null_values_resource.md
	nullValuesResourceDescription string
)

// NullValuesResource stores a structured (HCL-decoded) values object in
// Terraform state. Because the attribute is Dynamic and not sensitive,
// Terraform's own plan engine diffs the structure natively — showing
// individual field changes like `~ cpu = "20m" -> "30m"` instead of a
// raw text diff.
//
// Usage:
//
//	resource "lara-utils_null_values" "values" {
//	  values = yamldecode(nonsensitive(var.values))
//	}
type NullValuesResource struct{}

func NewNullValuesResource() resource.Resource {
	return &NullValuesResource{}
}

type NullValuesResourceModel struct {
	ID     types.String  `tfsdk:"id"`
	Values types.Dynamic `tfsdk:"values"`
}

func (r *NullValuesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_null_values"
}

func (r *NullValuesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: nullValuesResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Static identifier, always `null_values`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"values": schema.DynamicAttribute{
				Required: true,
				MarkdownDescription: "Structured values to track across Terraform applies. " +
					"Pass `yamldecode(nonsensitive(var.values))` to get field-level change visibility in plan output. " +
					"Terraform's plan engine diffs this attribute natively — no text diff needed.",
			},
		},
	}
}

func (r *NullValuesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NullValuesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue("null_values")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NullValuesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NullValuesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NullValuesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NullValuesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NullValuesResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
