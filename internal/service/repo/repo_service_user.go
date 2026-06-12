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
	Resource            types.String          `tfsdk:"resource"`
	AWSSecretsManager   basetypes.ObjectValue `tfsdk:"aws_secrets_manager"`
	AzureKeyVault       basetypes.ObjectValue `tfsdk:"azure_key_vault"`
	EnvironmentVariable basetypes.ObjectValue `tfsdk:"environment_variable"`
	SecretFile          basetypes.ObjectValue `tfsdk:"secret_file"`
	TaskCount           types.Int64           `tfsdk:"task_count"`
	CreatedAt           types.String          `tfsdk:"created_at"`
	UpdatedAt           types.String          `tfsdk:"updated_at"`
}

func (r *ServiceUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (r *ServiceUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
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
		"resource": schema.StringAttribute{
			Description: "Database entrypoint the agent connects to with this service user (e.g. an Oracle service name like \"ORCL\"). This is the real database identifier, which may differ from repo_name (the ALTR-side name of the repository). Once connected, the agent fans out to other accessible databases.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
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
	}

	for name, attribute := range credentialProviderSchemaAttributes() {
		attributes[name] = attribute
	}

	resp.Schema = schema.Schema{
		Description: "Manages a repository service user for agent task authentication. Exactly one credential provider must be configured.",
		Attributes:  attributes,
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

	if err := validateSingleCredentialProvider(plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.CreateServiceUserInput{
		Username: plan.Username.ValueString(),
		Resource: plan.Resource.ValueString(),
	}

	input.AWSSecretsManager, input.AzureKeyVault, input.EnvironmentVariable, input.SecretFile = credentialProvidersFromObjects(
		plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile,
	)

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

	if err := validateSingleCredentialProvider(plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.UpdateServiceUserInput{
		Resource: plan.Resource.ValueString(),
	}

	input.AWSSecretsManager, input.AzureKeyVault, input.EnvironmentVariable, input.SecretFile = credentialProvidersFromObjects(
		plan.AWSSecretsManager, plan.AzureKeyVault, plan.EnvironmentVariable, plan.SecretFile,
	)

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

func (r *ServiceUserResource) mapServiceUserToModel(su *client.ServiceUser, model *ServiceUserResourceModel) {
	model.RepoName = types.StringValue(su.RepoName)
	model.Username = types.StringValue(su.Username)
	model.TaskCount = types.Int64Value(int64(su.TaskCount))
	model.CreatedAt = types.StringValue(su.CreatedAt)
	model.UpdatedAt = types.StringValue(su.UpdatedAt)

	// Only overwrite when the API echoes resource, so a response that omits it
	// does not clobber the configured value and cause a perpetual diff.
	if su.Resource != "" {
		model.Resource = types.StringValue(su.Resource)
	}

	model.AWSSecretsManager, model.AzureKeyVault, model.EnvironmentVariable, model.SecretFile = credentialProvidersToObjects(
		su.AWSSecretsManager, su.AzureKeyVault, su.EnvironmentVariable, su.SecretFile,
	)
}
