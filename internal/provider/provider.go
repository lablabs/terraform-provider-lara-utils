// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure LaraUtilsProvider satisfies various provider interfaces.
var _ provider.Provider = &LaraUtilsProvider{}
var _ provider.ProviderWithFunctions = &LaraUtilsProvider{}

// LaraUtilsProvider defines the provider implementation.
type LaraUtilsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// LaraUtilsProviderModel describes the provider data model.
type LaraUtilsProviderModel struct {
}

func (p *LaraUtilsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lara-utils"
	resp.Version = p.version
}

func (p *LaraUtilsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema.Description = `LARA utilities, e.g. map deep merge function.`
}

func (p *LaraUtilsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *LaraUtilsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *LaraUtilsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *LaraUtilsProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewDeepMergeFunction,
		NewYamlDeepMergeFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LaraUtilsProvider{
			version: version,
		}
	}
}
