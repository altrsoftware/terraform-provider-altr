package policy

import (
	"context"
	"fmt"
	customvalidation "terraform-provider-altr/internal/validation"

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

	"terraform-provider-altr/internal/client"
)

var _ resource.Resource = &AccessManagementSnowflakePolicyResource{}
var _ resource.ResourceWithImportState = &AccessManagementSnowflakePolicyResource{}

func NewAccessManagementSnowflakePolicyDataResource() resource.Resource {
	return &AccessManagementSnowflakePolicyResource{}
}

type AccessManagementSnowflakePolicyResource struct {
	client *client.Client
}

type AccessManagementSnowflakePolicyResourceModel struct {
	ID                types.String                              `tfsdk:"id"`
	Name              types.String                              `tfsdk:"name"`
	Description       types.String                              `tfsdk:"description"`
	ConnectionIds     []int64                                   `tfsdk:"connection_ids"`
	Rules             types.List                                `tfsdk:"rules"`
	PolicyMaintenance *client.AccessManagementPolicyMaintenance `tfsdk:"policy_maintenance"`
	CreatedAt         types.String                              `tfsdk:"created_at"`
	UpdatedAt         types.String                              `tfsdk:"updated_at"`
}

var SnowflakeActorType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"identifiers": types.ListType{
			ElemType: types.StringType,
		},
		"condition": types.StringType,
	},
}

var SnowflakeFullyQualifiedIdentifiersType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database": types.StringType,
		"schema":   types.StringType,
		"table":    types.StringType,
		"view":     types.StringType,
	},
}

var SnowflakeObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":      types.StringType,
		"condition": types.StringType,
		"identifiers": types.ListType{
			ElemType: types.StringType,
		},
		"fully_qualified_identifiers": types.ListType{
			ElemType: SnowflakeFullyQualifiedIdentifiersType, // Ensure this is a list
		},
	},
}

var SnowflakeTaggedWithType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"database": types.StringType,
		"schema":   types.StringType,
		"name":     types.StringType,
		"value":    types.StringType,
	},
}

var SnowflakeTaggedObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"check_against": types.ListType{
			ElemType: types.StringType,
		},
		"tagged_with": types.ListType{
			ElemType: SnowflakeTaggedWithType,
		},
		"tag_condition": types.StringType,
	},
}

var SnowflakeAccessType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
	},
}

var SnowflakeRuleType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"actors": types.ListType{
			ElemType: SnowflakeActorType,
		},
		"objects": types.ListType{
			ElemType: SnowflakeObjectType,
		},
		"tagged_objects": types.ListType{
			ElemType: SnowflakeTaggedObjectType,
		},
		"access": types.ListType{
			ElemType: SnowflakeAccessType,
		},
	},
}

func (d *AccessManagementSnowflakePolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_management_snowflake_policy"
}

func (d *AccessManagementSnowflakePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Snowflake access management policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the Snowflake access management policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Snowflake access management policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the Snowflake access management policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"connection_ids": schema.ListAttribute{
				Description: "List of connection IDs associated with the policy.",
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"policy_maintenance": schema.SingleNestedAttribute{
				Description: "Policy maintenance configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"rate": schema.StringAttribute{
						Description: "Rate at which the policy maintenance occurs.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("day", "cron"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"value": schema.StringAttribute{
						Description: "Value for the policy maintenance rate.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},

			"rules": schema.ListNestedAttribute{
				Description: "List of rules for the Snowflake access management policy.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
											stringvalidator.OneOf("role"),
										},
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the actor (e.g., equals, starts_with, ends_with).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("equals", "starts_with", "ends_with"),
										},
									},
									"identifiers": schema.ListAttribute{
										Description: "List of identifiers for the actor.",
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											customvalidation.UniqueStringList(),
										},
									},
								},
							},
						},
						"objects": schema.ListNestedAttribute{
							Description: "List of objects for the rule.",
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the object (e.g., database, schema, table, view).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("database", "schema", "table", "view"),
										},
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the object (e.g., equals, starts_with, ends_with, fully_qualified).",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("equals", "starts_with", "ends_with", "fully_qualified"),
										},
									},
									"identifiers": schema.ListAttribute{
										Description: "List of identifiers for the object.",
										ElementType: types.StringType,
										Optional:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											customvalidation.UniqueStringList(),
										},
									},
									"fully_qualified_identifiers": schema.ListNestedAttribute{
										Description: "List of fully qualified object reference.",
										Optional:    true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"database": schema.StringAttribute{
													Description: "Database name.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
												"schema": schema.StringAttribute{
													Description: "Schema name.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
												"table": schema.StringAttribute{
													Description: "Table name.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
												"view": schema.StringAttribute{
													Description: "View name.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 255),
													},
												},
											},
										},
									},
								},
							},
						},
						"tagged_objects": schema.ListNestedAttribute{
							Description: "Tagged objects for the rule.",
							Optional:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"check_against": schema.ListAttribute{
										Description: "Check against these objects.",
										ElementType: types.StringType,
										Optional:    true,
										Validators: []validator.List{
											listvalidator.ValueStringsAre(
												stringvalidator.OneOf("databases", "schemas", "tables", "views"),
											),
										},
									},
									"tagged_with": schema.ListNestedAttribute{
										Description: "Tagged with these object references.",
										Optional:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"database": schema.StringAttribute{
													Description: "Database name.",
													Optional:    true,
												},
												"schema": schema.StringAttribute{
													Description: "Schema name.",
													Optional:    true,
												},
												"name": schema.StringAttribute{
													Description: "Tag name.",
													Optional:    true,
												},
												"value": schema.StringAttribute{
													Description: "Tag value.",
													Optional:    true,
												},
											},
										},
									},
									"tag_condition": schema.StringAttribute{
										Description: "Tag condition for the tagged objects.",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("or", "and"),
										},
									},
								},
							},
						},
						"access": schema.ListNestedAttribute{
							Description: "Access for the rule.",
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 2),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the access permission.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("read", "write"),
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

func (r *AccessManagementSnowflakePolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccessManagementSnowflakePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccessManagementSnowflakePolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert rules from Terraform model to client model
	rules := convertAccessManagementSnowflakeRulesFromTerraform(plan.Rules)

	// Create the input for the API call
	input := client.CreateAccessManagementSnowflakePolicyInput{
		Name:          plan.Name.ValueString(),
		Description:   plan.Description.ValueString(),
		ConnectionIds: plan.ConnectionIds,
		Rules:         rules,
	}

	if plan.PolicyMaintenance != nil {
		input.PolicyMaintenance = plan.PolicyMaintenance
	} else {
		input.PolicyMaintenance = nil
	}

	// Call the API to create the access management snowflake policy
	policy, err := r.client.CreateAccessManagementSnowflakePolicy(input)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating access management snowflake policy",
			"Could not create access management snowflake policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AccessManagementSnowflakePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccessManagementSnowflakePolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get access management snowflake policy from API
	policy, err := r.client.GetAccessManagementSnowflakePolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading access management snowflake policy",
			"Could not read access management snowflake policy ID "+state.ID.ValueString()+": "+err.Error(),
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

func (r *AccessManagementSnowflakePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AccessManagementSnowflakePolicyResourceModel
	var state AccessManagementSnowflakePolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert rules from Terraform model to client model
	rules := convertAccessManagementSnowflakeRulesFromTerraform(plan.Rules)

	// Create the input for the API call
	input := client.UpdateAccessManagementSnowflakePolicyInput{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Rules:       rules,
	}

	// Call the API to update the access management snowflake policy
	policy, err := r.client.UpdateAccessManagementSnowflakePolicy(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating access management snowflake policy",
			"Could not update access management snowflake policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to the model
	r.mapPolicyToModel(policy, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AccessManagementSnowflakePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccessManagementSnowflakePolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the access management snowflake policy
	err := r.client.DeleteAccessManagementSnowflakePolicy(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting access management snowflake policy",
			"Could not delete access management snowflake policy, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *AccessManagementSnowflakePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map API response to Terraform model
func (r *AccessManagementSnowflakePolicyResource) mapPolicyToModel(policy *client.AccessManagementSnowflakePolicy, model *AccessManagementSnowflakePolicyResourceModel) {
	model.ID = types.StringValue(policy.ID)
	model.Name = types.StringValue(policy.Name)
	model.Description = types.StringValue(policy.Description)
	model.Rules = convertAccessManagementSnowflakeRulesToTerraform(policy)
	model.CreatedAt = types.StringValue(policy.CreatedAt)
	model.UpdatedAt = types.StringValue(policy.UpdatedAt)
}

func convertAccessManagementSnowflakeRulesFromTerraform(rules types.List) []client.AccessManagementSnowflakeRule {
	if rules.IsNull() || rules.IsUnknown() {
		return nil
	}

	var clientRules []client.AccessManagementSnowflakeRule

	for _, rule := range rules.Elements() {
		ruleObj, ok := rule.(types.Object)
		if !ok {
			continue
		}
		ruleAttrs := ruleObj.Attributes()

		actorsList, _ := ruleAttrs["actors"].(types.List)
		actors := convertAccessManagementSnowflakeActorsFromTerraform(actorsList)

		objectsList, _ := ruleAttrs["objects"].(types.List)
		objects := convertAccessManagementSnowflakeObjectsFromTerraform(objectsList)

		taggedObjectsList, _ := ruleAttrs["tagged_objects"].(types.List)
		taggedObjects := convertAccessManagementSnowflakeTaggedObjectsFromTerraform(taggedObjectsList)

		accessList, _ := ruleAttrs["access"].(types.List)
		access := convertAccessManagementSnowflakeAccessFromTerraform(accessList)

		clientRules = append(clientRules, client.AccessManagementSnowflakeRule{
			Actors:        actors,
			Objects:       objects,
			TaggedObjects: taggedObjects,
			Access:        access,
		})
	}

	return clientRules
}

func convertAccessManagementSnowflakeActorsFromTerraform(actors types.List) []client.AccessManagementSnowflakeActor {
	if actors.IsNull() || actors.IsUnknown() {
		return nil
	}

	var clientActors []client.AccessManagementSnowflakeActor

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

		clientActors = append(clientActors, client.AccessManagementSnowflakeActor{
			Type:        actorType,
			Condition:   condition,
			Identifiers: identifiers,
		})
	}

	return clientActors
}

func convertAccessManagementSnowflakeObjectsToTerraform(objects []client.AccessManagementSnowflakeObject) []attr.Value {
	var terraformObjects []attr.Value
	for _, object := range objects {
		var terraformFQIdentifiers []attr.Value
		if len(object.FullyQualifiedIdentifiers) > 0 {
			for _, fqIdentifier := range object.FullyQualifiedIdentifiers {
				fqValue, _ := types.ObjectValue(
					SnowflakeFullyQualifiedIdentifiersType.AttrTypes,
					map[string]attr.Value{
						"database": types.StringValue(fqIdentifier.Database),
						"schema":   types.StringValue(fqIdentifier.Schema),
						"table":    types.StringValue(fqIdentifier.Table),
						"view":     types.StringValue(fqIdentifier.View),
					},
				)
				terraformFQIdentifiers = append(terraformFQIdentifiers, fqValue)
			}
		}

		// Set `fully_qualified_identifiers` to null if empty
		var fqIdentifiersList attr.Value
		if len(terraformFQIdentifiers) == 0 {
			fqIdentifiersList = types.ListNull(SnowflakeFullyQualifiedIdentifiersType)
		} else {
			fqIdentifiersList = types.ListValueMust(SnowflakeFullyQualifiedIdentifiersType, terraformFQIdentifiers)
		}

		objectValue, _ := types.ObjectValue(
			SnowflakeObjectType.AttrTypes,
			map[string]attr.Value{
				"type":                        types.StringValue(object.Type),
				"condition":                   types.StringValue(object.Condition),
				"identifiers":                 convertStringListToTerraform(object.Identifiers),
				"fully_qualified_identifiers": fqIdentifiersList,
			},
		)
		terraformObjects = append(terraformObjects, objectValue)
	}
	return terraformObjects
}

func convertAccessManagementSnowflakeObjectsFromTerraform(objects types.List) []client.AccessManagementSnowflakeObject {
	if objects.IsNull() || objects.IsUnknown() {
		return nil
	}

	var clientObjects []client.AccessManagementSnowflakeObject
	for _, object := range objects.Elements() {
		objectObj, ok := object.(types.Object)
		if !ok {
			continue
		}
		objectAttrs := objectObj.Attributes()

		ruleTypeAttr, _ := objectAttrs["type"].(types.String)
		ruleType := ruleTypeAttr.ValueString()

		conditionAttr, _ := objectAttrs["condition"].(types.String)
		condition := conditionAttr.ValueString()

		identifiersAttr, _ := objectAttrs["identifiers"].(types.List)
		var identifiers []string
		for _, identifier := range identifiersAttr.Elements() {
			if idStr, ok := identifier.(types.String); ok {
				identifiers = append(identifiers, idStr.ValueString())
			}
		}

		var fullyQualifiedIdentifiers []client.AccessManagementSnowflakeFullyQualifiedIdentifiers
		fqIdentifiersAttr, _ := objectAttrs["fully_qualified_identifiers"].(types.List)
		for _, fqIdentifier := range fqIdentifiersAttr.Elements() {
			fqObj, ok := fqIdentifier.(types.Object)
			if !ok {
				continue
			}
			fqAttrs := fqObj.Attributes()

			fullyQualifiedIdentifiers = append(fullyQualifiedIdentifiers, client.AccessManagementSnowflakeFullyQualifiedIdentifiers{
				Database: fqAttrs["database"].(types.String).ValueString(),
				Schema:   fqAttrs["schema"].(types.String).ValueString(),
				Table:    fqAttrs["table"].(types.String).ValueString(),
				View:     fqAttrs["view"].(types.String).ValueString(),
			})
		}

		clientObjects = append(clientObjects, client.AccessManagementSnowflakeObject{
			Type:                      ruleType,
			Condition:                 condition,
			Identifiers:               identifiers,
			FullyQualifiedIdentifiers: fullyQualifiedIdentifiers,
		})
	}

	return clientObjects
}

func convertAccessManagementSnowflakeTaggedObjectsFromTerraform(taggedObjects types.List) []client.AccessManagementSnowflakeTaggedObject {
	if taggedObjects.IsNull() || taggedObjects.IsUnknown() {
		return nil
	}

	var clientTaggedObjects []client.AccessManagementSnowflakeTaggedObject

	for _, taggedObject := range taggedObjects.Elements() {
		taggedObjectObj, ok := taggedObject.(types.Object)
		if !ok {
			continue
		}
		taggedObjectAttrs := taggedObjectObj.Attributes()

		checkAgainstAttr, _ := taggedObjectAttrs["check_against"].(types.List)
		var checkAgainst []string
		for _, check := range checkAgainstAttr.Elements() {
			if checkStr, ok := check.(types.String); ok {
				checkAgainst = append(checkAgainst, checkStr.ValueString())
			}
		}

		taggedWithAttr, _ := taggedObjectAttrs["tagged_with"].(types.List)
		taggedWith := convertAccessManagementSnowflakeTaggedWithFromTerraform(taggedWithAttr)

		tagConditionAttr, _ := taggedObjectAttrs["tag_condition"].(types.String)
		tagCondition := tagConditionAttr.ValueString()

		clientTaggedObjects = append(clientTaggedObjects, client.AccessManagementSnowflakeTaggedObject{
			CheckAgainst: checkAgainst,
			TaggedWith:   taggedWith,
			TagCondition: tagCondition,
		})
	}

	return clientTaggedObjects
}

func convertAccessManagementSnowflakeTaggedWithFromTerraform(taggedWith types.List) []client.AccessManagementSnowflakeTaggedWith {
	if taggedWith.IsNull() || taggedWith.IsUnknown() {
		return nil
	}

	var clientTaggedWith []client.AccessManagementSnowflakeTaggedWith

	for _, tag := range taggedWith.Elements() {
		tagObj, ok := tag.(types.Object)
		if !ok {
			continue
		}
		tagAttrs := tagObj.Attributes()

		databaseAttr, _ := tagAttrs["database"].(types.String)
		schemaAttr, _ := tagAttrs["schema"].(types.String)
		nameAttr, _ := tagAttrs["name"].(types.String)
		valueAttr, _ := tagAttrs["value"].(types.String)

		clientTaggedWith = append(clientTaggedWith, client.AccessManagementSnowflakeTaggedWith{
			Database: databaseAttr.ValueString(),
			Schema:   schemaAttr.ValueString(),
			Name:     nameAttr.ValueString(),
			Value:    valueAttr.ValueString(),
		})
	}

	return clientTaggedWith
}

func convertAccessManagementSnowflakeAccessFromTerraform(access types.List) []client.AccessManagementSnowflakeAccess {
	if access.IsNull() || access.IsUnknown() {
		return nil
	}

	var clientAccess []client.AccessManagementSnowflakeAccess

	for _, acc := range access.Elements() {
		accObj, ok := acc.(types.Object)
		if !ok {
			continue
		}
		accAttrs := accObj.Attributes()

		nameAttr, _ := accAttrs["name"].(types.String)
		name := nameAttr.ValueString()

		clientAccess = append(clientAccess, client.AccessManagementSnowflakeAccess{
			Name: name,
		})
	}

	return clientAccess
}

func convertAccessManagementSnowflakeRulesToTerraform(policy *client.AccessManagementSnowflakePolicy) types.List {
	var rules []client.AccessManagementSnowflakeRule
	if policy.PendingRules != nil {
		rules = policy.PendingRules
	} else if policy.FailedRules != nil {
		rules = policy.FailedRules
	} else {
		rules = policy.AppliedRules
	}

	if len(rules) == 0 {
		return types.ListNull(SnowflakeRuleType)
	}

	var terraformRules []attr.Value
	var diagnostics diag.Diagnostics
	for _, rule := range rules {
		// Convert actors
		var terraformActors []attr.Value
		for _, actor := range rule.Actors {
			actorValue, actorDiags := types.ObjectValue(
				SnowflakeActorType.AttrTypes,
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
		if rule.Objects != nil && len(rule.Objects) > 0 {
			for _, object := range rule.Objects {
				objectValue, objectDiags := types.ObjectValue(
					SnowflakeObjectType.AttrTypes,
					map[string]attr.Value{
						"type":                        types.StringValue(object.Type),
						"condition":                   types.StringValue(object.Condition),
						"identifiers":                 convertStringListToTerraform(object.Identifiers),
						"fully_qualified_identifiers": convertFullyQualifiedIdentifiersToTerraform(object.FullyQualifiedIdentifiers),
					},
				)
				diagnostics.Append(objectDiags...)
				terraformObjects = append(terraformObjects, objectValue)
			}
		}

		// Set `objects` to null if empty
		var objectsList attr.Value
		if len(terraformObjects) == 0 {
			objectsList = types.ListNull(SnowflakeObjectType)
		} else {
			objectsList = types.ListValueMust(SnowflakeObjectType, terraformObjects)
		}

		// Convert tagged objects
		var terraformTaggedObjects []attr.Value
		if rule.TaggedObjects != nil && len(rule.TaggedObjects) > 0 {
			for _, taggedObject := range rule.TaggedObjects {
				var terraformTaggedWith []attr.Value
				for _, tag := range taggedObject.TaggedWith {
					tagValue, tagDiags := types.ObjectValue(
						SnowflakeTaggedWithType.AttrTypes,
						map[string]attr.Value{
							"database": types.StringValue(tag.Database),
							"schema":   types.StringValue(tag.Schema),
							"name":     types.StringValue(tag.Name),
							"value":    types.StringValue(tag.Value),
						},
					)
					diagnostics.Append(tagDiags...)
					terraformTaggedWith = append(terraformTaggedWith, tagValue)
				}

				taggedObjectValue, taggedObjectDiags := types.ObjectValue(
					SnowflakeTaggedObjectType.AttrTypes,
					map[string]attr.Value{
						"check_against": convertStringListToTerraform(taggedObject.CheckAgainst),
						"tagged_with":   types.ListValueMust(SnowflakeTaggedWithType, terraformTaggedWith),
						"tag_condition": types.StringValue(taggedObject.TagCondition),
					},
				)
				diagnostics.Append(taggedObjectDiags...)
				terraformTaggedObjects = append(terraformTaggedObjects, taggedObjectValue)
			}
		}

		// Set `tagged_objects` to null if empty
		var taggedObjectsList attr.Value
		if len(terraformTaggedObjects) == 0 {
			taggedObjectsList = types.ListNull(SnowflakeTaggedObjectType)
		} else {
			taggedObjectsList = types.ListValueMust(SnowflakeTaggedObjectType, terraformTaggedObjects)
		}

		// Convert access
		var terraformAccess []attr.Value
		for _, access := range rule.Access {
			accessValue, accessDiags := types.ObjectValue(
				SnowflakeAccessType.AttrTypes,
				map[string]attr.Value{
					"name": types.StringValue(access.Name),
				},
			)
			diagnostics.Append(accessDiags...)
			terraformAccess = append(terraformAccess, accessValue)
		}

		ruleValue, ruleDiags := types.ObjectValue(
			SnowflakeRuleType.AttrTypes,
			map[string]attr.Value{
				"actors":         types.ListValueMust(SnowflakeActorType, terraformActors),
				"objects":        objectsList,
				"tagged_objects": taggedObjectsList,
				"access":         types.ListValueMust(SnowflakeAccessType, terraformAccess),
			},
		)
		diagnostics.Append(ruleDiags...)
		terraformRules = append(terraformRules, ruleValue)
	}

	if diagnostics.HasError() {
		return types.ListNull(SnowflakeRuleType)
	}

	return types.ListValueMust(SnowflakeRuleType, terraformRules)
}

func convertFullyQualifiedIdentifiersToTerraform(fqIdentifiers []client.AccessManagementSnowflakeFullyQualifiedIdentifiers) attr.Value {
	if len(fqIdentifiers) == 0 {
		return types.ListNull(SnowflakeFullyQualifiedIdentifiersType)
	}

	var terraformFQIdentifiers []attr.Value
	for _, fqIdentifier := range fqIdentifiers {
		// Convert each fully qualified identifier to a Terraform object
		fqValue, _ := types.ObjectValue(
			SnowflakeFullyQualifiedIdentifiersType.AttrTypes,
			map[string]attr.Value{
				"database": convertStringToTerraformValue(fqIdentifier.Database),
				"schema":   convertStringToTerraformValue(fqIdentifier.Schema),
				"table":    convertStringToTerraformValue(fqIdentifier.Table),
				"view":     convertStringToTerraformValue(fqIdentifier.View),
			},
		)
		terraformFQIdentifiers = append(terraformFQIdentifiers, fqValue)
	}

	// Return the list of fully qualified identifiers
	return types.ListValueMust(SnowflakeFullyQualifiedIdentifiersType, terraformFQIdentifiers)
}

// Helper function to convert a string to a Terraform value
func convertStringToTerraformValue(value string) attr.Value {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
