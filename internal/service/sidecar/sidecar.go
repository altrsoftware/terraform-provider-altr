// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package sidecar

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/altrsoftware/terraform-provider-altr/internal/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &SidecarResource{}
	_ resource.ResourceWithImportState = &SidecarResource{}
)

func NewSidecarResource() resource.Resource {
	return &SidecarResource{}
}

type SidecarResource struct {
	client *client.Client
}

type SidecarResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Hostname                 types.String `tfsdk:"hostname"`
	PublicKey1               types.String `tfsdk:"public_key_1"`
	PublicKey2               types.String `tfsdk:"public_key_2"`
	UnsupportedQueryBypass   types.Bool   `tfsdk:"unsupported_query_bypass"`
	DataPlaneURL             types.String `tfsdk:"data_plane_url"`
	ListenerCount            types.Int64  `tfsdk:"listener_count"`
	ListenerRepoBindingCount types.Int64  `tfsdk:"listener_repo_binding_count"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
}

func (r *SidecarResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sidecar"
}

func (r *SidecarResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a sidecar.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Sidecar ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the sidecar.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the sidecar.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(400),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the sidecar.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.HostnameRegexStringRFC1123),
						"must be a valid hostname",
					),
				},
			},
			"public_key_1": schema.StringAttribute{
				Description: "First public key for the sidecar.",
				Optional:    true,
				Computed:    true,
			},
			"public_key_2": schema.StringAttribute{
				Description: "Second public key for the sidecar.",
				Optional:    true,
				Computed:    true,
			},
			"unsupported_query_bypass": schema.BoolAttribute{
				Description: "When true, unsupported queries will bypass the query parser and return all results without applying policy instead of returning an error.",
				Optional:    true,
				Computed:    true,
			},
			"data_plane_url": schema.StringAttribute{
				Description: "Data plane URL of the sidecar.",
				Computed:    true,
			},
			"listener_count": schema.Int64Attribute{
				Description: "Number of listeners for this sidecar.",
				Computed:    true,
			},
			"listener_repo_binding_count": schema.Int64Attribute{
				Description: "Number of listener repo bindings for this sidecar.",
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

// Helper function to validate public key requirements
func (r *SidecarResource) validatePublicKeys(model *SidecarResourceModel) error {
	hasPublicKey1 := !model.PublicKey1.IsNull() && !model.PublicKey1.IsUnknown() && model.PublicKey1.ValueString() != ""
	hasPublicKey2 := !model.PublicKey2.IsNull() && !model.PublicKey2.IsUnknown() && model.PublicKey2.ValueString() != ""

	if !hasPublicKey1 && !hasPublicKey2 {
		return errors.New("at least one of 'public_key_1' or 'public_key_2' must be specified")
	}

	return nil
}

func (r *SidecarResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SidecarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SidecarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one public key is provided
	if err := r.validatePublicKeys(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	// Create the input for the API call
	input := client.CreateSidecarInput{
		Name:                   plan.Name.ValueString(),
		Hostname:               plan.Hostname.ValueString(),
		Description:            plan.Description.ValueString(),
		PublicKey1:             plan.PublicKey1.ValueString(),
		PublicKey2:             plan.PublicKey2.ValueString(),
		UnsupportedQueryBypass: plan.UnsupportedQueryBypass.ValueBool(),
	}

	// Call the API to create the sidecar
	sidecar, err := r.client.CreateSidecar(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating sidecar",
			"Could not create sidecar, unexpected error: "+err.Error(),
		)

		return
	}

	// Map response to the model
	r.mapSidecarToModel(sidecar, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SidecarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SidecarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get sidecar from API
	sidecar, err := r.client.GetSidecar(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sidecar",
			"Could not read sidecar ID "+state.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	// If sidecar doesn't exist, remove it from state
	if sidecar == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	// Map response to the model
	r.mapSidecarToModel(sidecar, &state)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SidecarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  SidecarResourceModel
		state SidecarResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one public key is provided
	if err := r.validatePublicKeys(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	// Create the input for the API call
	input := client.UpdateSidecarInput{}

	// Only set fields that have changed
	if !plan.Name.Equal(state.Name) {
		input.Name = plan.Name.ValueStringPointer()
	}

	if !plan.Description.Equal(state.Description) {
		input.Description = plan.Description.ValueStringPointer()
	}

	if !plan.Hostname.Equal(state.Hostname) {
		input.Hostname = plan.Hostname.ValueStringPointer()
	}

	if !plan.PublicKey1.Equal(state.PublicKey1) {
		input.PublicKey1 = plan.PublicKey1.ValueStringPointer()
	}

	if !plan.PublicKey2.Equal(state.PublicKey2) {
		input.PublicKey2 = plan.PublicKey2.ValueStringPointer()
	}

	if !plan.UnsupportedQueryBypass.Equal(state.UnsupportedQueryBypass) {
		input.UnsupportedQueryBypass = plan.UnsupportedQueryBypass.ValueBoolPointer()
	}

	// Call the API to update the sidecar
	sidecar, err := r.client.UpdateSidecar(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating sidecar",
			"Could not update sidecar, unexpected error: "+err.Error(),
		)

		return
	}

	// Map response to the model
	r.mapSidecarToModel(sidecar, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SidecarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SidecarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the sidecar
	err := r.client.DeleteSidecar(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting sidecar",
			"Could not delete sidecar, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *SidecarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map API response to Terraform model
func (r *SidecarResource) mapSidecarToModel(sidecar *client.Sidecar, model *SidecarResourceModel) {
	model.ID = types.StringValue(sidecar.ID)
	model.Name = types.StringValue(sidecar.Name)
	model.Description = types.StringValue(sidecar.Description)
	model.Hostname = types.StringValue(sidecar.Hostname)
	model.DataPlaneURL = types.StringValue(sidecar.DataPlaneURL)
	model.ListenerCount = types.Int64Value(int64(sidecar.ListenerCount))
	model.ListenerRepoBindingCount = types.Int64Value(int64(sidecar.ListenerRepoBindingCount))
	model.UnsupportedQueryBypass = types.BoolValue(sidecar.UnsupportedQueryBypass)
	model.CreatedAt = types.StringValue(sidecar.CreatedAt)
	model.UpdatedAt = types.StringValue(sidecar.UpdatedAt)

	// Handle public keys - only set if they exist and contain data
	if sidecar.PublicKey1 != nil && sidecar.PublicKey1.RSAKey != "" {
		model.PublicKey1 = types.StringValue(sidecar.PublicKey1.RSAKey)
	} else {
		model.PublicKey1 = types.StringNull()
	}

	if sidecar.PublicKey2 != nil && sidecar.PublicKey2.RSAKey != "" {
		model.PublicKey2 = types.StringValue(sidecar.PublicKey2.RSAKey)
	} else {
		model.PublicKey2 = types.StringNull()
	}
}
