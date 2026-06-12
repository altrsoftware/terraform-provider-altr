// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	ID                  types.String          `tfsdk:"id"`
	RepoName            types.String          `tfsdk:"repo_name"`
	Username            types.String          `tfsdk:"username"`
	AWSSecretsManager   basetypes.ObjectValue `tfsdk:"aws_secrets_manager"`
	AzureKeyVault       basetypes.ObjectValue `tfsdk:"azure_key_vault"`
	EnvironmentVariable basetypes.ObjectValue `tfsdk:"environment_variable"`
	SecretFile          basetypes.ObjectValue `tfsdk:"secret_file"`
	CreatedAt           types.String          `tfsdk:"created_at"`
	UpdatedAt           types.String          `tfsdk:"updated_at"`
}

func (r *RepoUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_user"
}

func (r *RepoUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
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
	}

	for name, attribute := range credentialProviderSchemaAttributes() {
		attributes[name] = attribute
	}

	resp.Schema = schema.Schema{
		Description: "Manages a repository user with credential storage configuration. Exactly one credential provider must be configured.",
		Attributes:  attributes,
	}
}

func (r *RepoUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepoUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepoUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateSingleCredentialProvider(plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.CreateRepoUserInput{
		Username: plan.Username.ValueString(),
	}

	input.AWSSecretsManager, input.AzureKeyVault, input.EnvironmentVariable, input.SecretFile = credentialProvidersFromObjects(
		plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile,
	)

	repoUser, err := r.client.CreateRepoUser(plan.RepoName.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository user",
			"Could not create repository user, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapRepoUserToModel(repoUser, &plan)

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.RepoName.ValueString(), plan.Username.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepoUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	repoUser, err := r.client.GetRepoUser(state.RepoName.ValueString(), state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository user",
			"Could not read repository user "+state.Username.ValueString()+" in repo "+state.RepoName.ValueString()+": "+err.Error(),
		)

		return
	}

	if repoUser == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	r.mapRepoUserToModel(repoUser, &state)

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

	if err := validateSingleCredentialProvider(plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.UpdateRepoUserInput{}

	input.AWSSecretsManager, input.AzureKeyVault, input.EnvironmentVariable, input.SecretFile = credentialProvidersFromObjects(
		plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile,
	)

	repoUser, err := r.client.UpdateRepoUser(state.RepoName.ValueString(), state.Username.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating repository user",
			"Could not update repository user, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapRepoUserToModel(repoUser, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepoUserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

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

func (r *RepoUserResource) mapRepoUserToModel(repoUser *client.RepoUser, model *RepoUserResourceModel) {
	model.RepoName = types.StringValue(repoUser.RepoName)
	model.Username = types.StringValue(repoUser.Username)
	model.CreatedAt = types.StringValue(repoUser.CreatedAt)
	model.UpdatedAt = types.StringValue(repoUser.UpdatedAt)

	model.AWSSecretsManager, model.AzureKeyVault, model.EnvironmentVariable, model.SecretFile = credentialProvidersToObjects(
		repoUser.AWSSecretsManager, repoUser.AzureKeyVault, repoUser.EnvironmentVariable, repoUser.SecretFile,
	)
}
