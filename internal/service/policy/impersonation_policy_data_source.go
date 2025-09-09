package policy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-altr/internal/client"
)

var _ datasource.DataSource = &ImpersonationPolicyDataSource{}

func NewImpersonationPolicyDataSource() datasource.DataSource {
	return &ImpersonationPolicyDataSource{}
}

type ImpersonationPolicyDataSource struct {
	client *client.Client
}

type ImpersonationPolicyDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	RepoName    types.String `tfsdk:"repo_name"`
	Rules       types.List   `tfsdk:"rules"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *ImpersonationPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_impersonation_policy"
}

func (d *ImpersonationPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving an impersonation policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the impersonation policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
						"must be a valid policy ID",
					),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the impersonation policy.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the impersonation policy.",
				Computed:    true,
			},
			"repo_name": schema.StringAttribute{
				Description: "The name of the repository where the impersonation policy is defined.",
				Computed:    true,
			},
			"rules": schema.ListNestedAttribute{
				Description: "List of rules for the impersonation policy.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"actors": schema.ListNestedAttribute{
							Description: "List of actors for the rule.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the actor (e.g., idp_user, idp_group).",
										Computed:    true,
									},
									"identifiers": schema.ListAttribute{
										Description: "List of user or group identifiers.",
										ElementType: types.StringType,
										Computed:    true,
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the actor (e.g., equals).",
										Computed:    true,
									},
								},
							},
						},
						"targets": schema.ListNestedAttribute{
							Description: "List of target users or groups.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the target (e.g., repo_user).",
										Computed:    true,
									},
									"identifiers": schema.ListAttribute{
										Description: "List of target identifiers.",
										ElementType: types.StringType,
										Computed:    true,
									},
									"condition": schema.StringAttribute{
										Description: "Condition for the target (e.g., equals).",
										Computed:    true,
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

func (d *ImpersonationPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ImpersonationPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ImpersonationPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the impersonation policy from the API
	policy, err := d.client.GetImpersonationPolicy(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading impersonation policy",
			fmt.Sprintf("Could not read impersonation policy with ID %s: %s", config.ID.ValueString(), err.Error()),
		)
		return
	}

	// If the policy doesn't exist, return an error
	if policy == nil {
		resp.Diagnostics.AddError(
			"Impersonation policy not found",
			fmt.Sprintf("Impersonation policy with ID '%s' does not exist.", config.ID.ValueString()),
		)
		return
	}

	// Map the API response to the Terraform model
	config.Name = types.StringValue(policy.Name)
	config.Description = types.StringValue(policy.Description)
	config.RepoName = types.StringValue(policy.RepoName)
	config.Rules = convertImpersonationRulesToTerraform(policy.Rules)
	config.CreatedAt = types.StringValue(policy.CreatedAt)
	config.UpdatedAt = types.StringValue(policy.UpdatedAt)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to convert rules from the API response to Terraform model
func convertImpersonationRulesToTerraform(rules []client.ImpersonationRule) types.List {
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

	var terraformRules []attr.Value
	for _, rule := range rules {
		// Convert actors
		var terraformActors []attr.Value
		for _, actor := range rule.Actors {
			actorValue, _ := types.ObjectValue(
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
			terraformActors = append(terraformActors, actorValue)
		}

		// Convert targets
		var terraformTargets []attr.Value
		for _, target := range rule.Targets {
			targetValue, _ := types.ObjectValue(
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
			terraformTargets = append(terraformTargets, targetValue)
		}

		// Append the rule to the Terraform rules
		ruleValue, _ := types.ObjectValue(
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
		terraformRules = append(terraformRules, ruleValue)
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
