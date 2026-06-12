// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &ServiceUserResource{}
	_ resource.ResourceWithImportState = &ServiceUserResource{}
)

func NewServiceUserResource() resource.Resource {
	return &ServiceUserResource{}
}

type ServiceUserResource struct {
	client *client.Client
}

type ServiceUserResourceModel struct {
	ID                  types.String          `tfsdk:"id"`
	RepoName            types.String          `tfsdk:"repo_name"`
	Username            types.String          `tfsdk:"username"`
	AWSSecretsManager   basetypes.ObjectValue `tfsdk:"aws_secrets_manager"`
	AzureKeyVault       basetypes.ObjectValue `tfsdk:"azure_key_vault"`
	EnvironmentVariable basetypes.ObjectValue `tfsdk:"environment_variable"`
	SecretFile          basetypes.ObjectValue `tfsdk:"secret_file"`
	TaskCount           types.Int64           `tfsdk:"task_count"`
	CreatedAt           types.String          `tfsdk:"created_at"`
	UpdatedAt           types.String          `tfsdk:"updated_at"`
}

var awsAttrTypes = map[string]attr.Type{
	"iam_role":     types.StringType,
	"secrets_path": types.StringType,
}

var azureAttrTypes = map[string]attr.Type{
	"key_vault_uri": types.StringType,
	"secret_name":   types.StringType,
}

var envVarAttrTypes = map[string]attr.Type{
	"variable_name": types.StringType,
}

var secretFileAttrTypes = map[string]attr.Type{
	"path": types.StringType,
}

func (r *ServiceUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (r *ServiceUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a repository service user for agent task authentication. Exactly one credential provider must be configured.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier (repo_name:username).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo_name": schema.StringAttribute{
				Description: "Name of the repository this service user belongs to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Database username. Must be unique within the repository.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_secrets_manager": schema.SingleNestedAttribute{
				Description: "AWS Secrets Manager credential provider.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"secrets_path": schema.StringAttribute{
						Description: "Path or name of the secret in AWS Secrets Manager.",
						Required:    true,
					},
					"iam_role": schema.StringAttribute{
						Description: "ARN of an IAM role to assume when retrieving the secret.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
			"azure_key_vault": schema.SingleNestedAttribute{
				Description: "Azure Key Vault credential provider.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"key_vault_uri": schema.StringAttribute{
						Description: "HTTPS URL of the Azure Key Vault.",
						Required:    true,
					},
					"secret_name": schema.StringAttribute{
						Description: "Name of the secret within the vault.",
						Required:    true,
					},
				},
			},
			"environment_variable": schema.SingleNestedAttribute{
				Description: "Environment variable credential provider.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"variable_name": schema.StringAttribute{
						Description: "Name of the OS environment variable containing the secret.",
						Required:    true,
					},
				},
			},
			"secret_file": schema.SingleNestedAttribute{
				Description: "Secret file credential provider. Reads from /altr/secrets/<path> at runtime.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						Description: "Simple filename (no path separators). Resolved under /altr/secrets/ at runtime.",
						Required:    true,
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *ServiceUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = c
}

func (r *ServiceUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.validateCredentialProvider(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.CreateServiceUserInput{
		Username: plan.Username.ValueString(),
	}

	r.applyCredentialProviderToCreateInput(&plan, &input)

	su, err := r.client.CreateServiceUser(plan.RepoName.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service user",
			"Could not create service user, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapServiceUserToModel(su, &plan)

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.RepoName.ValueString(), plan.Username.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServiceUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	su, err := r.client.GetServiceUser(state.RepoName.ValueString(), state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service user",
			"Could not read service user "+state.Username.ValueString()+" in repo "+state.RepoName.ValueString()+": "+err.Error(),
		)

		return
	}

	if su == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	r.mapServiceUserToModel(su, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ServiceUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  ServiceUserResourceModel
		state ServiceUserResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.validateCredentialProvider(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.UpdateServiceUserInput{}

	r.applyCredentialProviderToUpdateInput(&plan, &input)

	su, err := r.client.UpdateServiceUser(state.RepoName.ValueString(), state.Username.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service user",
			"Could not update service user, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapServiceUserToModel(su, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ServiceUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServiceUser(state.RepoName.ValueString(), state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting service user",
			"Could not delete service user. Note: a service user cannot be deleted while it has active tasks. Error: "+err.Error(),
		)

		return
	}
}

func (r *ServiceUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: repo_name:username",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repo_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *ServiceUserResource) validateCredentialProvider(model *ServiceUserResourceModel) error {
	count := 0

	if !model.AWSSecretsManager.IsNull() {
		count++
	}

	if !model.AzureKeyVault.IsNull() {
		count++
	}

	if !model.EnvironmentVariable.IsNull() {
		count++
	}

	if !model.SecretFile.IsNull() {
		count++
	}

	if count == 0 {
		return errors.New("exactly one credential provider must be specified (aws_secrets_manager, azure_key_vault, environment_variable, or secret_file)")
	}

	if count > 1 {
		return errors.New("only one credential provider can be specified at a time")
	}

	return nil
}

// credentialsFromModel extracts the active credential provider from the model
// and returns the four credential fields used by both create and update inputs.
func (r *ServiceUserResource) credentialsFromModel(model *ServiceUserResourceModel) (
	aws *client.AWSSecretsManager,
	azure *client.AzureKeyVault,
	envVar *client.EnvironmentVariable,
	sf *client.SecretFile,
) {
	if !model.AWSSecretsManager.IsNull() {
		secretsPath := model.AWSSecretsManager.Attributes()["secrets_path"].(types.String)
		aws = &client.AWSSecretsManager{SecretsPath: secretsPath.ValueString()}

		iamRole := model.AWSSecretsManager.Attributes()["iam_role"].(types.String)
		if !iamRole.IsNull() && iamRole.ValueString() != "" {
			aws.IAMRole = iamRole.ValueString()
		}
	}

	if !model.AzureKeyVault.IsNull() {
		azure = &client.AzureKeyVault{
			KeyVaultURI: model.AzureKeyVault.Attributes()["key_vault_uri"].(types.String).ValueString(),
			SecretName:  model.AzureKeyVault.Attributes()["secret_name"].(types.String).ValueString(),
		}
	}

	if !model.EnvironmentVariable.IsNull() {
		envVar = &client.EnvironmentVariable{
			VariableName: model.EnvironmentVariable.Attributes()["variable_name"].(types.String).ValueString(),
		}
	}

	if !model.SecretFile.IsNull() {
		sf = &client.SecretFile{
			Path: model.SecretFile.Attributes()["path"].(types.String).ValueString(),
		}
	}

	return
}

func (r *ServiceUserResource) applyCredentialProviderToCreateInput(model *ServiceUserResourceModel, input *client.CreateServiceUserInput) {
	input.AWSSecretsManager, input.AzureKeyVault, input.EnvironmentVariable, input.SecretFile = r.credentialsFromModel(model)
}

func (r *ServiceUserResource) applyCredentialProviderToUpdateInput(model *ServiceUserResourceModel, input *client.UpdateServiceUserInput) {
	input.AWSSecretsManager, input.AzureKeyVault, input.EnvironmentVariable, input.SecretFile = r.credentialsFromModel(model)
}

func (r *ServiceUserResource) mapServiceUserToModel(su *client.ServiceUser, model *ServiceUserResourceModel) {
	model.RepoName = types.StringValue(su.RepoName)
	model.Username = types.StringValue(su.Username)
	model.TaskCount = types.Int64Value(int64(su.TaskCount))
	model.CreatedAt = types.StringValue(su.CreatedAt)
	model.UpdatedAt = types.StringValue(su.UpdatedAt)

	if su.AWSSecretsManager != nil {
		model.AWSSecretsManager = basetypes.NewObjectValueMust(awsAttrTypes, map[string]attr.Value{
			"iam_role":     types.StringValue(su.AWSSecretsManager.IAMRole),
			"secrets_path": types.StringValue(su.AWSSecretsManager.SecretsPath),
		})
	} else {
		model.AWSSecretsManager = basetypes.NewObjectNull(awsAttrTypes)
	}

	if su.AzureKeyVault != nil {
		model.AzureKeyVault = basetypes.NewObjectValueMust(azureAttrTypes, map[string]attr.Value{
			"key_vault_uri": types.StringValue(su.AzureKeyVault.KeyVaultURI),
			"secret_name":   types.StringValue(su.AzureKeyVault.SecretName),
		})
	} else {
		model.AzureKeyVault = basetypes.NewObjectNull(azureAttrTypes)
	}

	if su.EnvironmentVariable != nil {
		model.EnvironmentVariable = basetypes.NewObjectValueMust(envVarAttrTypes, map[string]attr.Value{
			"variable_name": types.StringValue(su.EnvironmentVariable.VariableName),
		})
	} else {
		model.EnvironmentVariable = basetypes.NewObjectNull(envVarAttrTypes)
	}

	if su.SecretFile != nil {
		model.SecretFile = basetypes.NewObjectValueMust(secretFileAttrTypes, map[string]attr.Value{
			"path": types.StringValue(su.SecretFile.Path),
		})
	} else {
		model.SecretFile = basetypes.NewObjectNull(secretFileAttrTypes)
	}
}
