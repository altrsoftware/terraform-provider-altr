// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
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
	_ resource.Resource                = &AgentResource{}
	_ resource.ResourceWithImportState = &AgentResource{}
)

func NewAgentResource() resource.Resource {
	return &AgentResource{}
}

type AgentResource struct {
	client *client.Client
}

type AgentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	PublicKey1   types.String `tfsdk:"public_key_1"`
	PublicKey2   types.String `tfsdk:"public_key_2"`
	TaskCount    types.Int64  `tfsdk:"task_count"`
	DataPlaneURL types.String `tfsdk:"data_plane_url"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *AgentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *AgentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ALTR CLASSIFIER agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Agent UUID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Agent type. Must be 'CLASSIFIER'. Cannot be changed after creation.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("CLASSIFIER"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the agent. Must be unique within the organization.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the agent.",
				Optional:    true,
				Computed:    true,
			},
			"public_key_1": schema.StringAttribute{
				Description: "PEM-encoded RSA public key used by the agent for authentication. At least one public key is required at creation.",
				Optional:    true,
				Computed:    true,
			},
			"public_key_2": schema.StringAttribute{
				Description: "Optional second PEM-encoded RSA public key for key rotation.",
				Optional:    true,
				Computed:    true,
			},
			"task_count": schema.Int64Attribute{
				Description: "Number of tasks currently assigned to this agent.",
				Computed:    true,
			},
			"data_plane_url": schema.StringAttribute{
				Description: "URL of the data plane this agent connects to.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *AgentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AgentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.validatePublicKeys(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.CreateAgentInput{
		Type:        plan.Type.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.PublicKey1.IsNull() && plan.PublicKey1.ValueString() != "" {
		input.PublicKey1 = plan.PublicKey1.ValueStringPointer()
	}

	if !plan.PublicKey2.IsNull() && plan.PublicKey2.ValueString() != "" {
		input.PublicKey2 = plan.PublicKey2.ValueStringPointer()
	}

	a, err := r.client.CreateAgent(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating agent",
			"Could not create agent, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapAgentToModel(a, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AgentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	a, err := r.client.GetAgent(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading agent",
			"Could not read agent "+state.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	if a == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	r.mapAgentToModel(a, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *AgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  AgentResourceModel
		state AgentResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.validatePublicKeys(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.UpdateAgentInput{}

	if !plan.Name.Equal(state.Name) {
		input.Name = plan.Name.ValueStringPointer()
	}

	if !plan.Description.Equal(state.Description) {
		input.Description = plan.Description.ValueStringPointer()
	}

	if !plan.PublicKey1.Equal(state.PublicKey1) {
		if !plan.PublicKey1.IsNull() && plan.PublicKey1.ValueString() != "" {
			input.PublicKey1 = plan.PublicKey1.ValueStringPointer()
		}
	}

	if !plan.PublicKey2.Equal(state.PublicKey2) {
		if !plan.PublicKey2.IsNull() && plan.PublicKey2.ValueString() != "" {
			input.PublicKey2 = plan.PublicKey2.ValueStringPointer()
		}
	}

	a, err := r.client.UpdateAgent(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating agent",
			"Could not update agent, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapAgentToModel(a, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AgentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAgent(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting agent",
			"Could not delete agent. Note: an agent cannot be deleted while it has tasks. Error: "+err.Error(),
		)

		return
	}
}

func (r *AgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AgentResource) validatePublicKeys(model *AgentResourceModel) error {
	has1 := !model.PublicKey1.IsNull() && !model.PublicKey1.IsUnknown() && model.PublicKey1.ValueString() != ""
	has2 := !model.PublicKey2.IsNull() && !model.PublicKey2.IsUnknown() && model.PublicKey2.ValueString() != ""

	if !has1 && !has2 {
		return errors.New("at least one of 'public_key_1' or 'public_key_2' must be specified")
	}

	return nil
}

func (r *AgentResource) mapAgentToModel(a *client.Agent, model *AgentResourceModel) {
	model.ID = types.StringValue(a.ID)
	model.Type = types.StringValue(a.Type)
	model.Name = types.StringValue(a.Name)
	model.Description = types.StringValue(a.Description)
	model.TaskCount = types.Int64Value(int64(a.TaskCount))
	model.DataPlaneURL = types.StringValue(a.DataPlaneURL)
	model.CreatedAt = types.StringValue(a.CreatedAt)
	model.UpdatedAt = types.StringValue(a.UpdatedAt)

	if a.PublicKey1 != nil && a.PublicKey1.RSAKey != "" {
		model.PublicKey1 = types.StringValue(a.PublicKey1.RSAKey)
	} else {
		model.PublicKey1 = types.StringNull()
	}

	if a.PublicKey2 != nil && a.PublicKey2.RSAKey != "" {
		model.PublicKey2 = types.StringValue(a.PublicKey2.RSAKey)
	} else {
		model.PublicKey2 = types.StringNull()
	}
}
