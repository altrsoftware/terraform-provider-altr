package policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-altr/internal/client"
	customvalidation "terraform-provider-altr/internal/validation"
)

var _ resource.Resource = &AccessManagementOLTPPolicyResource{}
var _ resource.ResourceWithImportState = &AccessManagementOLTPPolicyResource{}

func NewAccessManagementOltpPolicyDataResource() resource.Resource {
	return &AccessManagementOLTPPolicyResource{}
}

type AccessManagementOLTPPolicyResource struct {
	client *client.Client
}

type AccessManagementOLTPPolicyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Rules            types.List   `tfsdk:"rules"`
	RepoName         types.String `tfsdk:"repo_name"`
	CaseSensitivity  types.String `tfsdk:"case_sensitivity"`
	DatabaseType     types.Int64  `tfsdk:"database_type"`
	DatabaseTypeName types.String `tfsdk:"database_type_name"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// Shared Terraform type definitions for OLTP policy
var OLTPIdentifierPartType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":     types.StringType,
		"wildcard": types.BoolType,
	},
}

var OLTPIdentifierType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database": OLTPIdentifierPartType,
		"schema":   OLTPIdentifierPartType,
		"table":    OLTPIdentifierPartType,
		"column":   OLTPIdentifierPartType,
	},
}

var OLTPObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"identifiers": types.ListType{
			ElemType: OLTPIdentifierType,
		},
	},
}

var OLTPActorType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":      types.StringType,
		"condition": types.StringType,
		"identifiers": types.ListType{
			ElemType: types.StringType,
		},
	},
}

var OLTPRuleType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"actors": types.ListType{
			ElemType: OLTPActorType,
		},
		"objects": types.ListType{
			ElemType: OLTPObjectType,
		},
	},
}

func (d *AccessManagementOLTPPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_management_oltp_policy"
}

func (d *AccessManagementOLTPPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an OLTP access management policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the OLTP access management policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the OLTP access management policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the OLTP access management policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo_name": schema.StringAttribute{
				Description: "The name of the repository this policy belongs to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"case_sensitivity": schema.StringAttribute{
				Description: "Case sensitivity for the policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("case_sensitive"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_type": schema.Int64Attribute{
				Description: "Database type ID for the policy.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.OneOf(4, 1, 2),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"database_type_name": schema.StringAttribute{
				Description: "Database type name for the policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("oracle", "mssql", "mysql", "postgres"),
				},
			},
			"rules": schema.ListNestedAttribute{
				Description: "List of rules for the OLTP access management policy.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Type of the rule.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("read", "update", "delete", "create"),
							},
						},
						"actors": schema.ListNestedAttribute{
							Description: "List of actors for the rule.",
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the actor.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("idp_user", "idp_group"),
										},
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the actor.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("equals"),
										},
									},
									"identifiers": schema.ListAttribute{
										Description: "List of identifiers for the actor.",
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.List{
											customvalidation.UniqueStringList(),
										},
									},
								},
							},
						},
						"objects": schema.ListNestedAttribute{
							Description: "List of objects for the rule.",
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the object.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("database", "schema", "table", "column"),
										},
									},
									"identifiers": schema.ListNestedAttribute{
										Description: "List of identifiers for the object.",
										Required:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"database": schema.SingleNestedAttribute{
													Description: "Database identifier part.",
													Optional:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the database.",
															Optional:    true,
															Validators: []validator.String{
																stringvalidator.LengthBetween(1, 255),
															},
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the database.",
															Optional:    true,
														},
													},
												},
												"schema": schema.SingleNestedAttribute{
													Description: "Schema identifier part.",
													Optional:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the schema.",
															Optional:    true,
															Validators: []validator.String{
																stringvalidator.LengthBetween(1, 255),
															},
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the schema.",
															Optional:    true,
														},
													},
												},
												"table": schema.SingleNestedAttribute{
													Description: "Table identifier part.",
													Optional:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the table.",
															Optional:    true,
															Validators: []validator.String{
																stringvalidator.LengthBetween(1, 255),
															},
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the table.",
															Optional:    true,
														},
													},
												},
												"column": schema.SingleNestedAttribute{
													Description: "Column identifier part.",
													Optional:    true,
													Attributes: map[string]schema.Attribute{
														"name": schema.StringAttribute{
															Description: "Name of the column.",
															Optional:    true,
															Validators: []validator.String{
																stringvalidator.LengthBetween(1, 255),
															},
														},
														"wildcard": schema.BoolAttribute{
															Description: "Wildcard for the column.",
															Optional:    true,
														},
													},
												},
											},
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

func (r *AccessManagementOLTPPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccessManagementOLTPPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccessManagementOLTPPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert rules from Terraform model to client model
	rules := convertAccessManagementOLTPRulesFromTerraform(plan.Rules)

	// Create the input for the API call
	input := client.CreateAccessManagementOLTPPolicyInput{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		DatabaseTypeName: plan.DatabaseTypeName.ValueString(),
		DatabaseType:     plan.DatabaseType.ValueInt64(),
		CaseSensitivity:  plan.CaseSensitivity.ValueString(),
		RepoName:         plan.RepoName.ValueString(),
		Rules:            rules,
	}

	// Call the API to create the access management oltp policy
	policy, err := r.client.CreateAccessManagementOLTPPolicy(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating access management oltp policy",
			"Could not create access management oltp policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AccessManagementOLTPPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccessManagementOLTPPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get access management oltp policy from API
	policy, err := r.client.GetAccessManagementOLTPPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading access management oltp policy",
			"Could not read access management oltp policy ID "+state.ID.ValueString()+": "+err.Error(),
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

func (r *AccessManagementOLTPPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AccessManagementOLTPPolicyResourceModel
	var state AccessManagementOLTPPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert rules from Terraform model to client model
	rules := convertAccessManagementOLTPRulesFromTerraform(plan.Rules)

	// Create the input for the API call
	input := client.UpdateAccessManagementOLTPPolicyInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Rules:       rules,
	}

	// Call the API to update the access management oltp policy
	policy, err := r.client.UpdateAccessManagementOLTPPolicy(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating access management oltp policy",
			"Could not update access management oltp policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AccessManagementOLTPPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccessManagementOLTPPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the access management oltp policy
	err := r.client.DeleteAccessManagementOLTPPolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting access management oltp policy",
			"Could not delete access management oltp policy, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *AccessManagementOLTPPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map API response to Terraform model
func (r *AccessManagementOLTPPolicyResource) mapPolicyToModel(policy *client.AccessManagementOLTPPolicy, model *AccessManagementOLTPPolicyResourceModel) {
	model.ID = types.StringValue(policy.ID)
	model.Name = types.StringValue(policy.Name)
	model.Description = types.StringValue(policy.Description)
	model.DatabaseTypeName = types.StringValue(policy.DatabaseTypeName)
	model.DatabaseType = types.Int64Value(policy.DatabaseType)
	model.CaseSensitivity = types.StringValue(policy.CaseSensitivity)
	model.RepoName = types.StringValue(policy.RepoName)
	model.Rules = convertAccessManagementOLTPRulesToTerraform(policy.Rules)
	model.CreatedAt = types.StringValue(policy.CreatedAt)
	model.UpdatedAt = types.StringValue(policy.UpdatedAt)
}

func convertAccessManagementOLTPRulesFromTerraform(rules types.List) []client.AccessManagementOLTPRule {
	if rules.IsNull() || rules.IsUnknown() {
		return nil
	}

	var clientRules []client.AccessManagementOLTPRule

	for _, rule := range rules.Elements() {
		ruleObj, ok := rule.(types.Object)
		if !ok {
			continue
		}
		ruleAttrs := ruleObj.Attributes()

		ruleTypeAttr, _ := ruleAttrs["type"].(types.String)
		ruleType := ruleTypeAttr.ValueString()

		actorsList, _ := ruleAttrs["actors"].(types.List)
		actors := convertAccessManagementOLTPActorsFromTerraform(actorsList)

		objectsList, _ := ruleAttrs["objects"].(types.List)
		objects := convertAccessManagementOLTPObjectsFromTerraform(objectsList)

		clientRules = append(clientRules, client.AccessManagementOLTPRule{
			Type:    ruleType,
			Actors:  actors,
			Objects: objects,
		})
	}

	return clientRules
}

func convertAccessManagementOLTPActorsFromTerraform(actors types.List) []client.AccessManagementOLTPActor {
	if actors.IsNull() || actors.IsUnknown() {
		return nil
	}

	var clientActors []client.AccessManagementOLTPActor

	for _, actor := range actors.Elements() {
		actorObj, ok := actor.(types.Object)
		if !ok {
			continue
		}
		actorAttrs := actorObj.Attributes()

		actorTypeAttr, _ := actorAttrs["type"].(types.String)
		actorType := actorTypeAttr.ValueString()

		conditionAttr, _ := actorAttrs["condition"].(types.String)
		condition := conditionAttr.ValueString()

		identifiersAttr, _ := actorAttrs["identifiers"].(types.List)
		var identifiers []string
		for _, identifier := range identifiersAttr.Elements() {
			if idStr, ok := identifier.(types.String); ok {
				identifiers = append(identifiers, idStr.ValueString())
			}
		}

		clientActors = append(clientActors, client.AccessManagementOLTPActor{
			Type:        actorType,
			Condition:   condition,
			Identifiers: identifiers,
		})
	}

	return clientActors
}

func convertAccessManagementOLTPObjectsFromTerraform(objects types.List) []client.AccessManagementOLTPObject {
	if objects.IsNull() || objects.IsUnknown() {
		return nil
	}

	var clientObjects []client.AccessManagementOLTPObject
	for _, object := range objects.Elements() {
		objectObj, ok := object.(types.Object)
		if !ok {
			continue
		}
		objectAttrs := objectObj.Attributes()

		ruleTypeAttr, _ := objectAttrs["type"].(types.String)
		ruleType := ruleTypeAttr.ValueString()

		identifiersAttr, _ := objectAttrs["identifiers"].(types.List)
		identifiers := convertAccessManagementOLTPIdentifiersFromTerraform(identifiersAttr)

		clientObjects = append(clientObjects, client.AccessManagementOLTPObject{
			Type:        ruleType,
			Identifiers: identifiers,
		})
	}

	return clientObjects
}

func convertAccessManagementOLTPIdentifiersFromTerraform(identifiers types.List) []client.AccessManagementOLTPIdentifier {
	if identifiers.IsNull() || identifiers.IsUnknown() {
		return nil
	}

	var clientIdentifiers []client.AccessManagementOLTPIdentifier

	for _, identifier := range identifiers.Elements() {
		identifierObj, ok := identifier.(types.Object)
		if !ok {
			continue
		}
		identifierAttrs := identifierObj.Attributes()

		databaseAttr, _ := identifierAttrs["database"].(types.Object)
		databaseAttrs := databaseAttr.Attributes()

		schemaAttr, _ := identifierAttrs["schema"].(types.Object)
		schemaAttrs := schemaAttr.Attributes()

		tableAttr, _ := identifierAttrs["table"].(types.Object)
		tableAttrs := tableAttr.Attributes()

		columnAttr, _ := identifierAttrs["column"].(types.Object)
		columnAttrs := columnAttr.Attributes()

		clientIdentifiers = append(clientIdentifiers, client.AccessManagementOLTPIdentifier{
			Database: client.AccessManagementOLTPIdentifierPart{
				Name:     databaseAttrs["name"].(types.String).ValueString(),
				Wildcard: databaseAttrs["wildcard"].(types.Bool).ValueBool(),
			},
			Schema: client.AccessManagementOLTPIdentifierPart{
				Name:     schemaAttrs["name"].(types.String).ValueString(),
				Wildcard: schemaAttrs["wildcard"].(types.Bool).ValueBool(),
			},
			Table: client.AccessManagementOLTPIdentifierPart{
				Name:     tableAttrs["name"].(types.String).ValueString(),
				Wildcard: tableAttrs["wildcard"].(types.Bool).ValueBool(),
			},
			Column: client.AccessManagementOLTPIdentifierPart{
				Name:     columnAttrs["name"].(types.String).ValueString(),
				Wildcard: columnAttrs["wildcard"].(types.Bool).ValueBool(),
			},
		})
	}

	return clientIdentifiers
}

func convertAccessManagementOLTPRulesToTerraform(rules []client.AccessManagementOLTPRule) types.List {
	if len(rules) == 0 {
		return types.ListNull(OLTPRuleType)
	}

	var terraformRules []attr.Value
	var diagnostics diag.Diagnostics
	for _, rule := range rules {
		// Convert actors
		var terraformActors []attr.Value
		for _, actor := range rule.Actors {
			actorValue, actorDiags := types.ObjectValue(
				OLTPActorType.AttrTypes,
				map[string]attr.Value{
					"type":        types.StringValue(actor.Type),
					"condition":   types.StringValue(actor.Condition),
					"identifiers": convertStringListToTerraform(actor.Identifiers),
				},
			)
			diagnostics.Append(actorDiags...)
			terraformActors = append(terraformActors, actorValue)
		}

		// Convert objects
		var terraformObjects []attr.Value
		for _, object := range rule.Objects {
			var identifierValues []attr.Value
			for _, identifier := range object.Identifiers {
				identifierValue, _ := types.ObjectValue(
					OLTPIdentifierType.AttrTypes,
					map[string]attr.Value{
						"database": types.ObjectValueMust(OLTPIdentifierPartType.AttrTypes, map[string]attr.Value{
							"name":     types.StringValue(identifier.Database.Name),
							"wildcard": types.BoolValue(identifier.Database.Wildcard),
						}),
						"schema": types.ObjectValueMust(OLTPIdentifierPartType.AttrTypes, map[string]attr.Value{
							"name":     types.StringValue(identifier.Schema.Name),
							"wildcard": types.BoolValue(identifier.Schema.Wildcard),
						}),
						"table": types.ObjectValueMust(OLTPIdentifierPartType.AttrTypes, map[string]attr.Value{
							"name":     types.StringValue(identifier.Table.Name),
							"wildcard": types.BoolValue(identifier.Table.Wildcard),
						}),
						"column": types.ObjectValueMust(OLTPIdentifierPartType.AttrTypes, map[string]attr.Value{
							"name":     types.StringValue(identifier.Column.Name),
							"wildcard": types.BoolValue(identifier.Column.Wildcard),
						}),
					},
				)
				identifierValues = append(identifierValues, identifierValue)
			}

			objectValue, objectDiags := types.ObjectValue(
				OLTPObjectType.AttrTypes,
				map[string]attr.Value{
					"type":        types.StringValue(object.Type),
					"identifiers": types.ListValueMust(OLTPIdentifierType, identifierValues),
				},
			)
			diagnostics.Append(objectDiags...)
			terraformObjects = append(terraformObjects, objectValue)
		}

		ruleValue, ruleDiags := types.ObjectValue(
			OLTPRuleType.AttrTypes,
			map[string]attr.Value{
				"type":    types.StringValue(rule.Type),
				"actors":  types.ListValueMust(OLTPActorType, terraformActors),
				"objects": types.ListValueMust(OLTPObjectType, terraformObjects),
			},
		)
		diagnostics.Append(ruleDiags...)
		terraformRules = append(terraformRules, ruleValue)
	}

	return types.ListValueMust(OLTPRuleType, terraformRules)
}
