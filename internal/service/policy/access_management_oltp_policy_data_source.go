// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AccessManagementOLTPPolicyDataSource{}

func NewAccessManagementOLTPPolicyDataSource() datasource.DataSource {
	return &AccessManagementOLTPPolicyDataSource{}
}

type AccessManagementOLTPPolicyDataSource struct {
	client *client.Client
}

type AccessManagementOLTPPolicyDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Rules            types.List   `tfsdk:"rules"`
	RepoName         types.String `tfsdk:"repo_name"`
	CaseSensitivity  types.String `tfsdk:"case_sensitivity"`
	DatabaseType     types.Int64  `tfsdk:"database_type"`
	DatabaseTypeName types.String `tfsdk:"database_type_name"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (d *AccessManagementOLTPPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_management_oltp_policy"
}

func (d *AccessManagementOLTPPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving an OLTP access management policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the OLTP access management policy.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the OLTP access management policy.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the OLTP access management policy.",
				Computed:    true,
			},
			"repo_name": schema.StringAttribute{
				Description: "The name of the repository this policy belongs to.",
				Computed:    true,
			},
			"case_sensitivity": schema.StringAttribute{
				Description: "Case sensitivity for the policy.",
				Computed:    true,
			},
			"database_type": schema.Int64Attribute{
				Description: "Database type ID for the policy.",
				Computed:    true,
			},
			"database_type_name": schema.StringAttribute{
				Description: "Database type name for the policy.",
				Computed:    true,
			},
			"rules": schema.ListNestedAttribute{
				Description: "List of rules for the OLTP access management policy.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Type of the rule.",
							Computed:    true,
						},
						"actors": schema.ListNestedAttribute{
							Description: "List of actors for the rule.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the actor.",
										Computed:    true,
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the actor.",
										Computed:    true,
									},
									"identifiers": schema.ListAttribute{
										Description: "List of identifiers for the actor.",
										ElementType: types.StringType,
										Computed:    true,
									},
								},
							},
						},
						"objects": schema.ListNestedAttribute{
							Description: "List of objects for the rule.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the object.",
										Computed:    true,
									},
									"identifiers": schema.ListNestedAttribute{
										Description: "List of identifiers for the object.",
										Computed:    true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"database": schema.SingleNestedAttribute{
													Description: "Database identifier part.",
													Computed:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the database.",
															Computed:    true,
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the database.",
															Computed:    true,
														},
													},
												},
												"schema": schema.SingleNestedAttribute{
													Description: "Schema identifier part.",
													Computed:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the schema.",
															Computed:    true,
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the schema.",
															Computed:    true,
														},
													},
												},
												"table": schema.SingleNestedAttribute{
													Description: "Table identifier part.",
													Computed:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the table.",
															Computed:    true,
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the table.",
															Computed:    true,
														},
													},
												},
												"column": schema.SingleNestedAttribute{
													Description: "Column identifier part.",
													Computed:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the column.",
															Optional:    true,
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the column.",
															Optional:    true,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
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

func (d *AccessManagementOLTPPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccessManagementOLTPPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccessManagementOLTPPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the OLTP policy from the API
	policy, err := d.client.GetAccessManagementOLTPPolicy(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OLTP access management policy",
			fmt.Sprintf("Could not read OLTP access management policy with ID %s: %s", config.ID.ValueString(), err.Error()),
		)
		return
	}

	// If the policy doesn't exist, return an error
	if policy == nil {
		resp.Diagnostics.AddError(
			"OLTP access management policy not found",
			fmt.Sprintf("OLTP access management policy with ID '%s' does not exist.", config.ID.ValueString()),
		)
		return
	}

	// Map response to the model
	d.mapPolicyToModel(policy, &config)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to map API response to Terraform model
func (d *AccessManagementOLTPPolicyDataSource) mapPolicyToModel(policy *client.AccessManagementOLTPPolicy, model *AccessManagementOLTPPolicyDataSourceModel) {
	model.ID = types.StringValue(policy.ID)
	model.Name = types.StringValue(policy.Name)
	model.Description = types.StringValue(policy.Description)
	model.DatabaseTypeName = types.StringValue(policy.DatabaseTypeName)
	model.DatabaseType = types.Int64Value(policy.DatabaseType)
	model.CaseSensitivity = types.StringValue(policy.CaseSensitivity)
	model.RepoName = types.StringValue(policy.RepoName)
	model.Rules = convertAccessManagementOLTPRulesToTerraform(policy.Rules)
	model.CreatedAt = types.StringValue(policy.CreatedAt)
	model.UpdatedAt = types.StringValue(policy.UpdatedAt)
}
