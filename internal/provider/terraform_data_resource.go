// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource = &TerraformDataResource{}
	//go:embed terraform_data_resource.md
	terraformDataResourceDescription string
)

// TerraformDataResource stores any HCL value in Terraform state so that
// Terraform's own plan engine diffs it natively — showing individual field
// changes like `~ cpu = "20m" -> "30m"` instead of a raw text diff.
// Equivalent to terraform_data but without the redundant output attribute.
type TerraformDataResource struct{}

func NewTerraformDataResource() resource.Resource {
	return &TerraformDataResource{}
}

type TerraformDataResourceModel struct {
	ID    types.String  `tfsdk:"id"`
	Input types.Dynamic `tfsdk:"input"`
}

func (r *TerraformDataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_data"
}

func (r *TerraformDataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: terraformDataResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier generated at creation time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"input": schema.DynamicAttribute{
				Required:            true,
				MarkdownDescription: "The value to track. Accepts any Terraform type — string, number, bool, map, list, object. Terraform's plan engine diffs this attribute natively.",
			},
		},
	}
}

func (r *TerraformDataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TerraformDataResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		resp.Diagnostics.AddError("error generating resource id", fmt.Sprintf("unexpected error generating UUID: %s", err))
		return
	}
	data.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerraformDataResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// state is the source of truth; nothing to refresh from a remote
}

func (r *TerraformDataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TerraformDataResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerraformDataResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
