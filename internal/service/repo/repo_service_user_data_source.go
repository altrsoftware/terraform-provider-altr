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

var _ datasource.DataSource = &ServiceUserDataSource{}

func NewServiceUserDataSource() datasource.DataSource {
	return &ServiceUserDataSource{}
}

type ServiceUserDataSource struct {
	client *client.Client
}

type ServiceUserDataSourceModel struct {
	RepoName            types.String                        `tfsdk:"repo_name"`
	Username            types.String                        `tfsdk:"username"`
	Resource            types.String                        `tfsdk:"resource"`
	AWSSecretsManager   *AWSSecretsManagerDataSourceModel   `tfsdk:"aws_secrets_manager"`
	AzureKeyVault       *AzureKeyVaultDataSourceModel       `tfsdk:"azure_key_vault"`
	EnvironmentVariable *EnvironmentVariableDataSourceModel `tfsdk:"environment_variable"`
	SecretFile          *SecretFileDataSourceModel          `tfsdk:"secret_file"`
	TaskCount           types.Int64                         `tfsdk:"task_count"`
	CreatedAt           types.String                        `tfsdk:"created_at"`
	UpdatedAt           types.String                        `tfsdk:"updated_at"`
}

type EnvironmentVariableDataSourceModel struct {
	VariableName types.String `tfsdk:"variable_name"`
}

type SecretFileDataSourceModel struct {
	Path types.String `tfsdk:"path"`
}

func (d *ServiceUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (d *ServiceUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving a repository service user used for agent task authentication.",
		Attributes: map[string]schema.Attribute{
			"repo_name": schema.StringAttribute{
				Description: "Name of the repository this service user belongs to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"username": schema.StringAttribute{
				Description: "Database username.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
			"resource": schema.StringAttribute{
				Description: "Database entrypoint the agent connects to with this service user (e.g. an Oracle service name like \"ORCL\").",
				Computed:    true,
			},
			"aws_secrets_manager": schema.SingleNestedAttribute{
				Description: "AWS Secrets Manager credential provider.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"iam_role": schema.StringAttribute{
						Description: "ARN of an IAM role to assume when retrieving the secret.",
						Computed:    true,
					},
					"secrets_path": schema.StringAttribute{
						Description: "Path or name of the secret in AWS Secrets Manager.",
						Computed:    true,
					},
				},
			},
			"azure_key_vault": schema.SingleNestedAttribute{
				Description: "Azure Key Vault credential provider.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"key_vault_uri": schema.StringAttribute{
						Description: "HTTPS URL of the Azure Key Vault.",
						Computed:    true,
					},
					"secret_name": schema.StringAttribute{
						Description: "Name of the secret within the vault.",
						Computed:    true,
					},
				},
			},
			"environment_variable": schema.SingleNestedAttribute{
				Description: "Environment variable credential provider.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"variable_name": schema.StringAttribute{
						Description: "Name of the OS environment variable containing the secret.",
						Computed:    true,
					},
				},
			},
			"secret_file": schema.SingleNestedAttribute{
				Description: "Secret file credential provider. Reads from /altr/secrets/<path> at runtime.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Description: "Simple filename resolved under /altr/secrets/ at runtime.",
						Computed:    true,
					},
				},
			},
			"task_count": schema.Int64Attribute{
				Description: "Number of agent tasks currently using this service user.",
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

func (d *ServiceUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = c
}

func (d *ServiceUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ServiceUserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	su, err := d.client.GetServiceUser(config.RepoName.ValueString(), config.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service user",
			fmt.Sprintf("Could not read service user %s in repo %s: %s",
				config.Username.ValueString(),
				config.RepoName.ValueString(),
				err.Error()),
		)

		return
	}

	if su == nil {
		resp.Diagnostics.AddError(
			"Service user not found",
			fmt.Sprintf("Service user '%s' in repo '%s' does not exist.",
				config.Username.ValueString(),
				config.RepoName.ValueString()),
		)

		return
	}

	d.mapServiceUserToModel(su, &config)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func (d *ServiceUserDataSource) mapServiceUserToModel(su *client.ServiceUser, model *ServiceUserDataSourceModel) {
	model.RepoName = types.StringValue(su.RepoName)
	model.Username = types.StringValue(su.Username)
	model.Resource = types.StringValue(su.Resource)
	model.TaskCount = types.Int64Value(int64(su.TaskCount))
	model.CreatedAt = types.StringValue(su.CreatedAt)
	model.UpdatedAt = types.StringValue(su.UpdatedAt)

	if su.AWSSecretsManager != nil && su.AWSSecretsManager.SecretsPath != "" {
		model.AWSSecretsManager = &AWSSecretsManagerDataSourceModel{
			SecretsPath: types.StringValue(su.AWSSecretsManager.SecretsPath),
		}
		if su.AWSSecretsManager.IAMRole != "" {
			model.AWSSecretsManager.IAMRole = types.StringValue(su.AWSSecretsManager.IAMRole)
		} else {
			model.AWSSecretsManager.IAMRole = types.StringNull()
		}
	} else {
		model.AWSSecretsManager = nil
	}

	if su.AzureKeyVault != nil && su.AzureKeyVault.SecretName != "" {
		model.AzureKeyVault = &AzureKeyVaultDataSourceModel{
			KeyVaultURI: types.StringValue(su.AzureKeyVault.KeyVaultURI),
			SecretName:  types.StringValue(su.AzureKeyVault.SecretName),
		}
	} else {
		model.AzureKeyVault = nil
	}

	if su.EnvironmentVariable != nil && su.EnvironmentVariable.VariableName != "" {
		model.EnvironmentVariable = &EnvironmentVariableDataSourceModel{
			VariableName: types.StringValue(su.EnvironmentVariable.VariableName),
		}
	} else {
		model.EnvironmentVariable = nil
	}

	if su.SecretFile != nil && su.SecretFile.Path != "" {
		model.SecretFile = &SecretFileDataSourceModel{
			Path: types.StringValue(su.SecretFile.Path),
		}
	} else {
		model.SecretFile = nil
	}
}
