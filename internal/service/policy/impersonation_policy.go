// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"fmt"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	customvalidation "github.com/altrsoftware/terraform-provider-altr/internal/validation"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ImpersonationPolicyResource{}
	_ resource.ResourceWithImportState = &ImpersonationPolicyResource{}
)

func NewImpersonationPolicyResource() resource.Resource {
	return &ImpersonationPolicyResource{}
}

type ImpersonationPolicyResource struct {
	client *client.Client
}

type ImpersonationPolicyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	RepoName    types.String `tfsdk:"repo_name"`
	Rules       types.List   `tfsdk:"rules"` // List of rules for the impersonation policy
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *ImpersonationPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_impersonation_policy"
}

func (r *ImpersonationPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an impersonation policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the impersonation policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the impersonation policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the impersonation policy.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"repo_name": schema.StringAttribute{
				Description: "The name of the repository where the impersonation policy is defined.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"rules": schema.ListNestedAttribute{
				Description: "List of rules for the impersonation policy.",
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"actors": schema.ListNestedAttribute{
							Description: "List of actors for the rule.",
							Required:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the actor (e.g., idp_user, idp_group).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(OltpActorTypes...),
										},
									},
									"identifiers": schema.ListAttribute{
										Description: "List of user or group identifiers.",
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.List{
											customvalidation.UniqueStringList(),
										},
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the actor (e.g., equals).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(OltpConditions...),
										},
									},
								},
							},
						},
						"targets": schema.ListNestedAttribute{
							Description: "List of target users or groups.",
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the target (e.g., repo_user).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(OltpTargetTypes...),
										},
									},
									"identifiers": schema.ListAttribute{
										Description: "List of target identifiers.",
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.List{
											customvalidation.UniqueStringList(),
										},
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the target (e.g., equals).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(OltpConditions...),
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

func (r *ImpersonationPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ImpersonationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ImpersonationPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert rules from Terraform model to client model
	rules := convertRulesFromTerraform(plan.Rules)

	// Create the input for the API call
	input := client.CreateImpersonationPolicyInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		RepoName:    plan.RepoName.ValueString(),
		Rules:       rules,
	}

	// Call the API to create the impersonation policy
	policy, err := r.client.CreateImpersonationPolicy(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating impersonation policy",
			"Could not create impersonation policy, unexpected error: "+err.Error(),
		)

		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ImpersonationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ImpersonationPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get impersonation policy from API
	policy, err := r.client.GetImpersonationPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading impersonation policy",
			"Could not read impersonation policy ID "+state.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	// If policy doesn't exist, remove it from state
	if policy == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &state)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ImpersonationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  ImpersonationPolicyResourceModel
		state ImpersonationPolicyResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert rules from Terraform model to client model
	rules := convertRulesFromTerraform(plan.Rules)

	// Create the input for the API call
	input := client.UpdateImpersonationPolicyInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Rules:       rules,
	}

	// Call the API to update the impersonation policy
	policy, err := r.client.UpdateImpersonationPolicy(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating impersonation policy",
			"Could not update impersonation policy, unexpected error: "+err.Error(),
		)

		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ImpersonationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ImpersonationPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the impersonation policy
	err := r.client.DeleteImpersonationPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting impersonation policy",
			"Could not delete impersonation policy, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *ImpersonationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map API response to Terraform model
func (r *ImpersonationPolicyResource) mapPolicyToModel(policy *client.ImpersonationPolicy, model *ImpersonationPolicyResourceModel) {
	model.ID = types.StringValue(policy.ID)
	model.Name = types.StringValue(policy.Name)
	model.Description = types.StringValue(policy.Description)
	model.RepoName = types.StringValue(policy.RepoName)
	model.Rules = convertRulesToTerraform(policy.Rules)
	model.CreatedAt = types.StringValue(policy.CreatedAt)
	model.UpdatedAt = types.StringValue(policy.UpdatedAt)
}

// Helper functions to convert rules between Terraform and client models
func convertRulesFromTerraform(rules types.List) []client.ImpersonationRule {
	if rules.IsNull() || rules.IsUnknown() {
		return nil
	}

	var clientRules []client.ImpersonationRule

	// Iterate over the elements in the rules list
	for _, rule := range rules.Elements() {
		// Ensure the rule is a valid object
		ruleObj, ok := rule.(types.Object)
		if !ok {
			continue // Skip invalid rules
		}

		// Retrieve attributes from the rule object
		ruleAttrs := ruleObj.Attributes()

		// Extract actors
		actorsAttr, ok := ruleAttrs["actors"]
		if !ok {
			continue // Skip if actors are missing
		}

		actorsList, ok := actorsAttr.(types.List)
		if !ok {
			continue // Skip if actors are not a list
		}

		actors := convertActorsFromTerraform(actorsList)

		// Extract targets
		targetsAttr, ok := ruleAttrs["targets"]
		if !ok {
			continue // Skip if targets are missing
		}

		targetsList, ok := targetsAttr.(types.List)
		if !ok {
			continue // Skip if targets are not a list
		}

		targets := convertActorsFromTerraform(targetsList)

		// Append the rule to the client rules
		clientRules = append(clientRules, client.ImpersonationRule{
			Actors:  actors,
			Targets: targets,
		})
	}

	return clientRules
}

// Helper function to convert actors from Terraform to client model
func convertActorsFromTerraform(actorsList types.List) []client.Actor {
	if actorsList.IsNull() || actorsList.IsUnknown() {
		return nil
	}

	var actors []client.Actor

	// Iterate over the elements in the actors list
	for _, actor := range actorsList.Elements() {
		actorObj, ok := actor.(types.Object)
		if !ok {
			continue // Skip invalid actors
		}

		// Retrieve attributes from the actor object
		actorAttrs := actorObj.Attributes()

		// Extract actor attributes
		actorType, _ := actorAttrs["type"].(types.String)
		identifiersAttr, _ := actorAttrs["identifiers"].(types.List)
		condition, _ := actorAttrs["condition"].(types.String)

		// Convert identifiers to a string slice
		identifiers := extractStringList(identifiersAttr)

		// Append the actor to the list
		actors = append(actors, client.Actor{
			Type:        actorType.ValueString(),
			Identifiers: identifiers,
			Condition:   condition.ValueString(),
		})
	}

	return actors
}

// Helper function to extract a list of strings from a Terraform list
func extractStringList(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string

	for _, element := range list.Elements() {
		str, ok := element.(types.String)
		if ok && !str.IsNull() && !str.IsUnknown() {
			result = append(result, str.ValueString())
		}
	}

	return result
}

func convertRulesToTerraform(rules []client.ImpersonationRule) types.List {
	if len(rules) == 0 {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"actors": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
				},
				"targets": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
				},
			},
		})
	}

	var (
		terraformRules []attr.Value
		diagnostics    diag.Diagnostics
	)

	for _, rule := range rules {
		// Convert actors
		var terraformActors []attr.Value
		for _, actor := range rule.Actors {
			actorValue, actorDiags := types.ObjectValue(
				map[string]attr.Type{
					"type":        types.StringType,
					"identifiers": types.ListType{ElemType: types.StringType},
					"condition":   types.StringType,
				},
				map[string]attr.Value{
					"type":        types.StringValue(actor.Type),
					"identifiers": convertStringListToTerraform(actor.Identifiers),
					"condition":   types.StringValue(actor.Condition),
				},
			)
			diagnostics.Append(actorDiags...)

			terraformActors = append(terraformActors, actorValue)
		}

		// Convert targets
		var terraformTargets []attr.Value
		for _, target := range rule.Targets {
			targetValue, targetDiags := types.ObjectValue(
				map[string]attr.Type{
					"type":        types.StringType,
					"identifiers": types.ListType{ElemType: types.StringType},
					"condition":   types.StringType,
				},
				map[string]attr.Value{
					"type":        types.StringValue(target.Type),
					"identifiers": convertStringListToTerraform(target.Identifiers),
					"condition":   types.StringValue(target.Condition),
				},
			)
			diagnostics.Append(targetDiags...)

			terraformTargets = append(terraformTargets, targetValue)
		}

		// Append the rule to the Terraform rules
		ruleValue, ruleDiags := types.ObjectValue(
			map[string]attr.Type{
				"actors": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
				},
				"targets": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
				},
			},
			map[string]attr.Value{
				"actors": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
					terraformActors,
				),
				"targets": types.ListValueMust(
					types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
					terraformTargets,
				),
			},
		)
		diagnostics.Append(ruleDiags...)

		terraformRules = append(terraformRules, ruleValue)
	}

	// Handle diagnostics (if needed)
	if diagnostics.HasError() {
		tflog.Error(context.Background(), "Diagnostics encountered during conversion", map[string]interface{}{
			"diagnostics": diagnostics,
		})
	}

	return types.ListValueMust(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"actors": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
				},
				"targets": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type":        types.StringType,
							"identifiers": types.ListType{ElemType: types.StringType},
							"condition":   types.StringType,
						},
					},
				},
			},
		},
		terraformRules,
	)
}

// Helper function to convert a slice of strings to a Terraform list
func convertStringListToTerraform(strings []string) types.List {
	if len(strings) == 0 {
		return types.ListNull(types.StringType)
	}

	var terraformStrings []attr.Value
	for _, str := range strings {
		terraformStrings = append(terraformStrings, types.StringValue(str))
	}

	return types.ListValueMust(types.StringType, terraformStrings)
}
