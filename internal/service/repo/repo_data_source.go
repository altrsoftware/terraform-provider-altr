// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"
	"fmt"
	"regexp"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/altrsoftware/terraform-provider-altr/internal/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RepoDataSource{}

func NewRepoDataSource() datasource.DataSource {
	return &RepoDataSource{}
}

type RepoDataSource struct {
	client *client.Client
}

type RepoDataSourceModel struct {
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	Hostname     types.String `tfsdk:"hostname"`
	Port         types.Int64  `tfsdk:"port"`
	UserCount    types.Int64  `tfsdk:"user_count"`
	BindingCount types.Int64  `tfsdk:"binding_count"`
	OrgID        types.String `tfsdk:"org_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (d *RepoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo"
}

func (d *RepoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving repository information.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the repository.",
				Required:    true,
				Validators: []validator.String{
					// Validate string value must be "one", "two", or "three"
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.AlphanumericAndUnderscoreRegex),
						"must be alphanumeric or underscore",
					),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the repository.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the repository (e.g., Oracle, etc.).",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the repository.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port number of the repository.",
				Computed:    true,
			},
			"user_count": schema.Int64Attribute{
				Description: "Number of users associated with this repository.",
				Computed:    true,
			},
			"binding_count": schema.Int64Attribute{
				Description: "Number of sidecar bindings for this repository.",
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID that owns this repository.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (d *RepoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RepoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RepoDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get repo from API
	repo, err := d.client.GetRepo(config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository",
			"Could not read repository "+config.Name.ValueString()+": "+err.Error(),
		)

		return
	}

	// If repo doesn't exist, return error
	if repo == nil {
		resp.Diagnostics.AddError(
			"Repository not found",
			"Repository with name '"+config.Name.ValueString()+"' does not exist.",
		)

		return
	}

	// Map response to the model
	d.mapRepoToModel(repo, &config)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to map API response to Terraform model
func (d *RepoDataSource) mapRepoToModel(repo *client.Repo, model *RepoDataSourceModel) {
	model.Name = types.StringValue(repo.Name)
	model.Description = types.StringValue(repo.Description)
	model.Type = types.StringValue(repo.Type)
	model.Hostname = types.StringValue(repo.Hostname)
	model.Port = types.Int64Value(int64(repo.Port))
	model.UserCount = types.Int64Value(int64(repo.UserCount))
	model.BindingCount = types.Int64Value(int64(repo.BindingCount))
	model.OrgID = types.StringValue(repo.OrgID)
	model.CreatedAt = types.StringValue(repo.CreatedAt)
	model.UpdatedAt = types.StringValue(repo.UpdatedAt)
}
