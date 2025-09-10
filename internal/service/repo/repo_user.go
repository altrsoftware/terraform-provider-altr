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
	_ resource.Resource                = &RepoUserResource{}
	_ resource.ResourceWithImportState = &RepoUserResource{}
)

func NewRepoUserResource() resource.Resource {
	return &RepoUserResource{}
}

type RepoUserResource struct {
	client *client.Client
}

type RepoUserResourceModel struct {
	ID                types.String          `tfsdk:"id"`
	RepoName          types.String          `tfsdk:"repo_name"`
	Username          types.String          `tfsdk:"username"`
	AWSSecretsManager basetypes.ObjectValue `tfsdk:"aws_secrets_manager"`
	AzureKeyVault     basetypes.ObjectValue `tfsdk:"azure_key_vault"`
	CreatedAt         types.String          `tfsdk:"created_at"`
	UpdatedAt         types.String          `tfsdk:"updated_at"`
}

type AWSSecretsManagerResourceModel struct {
	IAMRole     types.String `tfsdk:"iam_role"`
	SecretsPath types.String `tfsdk:"secrets_path"`
}

type AzureKeyVaultResourceModel struct {
	KeyVaultURI types.String `tfsdk:"key_vault_uri"`
	SecretName  types.String `tfsdk:"secret_name"`
}

func (r *RepoUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_user"
}

func (r *RepoUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a repository user with credential storage configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the repo user (repo_name:username).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo_name": schema.StringAttribute{
				Description: "Name of the repository.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Username for the repository user.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_secrets_manager": schema.SingleNestedAttribute{
				Description: "AWS Secrets Manager configuration for storing credentials.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"iam_role": schema.StringAttribute{
						Description: "IAM role ARN for accessing the secret.",
						Optional:    true,
						Computed:    true,
					},
					"secrets_path": schema.StringAttribute{
						Description: "Path to the secret in AWS Secrets Manager.",
						Required:    true,
					},
				},
			},
			"azure_key_vault": schema.SingleNestedAttribute{
				Description: "Azure Key Vault configuration for storing credentials.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"key_vault_uri": schema.StringAttribute{
						Description: "URI of the Azure Key Vault.",
						Required:    true,
					},
					"secret_name": schema.StringAttribute{
						Description: "Name of the secret in Azure Key Vault.",
						Required:    true,
					},
				},
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

func (r *RepoUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *RepoUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepoUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one credential store is specified
	if err := r.validateCredentialStore(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	// Create the input for the API call
	input := client.CreateRepoUserInput{
		Username: plan.Username.ValueString(),
	}

	// Set credential store configuration
	if !plan.AWSSecretsManager.IsNull() {
		secretsPath := plan.AWSSecretsManager.Attributes()["secrets_path"].(types.String)
		if secretsPath.IsNull() || secretsPath.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid Configuration", "aws_secrets_manager.secrets_path must be specified and non-empty")

			return
		}

		input.AWSSecretsManager = &client.AWSSecretsManager{
			SecretsPath: secretsPath.ValueString(),
		}

		iamRole := plan.AWSSecretsManager.Attributes()["iam_role"].(types.String)
		if !iamRole.IsNull() && iamRole.ValueString() != "" {
			input.AWSSecretsManager.IAMRole = iamRole.ValueString()
		}
	} else {
		plan.AWSSecretsManager = basetypes.NewObjectNull(
			map[string]attr.Type{},
		)
	}

	if !plan.AzureKeyVault.IsNull() {
		keyVaultURI := plan.AzureKeyVault.Attributes()["key_vault_uri"].(types.String)
		secretName := plan.AzureKeyVault.Attributes()["secret_name"].(types.String)

		if keyVaultURI.IsNull() || keyVaultURI.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid Configuration", "azure_key_vault.key_vault_uri must be specified and non-empty")

			return
		}

		if secretName.IsNull() || secretName.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid Configuration", "azure_key_vault.secret_name must be specified and non-empty")

			return
		}

		input.AzureKeyVault = &client.AzureKeyVault{
			KeyVaultURI: keyVaultURI.ValueString(),
			SecretName:  secretName.ValueString(),
		}
	} else {
		plan.AzureKeyVault = basetypes.NewObjectNull(
			map[string]attr.Type{},
		)
	}

	// Call the API to create the repo user
	repoUser, err := r.client.CreateRepoUser(plan.RepoName.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository user",
			"Could not create repository user, unexpected error: "+err.Error(),
		)

		return
	}

	// Map response to the model
	r.mapRepoUserToModel(repoUser, &plan)

	fmt.Println("Created repo user:", plan.RepoName.ValueString(), plan.Username.ValueString())

	// Set the ID
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.RepoName.ValueString(), plan.Username.ValueString()))

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepoUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get repo user from API
	repoUser, err := r.client.GetRepoUser(state.RepoName.ValueString(), state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository user",
			"Could not read repository user "+state.Username.ValueString()+" in repo "+state.RepoName.ValueString()+": "+err.Error(),
		)

		return
	}

	// If repo user doesn't exist, remove it from state
	if repoUser == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	// Map response to the model
	r.mapRepoUserToModel(repoUser, &state)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *RepoUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  RepoUserResourceModel
		state RepoUserResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that exactly one credential store is specified
	if err := r.validateCredentialStore(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	// Create the input for the API call
	input := client.UpdateRepoUserInput{}

	if !plan.AWSSecretsManager.Equal(state.AWSSecretsManager) {
		if !plan.AWSSecretsManager.IsNull() {
			secretsPath := plan.AWSSecretsManager.Attributes()["secrets_path"].(types.String)
			input.AWSSecretsManager = &client.AWSSecretsManager{
				SecretsPath: secretsPath.ValueString(),
			}

			iamRole := plan.AWSSecretsManager.Attributes()["iam_role"].(types.String)
			if !iamRole.IsNull() && iamRole.ValueString() != "" {
				input.AWSSecretsManager.IAMRole = iamRole.ValueString()
			}
		} else {
			input.AWSSecretsManager = &client.AWSSecretsManager{}
		}
	}

	if !plan.AzureKeyVault.Equal(state.AzureKeyVault) {
		if !plan.AzureKeyVault.IsNull() {
			keyVaultURI := plan.AzureKeyVault.Attributes()["key_vault_uri"].(types.String)
			secretName := plan.AzureKeyVault.Attributes()["secret_name"].(types.String)
			input.AzureKeyVault = &client.AzureKeyVault{
				KeyVaultURI: keyVaultURI.ValueString(),
				SecretName:  secretName.ValueString(),
			}
		} else {
			input.AzureKeyVault = &client.AzureKeyVault{}
		}
	}

	// Call the API to update the repo user
	repoUser, err := r.client.UpdateRepoUser(state.RepoName.ValueString(), state.Username.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating repository user",
			"Could not update repository user, unexpected error: "+err.Error(),
		)

		return
	}

	// Map response to the model
	r.mapRepoUserToModel(repoUser, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepoUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the repo user
	err := r.client.DeleteRepoUser(state.RepoName.ValueString(), state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting repository user",
			"Could not delete repository user, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *RepoUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import ID format: "repo_name:username"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: repo_name:username",
		)

		return
	}

	repoName := parts[0]
	username := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repo_name"), repoName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), username)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Helper function to validate that exactly one credential store is configured
func (r *RepoUserResource) validateCredentialStore(model *RepoUserResourceModel) error {
	credentialStores := 0

	if !model.AWSSecretsManager.IsNull() {
		credentialStores++
	}

	if !model.AzureKeyVault.IsNull() {
		credentialStores++
	}

	if credentialStores == 0 {
		return errors.New("exactly one credential store must be specified (aws_secrets_manager or azure_key_vault)")
	}

	if credentialStores > 1 {
		return errors.New("only one credential store can be specified at a time")
	}

	return nil
}

// Helper function to map API response to Terraform model
func (r *RepoUserResource) mapRepoUserToModel(repoUser *client.RepoUser, model *RepoUserResourceModel) {
	model.RepoName = types.StringValue(repoUser.RepoName)
	model.Username = types.StringValue(repoUser.Username)
	model.CreatedAt = types.StringValue(repoUser.CreatedAt)
	model.UpdatedAt = types.StringValue(repoUser.UpdatedAt)

	// Handle AWS Secrets Manager
	if repoUser.AWSSecretsManager != nil && (repoUser.AWSSecretsManager.IAMRole != "" || repoUser.AWSSecretsManager.SecretsPath != "") {
		model.AWSSecretsManager = basetypes.NewObjectValueMust(
			map[string]attr.Type{
				"iam_role":     types.StringType,
				"secrets_path": types.StringType,
			},
			map[string]attr.Value{
				"iam_role":     types.StringValue(repoUser.AWSSecretsManager.IAMRole),
				"secrets_path": types.StringValue(repoUser.AWSSecretsManager.SecretsPath),
			},
		)
	} else {
		model.AWSSecretsManager = basetypes.NewObjectNull(
			map[string]attr.Type{
				"iam_role":     types.StringType,
				"secrets_path": types.StringType,
			},
		)
	}

	// Handle Azure Key Vault
	if repoUser.AzureKeyVault != nil && (repoUser.AzureKeyVault.KeyVaultURI != "" || repoUser.AzureKeyVault.SecretName != "") {
		model.AzureKeyVault = basetypes.NewObjectValueMust(
			map[string]attr.Type{
				"key_vault_uri": types.StringType,
				"secret_name":   types.StringType,
			},
			map[string]attr.Value{
				"key_vault_uri": types.StringValue(repoUser.AzureKeyVault.KeyVaultURI),
				"secret_name":   types.StringValue(repoUser.AzureKeyVault.SecretName),
			},
		)
	} else {
		model.AzureKeyVault = basetypes.NewObjectNull(
			map[string]attr.Type{
				"key_vault_uri": types.StringType,
				"secret_name":   types.StringType,
			},
		)
	}
}
