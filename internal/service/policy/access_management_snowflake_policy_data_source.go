package policy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AccessManagementSnowflakePolicyDataSource{}

func NewAccessManagementSnowflakePolicyDataSource() datasource.DataSource {
	return &AccessManagementSnowflakePolicyDataSource{}
}

type AccessManagementSnowflakePolicyDataSource struct {
	client *client.Client
}

type AccessManagementSnowflakePolicyDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	ConnectionIDs     types.List   `tfsdk:"connection_ids"`
	Rules             types.List   `tfsdk:"rules"`
	PolicyMaintenance types.Object `tfsdk:"policy_maintenance"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func (d *AccessManagementSnowflakePolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_management_snowflake_policy"
}

func (d *AccessManagementSnowflakePolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving a Snowflake access management policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the Snowflake access management policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
						"must be a valid policy ID",
					),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Snowflake access management policy.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Snowflake access management policy.",
				Computed:    true,
			},
			"connection_ids": schema.ListAttribute{
				Description: "List of connection IDs associated with the policy.",
				ElementType: types.Int64Type,
				Computed:    true,
			},
			"rules": schema.ListAttribute{
				Description: "List of rules for the Snowflake access management policy.",
				ElementType: types.ObjectType{
					AttrTypes: SnowflakeRuleType.AttrTypes,
				},
				Computed: true,
			},
			"policy_maintenance": schema.SingleNestedAttribute{
				Description: "Policy maintenance configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"rate": schema.StringAttribute{
						Description: "Rate at which the policy maintenance occurs.",
						Computed:    true,
					},
					"value": schema.StringAttribute{
						Description: "Value for the policy maintenance rate.",
						Computed:    true,
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

func (d *AccessManagementSnowflakePolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccessManagementSnowflakePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccessManagementSnowflakePolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the Snowflake policy from the API
	policy, err := d.client.GetAccessManagementSnowflakePolicy(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Snowflake access management policy",
			fmt.Sprintf("Could not read Snowflake access management policy with ID %s: %s", config.ID.ValueString(), err.Error()),
		)
		return
	}

	// If the policy doesn't exist, return an error
	if policy == nil {
		resp.Diagnostics.AddError(
			"Snowflake access management policy not found",
			fmt.Sprintf("Snowflake access management policy with ID '%s' does not exist.", config.ID.ValueString()),
		)
		return
	}

	// Map the API response to the Terraform model
	config.Name = types.StringValue(policy.Name)
	config.Description = types.StringValue(policy.Description)
	//config.ConnectionIDs = convertInt64ListToTerraform(policy.ConnectionIDs)
	config.Rules = convertAccessManagementSnowflakeRulesToTerraform(policy)
	//config.PolicyMaintenance = convertPolicyMaintenanceToTerraform(policy.PolicyMaintenance)
	config.CreatedAt = types.StringValue(policy.CreatedAt)
	config.UpdatedAt = types.StringValue(policy.UpdatedAt)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper functions to map API responses to Terraform models
func convertInt64ListToTerraform(ids []int64) types.List {
	if len(ids) == 0 {
		return types.ListNull(types.Int64Type)
	}

	var terraformIDs []attr.Value
	for _, id := range ids {
		terraformIDs = append(terraformIDs, types.Int64Value(id))
	}

	return types.ListValueMust(types.Int64Type, terraformIDs)
}

func convertPolicyMaintenanceToTerraform(maintenance *client.AccessManagementPolicyMaintenance) types.Object {
	if maintenance == nil {
		return types.ObjectNull(map[string]attr.Type{
			"rate":  types.StringType,
			"value": types.StringType,
		})
	}

	return types.ObjectValueMust(map[string]attr.Type{
		"rate":  types.StringType,
		"value": types.StringType,
	}, map[string]attr.Value{
		"rate":  types.StringValue(maintenance.Rate),
		"value": types.StringValue(maintenance.Value),
	})
}
