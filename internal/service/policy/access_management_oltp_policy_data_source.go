package policy

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-altr/internal/client"
)

var _ datasource.DataSource = &AccessManagementOLTPPolicyDataSource{}

func NewAccessManagementOLTPPolicyDataSource() datasource.DataSource {
	return &AccessManagementOLTPPolicyDataSource{}
}

type AccessManagementOLTPPolicyDataSource struct {
	client *client.Client
}

type AccessManagementOLTPPolicyDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	RepoName         types.String `tfsdk:"repo_name"`
	CaseSensitivity  types.String `tfsdk:"case_sensitivity"`
	DatabaseType     types.Int64  `tfsdk:"database_type"`
	DatabaseTypeName types.String `tfsdk:"database_type_name"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (d *AccessManagementOLTPPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_management_oltp_policy"
}

func (d *AccessManagementOLTPPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving an OLTP access management policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the OLTP access management policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
						"must be a valid policy ID",
					),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the OLTP access management policy.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the OLTP access management policy.",
				Computed:    true,
			},
			"repo_name": schema.StringAttribute{
				Description: "The name of the repository this policy belongs to.",
				Computed:    true,
			},
			"case_sensitivity": schema.StringAttribute{
				Description: "Case sensitivity for the policy.",
				Computed:    true,
			},
			"database_type": schema.StringAttribute{
				Description: "Database type ID for the policy.",
				Computed:    true,
			},
			"database_type_name": schema.StringAttribute{
				Description: "Database type name for the policy.",
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

func (d *AccessManagementOLTPPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccessManagementOLTPPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccessManagementOLTPPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the OLTP policy from the API
	policy, err := d.client.GetAccessManagementOLTPPolicy(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading OLTP access management policy",
			fmt.Sprintf("Could not read OLTP access management policy with ID %s: %s", config.ID.ValueString(), err.Error()),
		)
		return
	}

	// If the policy doesn't exist, return an error
	if policy == nil {
		resp.Diagnostics.AddError(
			"OLTP access management policy not found",
			fmt.Sprintf("OLTP access management policy with ID '%s' does not exist.", config.ID.ValueString()),
		)
		return
	}

	// Map the API response to the Terraform model
	config.Name = types.StringValue(policy.Name)
	config.Description = types.StringValue(policy.Description)
	config.RepoName = types.StringValue(policy.RepoName)
	config.CaseSensitivity = types.StringValue(policy.CaseSensitivity)
	config.DatabaseType = types.Int64Value(policy.DatabaseType)
	config.DatabaseTypeName = types.StringValue(policy.DatabaseTypeName)
	config.CreatedAt = types.StringValue(policy.CreatedAt)
	config.UpdatedAt = types.StringValue(policy.UpdatedAt)

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
