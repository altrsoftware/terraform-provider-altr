// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RepoUserDataSource{}

func NewRepoUserDataSource() datasource.DataSource {
	return &RepoUserDataSource{}
}

type RepoUserDataSource struct {
	client *client.Client
}

type RepoUserDataSourceModel struct {
	RepoName          types.String                      `tfsdk:"repo_name"`
	Username          types.String                      `tfsdk:"username"`
	AWSSecretsManager *AWSSecretsManagerDataSourceModel `tfsdk:"aws_secrets_manager"`
	AzureKeyVault     *AzureKeyVaultDataSourceModel     `tfsdk:"azure_key_vault"`
	CreatedAt         types.String                      `tfsdk:"created_at"`
	UpdatedAt         types.String                      `tfsdk:"updated_at"`
}

type AWSSecretsManagerDataSourceModel struct {
	IAMRole     types.String `tfsdk:"iam_role"`
	SecretsPath types.String `tfsdk:"secrets_path"`
}

type AzureKeyVaultDataSourceModel struct {
	KeyVaultURI types.String `tfsdk:"key_vault_uri"`
	SecretName  types.String `tfsdk:"secret_name"`
}

func (d *RepoUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_user"
}

func (d *RepoUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving repository user information with credential storage configuration.",
		Attributes: map[string]schema.Attribute{
			"repo_name": schema.StringAttribute{
				Description: "Name of the repository.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username of the repository user.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"aws_secrets_manager": schema.SingleNestedAttribute{
				Description: "AWS Secrets Manager configuration for storing credentials.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"iam_role": schema.StringAttribute{
						Description: "IAM role ARN for accessing the secret.",
						Computed:    true,
					},
					"secrets_path": schema.StringAttribute{
						Description: "Path to the secret in AWS Secrets Manager.",
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
						},
						Computed: true,
					},
				},
			},
			"azure_key_vault": schema.SingleNestedAttribute{
				Description: "Azure Key Vault configuration for storing credentials.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"key_vault_uri": schema.StringAttribute{
						Description: "URI of the Azure Key Vault.",
						Computed:    true,
					},
					"secret_name": schema.StringAttribute{
						Description: "Name of the secret in Azure Key Vault.",
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.UTF8LengthAtLeast(1),
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

func (d *RepoUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RepoUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RepoUserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get repo user from API
	repoUser, err := d.client.GetRepoUser(config.RepoName.ValueString(), config.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository user",
			fmt.Sprintf("Could not read repository user %s in repo %s: %s",
				config.Username.ValueString(),
				config.RepoName.ValueString(),
				err.Error()),
		)

		return
	}

	// If repo user doesn't exist, return error
	if repoUser == nil {
		resp.Diagnostics.AddError(
			"Repository user not found",
			fmt.Sprintf("Repository user '%s' in repo '%s' does not exist.",
				config.Username.ValueString(),
				config.RepoName.ValueString()),
		)

		return
	}

	// Map response to the model
	d.mapRepoUserToModel(repoUser, &config)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to map API response to Terraform model
func (d *RepoUserDataSource) mapRepoUserToModel(repoUser *client.RepoUser, model *RepoUserDataSourceModel) {
	model.RepoName = types.StringValue(repoUser.RepoName)
	model.Username = types.StringValue(repoUser.Username)
	model.CreatedAt = types.StringValue(repoUser.CreatedAt)
	model.UpdatedAt = types.StringValue(repoUser.UpdatedAt)

	// Handle AWS Secrets Manager
	if repoUser.AWSSecretsManager != nil && repoUser.AWSSecretsManager.SecretsPath != "" {
		model.AWSSecretsManager = &AWSSecretsManagerDataSourceModel{
			SecretsPath: types.StringValue(repoUser.AWSSecretsManager.SecretsPath),
		}
		if repoUser.AWSSecretsManager.IAMRole != "" {
			model.AWSSecretsManager.IAMRole = types.StringValue(repoUser.AWSSecretsManager.IAMRole)
		} else {
			model.AWSSecretsManager.IAMRole = types.StringNull()
		}
	} else {
		model.AWSSecretsManager = nil
	}

	// Handle Azure Key Vault
	if repoUser.AzureKeyVault != nil && repoUser.AzureKeyVault.SecretName != "" {
		model.AzureKeyVault = &AzureKeyVaultDataSourceModel{
			KeyVaultURI: types.StringValue(repoUser.AzureKeyVault.KeyVaultURI),
			SecretName:  types.StringValue(repoUser.AzureKeyVault.SecretName),
		}
	} else {
		model.AzureKeyVault = nil
	}
}
