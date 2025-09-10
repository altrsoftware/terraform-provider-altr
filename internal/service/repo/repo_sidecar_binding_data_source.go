// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"
	"fmt"
	"regexp"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/altrsoftware/terraform-provider-altr/internal/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RepoSidecarBindingDataSource{}

func NewRepoSidecarBindingDataSource() datasource.DataSource {
	return &RepoSidecarBindingDataSource{}
}

type RepoSidecarBindingDataSource struct {
	client *client.Client
}

type RepoSidecarBindingDataSourceModel struct {
	SidecarID types.String `tfsdk:"sidecar_id"`
	RepoName  types.String `tfsdk:"repo_name"`
	Port      types.Int64  `tfsdk:"port"`
}

func (d *RepoSidecarBindingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_sidecar_binding"
}

func (d *RepoSidecarBindingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving a specific repository sidecar binding.",
		Attributes: map[string]schema.Attribute{
			"sidecar_id": schema.StringAttribute{
				Description: "ID of the sidecar.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			"repo_name": schema.StringAttribute{
				Description: "Name of the repository.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Sidecar listener port bound to the repository.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		},
	}
}

func (d *RepoSidecarBindingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *RepoSidecarBindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RepoSidecarBindingDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get repo sidecar binding from API
	binding, err := d.client.GetRepoSidecarBinding(
		config.SidecarID.ValueString(),
		config.RepoName.ValueString(),
		int(config.Port.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repo sidecar binding",
			fmt.Sprintf("Could not read repo sidecar binding for sidecar %s, repo %s, port %d: %s",
				config.SidecarID.ValueString(),
				config.RepoName.ValueString(),
				config.Port.ValueInt64(),
				err.Error()),
		)

		return
	}

	// If binding doesn't exist, return error
	if binding == nil {
		resp.Diagnostics.AddError(
			"Repo sidecar binding not found",
			fmt.Sprintf("Repo sidecar binding for sidecar '%s', repo '%s', port %d does not exist.",
				config.SidecarID.ValueString(),
				config.RepoName.ValueString(),
				config.Port.ValueInt64()),
		)

		return
	}

	// Map response to the model
	d.mapBindingToModel(binding, &config)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to map API response to Terraform model
func (d *RepoSidecarBindingDataSource) mapBindingToModel(binding *client.RepoSidecarBinding, model *RepoSidecarBindingDataSourceModel) {
	model.SidecarID = types.StringValue(binding.SidecarID)
	model.RepoName = types.StringValue(binding.RepoName)
	model.Port = types.Int64Value(int64(binding.Port))
}
