// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/altrsoftware/terraform-provider-altr/internal/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &RepoSidecarBindingResource{}
	_ resource.ResourceWithImportState = &RepoSidecarBindingResource{}
)

func NewRepoSidecarBindingResource() resource.Resource {
	return &RepoSidecarBindingResource{}
}

type RepoSidecarBindingResource struct {
	client *client.Client
}

type RepoSidecarBindingResourceModel struct {
	ID        types.String `tfsdk:"id"`
	SidecarID types.String `tfsdk:"sidecar_id"`
	RepoName  types.String `tfsdk:"repo_name"`
	Port      types.Int64  `tfsdk:"port"`
}

func (r *RepoSidecarBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo_sidecar_binding"
}

func (r *RepoSidecarBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a binding between a repository and a sidecar listener port.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the binding (sidecar_id:port:repo_name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sidecar_id": schema.StringAttribute{
				Description: "ID of the sidecar.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo_name": schema.StringAttribute{
				Description: "Name of the repository to bind.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Sidecar listener port to bind to the repository.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RepoSidecarBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepoSidecarBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepoSidecarBindingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to create the repo sidecar binding
	err := r.client.CreateRepoSidecarBinding(
		plan.SidecarID.ValueString(),
		plan.RepoName.ValueString(),
		int(plan.Port.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repo sidecar binding",
			"Could not create repo sidecar binding, unexpected error: "+err.Error(),
		)

		return
	}

	// Set the ID
	plan.ID = types.StringValue(fmt.Sprintf("%s:%d:%s",
		plan.SidecarID.ValueString(),
		plan.Port.ValueInt64(),
		plan.RepoName.ValueString()))

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoSidecarBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepoSidecarBindingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get repo sidecar binding from API
	binding, err := r.client.GetRepoSidecarBinding(
		state.SidecarID.ValueString(),
		state.RepoName.ValueString(),
		int(state.Port.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repo sidecar binding",
			fmt.Sprintf("Could not read repo sidecar binding for sidecar %s, port %d, repo %s: %s",
				state.SidecarID.ValueString(),
				state.Port.ValueInt64(),
				state.RepoName.ValueString(),
				err.Error()),
		)

		return
	}

	// If binding doesn't exist, remove it from state
	if binding == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	// Map response to the model
	r.mapBindingToModel(binding, &state)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *RepoSidecarBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// According to the API specification, repo sidecar bindings cannot be updated
	// They can only be created and deleted
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repo sidecar bindings cannot be updated. Please recreate the resource to make changes.",
	)
}

func (r *RepoSidecarBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepoSidecarBindingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the repo sidecar binding
	err := r.client.DeleteRepoSidecarBinding(
		state.SidecarID.ValueString(),
		state.RepoName.ValueString(),
		int(state.Port.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting repo sidecar binding",
			"Could not delete repo sidecar binding, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *RepoSidecarBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import ID format: "sidecar_id:port:repo_name"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: sidecar_id:port:repo_name",
		)

		return
	}

	sidecarID := parts[0]
	portStr := parts[1]
	repoName := parts[2]

	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Port in Import ID",
			"Port must be a valid integer: "+err.Error(),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sidecar_id"), sidecarID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), port)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repo_name"), repoName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Helper function to map API response to Terraform model
func (r *RepoSidecarBindingResource) mapBindingToModel(binding *client.RepoSidecarBinding, model *RepoSidecarBindingResourceModel) {
	model.SidecarID = types.StringValue(binding.SidecarID)
	model.RepoName = types.StringValue(binding.RepoName)
	model.Port = types.Int64Value(int64(binding.Port))

	// Set the ID
	model.ID = types.StringValue(fmt.Sprintf("%s:%d:%s",
		binding.SidecarID,
		binding.Port,
		binding.RepoName))
}
