// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"
	"regexp"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/altrsoftware/terraform-provider-altr/internal/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AgentDataSource{}

func NewAgentDataSource() datasource.DataSource {
	return &AgentDataSource{}
}

type AgentDataSource struct {
	client *client.Client
}

type AgentDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DataPlaneURL types.String `tfsdk:"data_plane_url"`
	PublicKey1   types.String `tfsdk:"public_key_1"`
	PublicKey2   types.String `tfsdk:"public_key_2"`
	TaskCount    types.Int64  `tfsdk:"task_count"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (d *AgentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (d *AgentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving information about an ALTR agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Agent UUID.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			"type": schema.StringAttribute{
				Description: "Agent type (e.g. CLASSIFIER).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the agent.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the agent.",
				Computed:    true,
			},
			"data_plane_url": schema.StringAttribute{
				Description: "URL of the data plane this agent connects to.",
				Computed:    true,
			},
			"public_key_1": schema.StringAttribute{
				Description: "First PEM-encoded RSA public key registered for the agent.",
				Computed:    true,
			},
			"public_key_2": schema.StringAttribute{
				Description: "Second PEM-encoded RSA public key registered for the agent.",
				Computed:    true,
			},
			"task_count": schema.Int64Attribute{
				Description: "Number of tasks currently assigned to this agent.",
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

func (d *AgentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = c
}

func (d *AgentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AgentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	a, err := d.client.GetAgent(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading agent",
			"Could not read agent "+config.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	if a == nil {
		resp.Diagnostics.AddError(
			"Agent not found",
			"Agent with ID '"+config.ID.ValueString()+"' does not exist.",
		)

		return
	}

	d.mapAgentToModel(a, &config)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func (d *AgentDataSource) mapAgentToModel(a *client.Agent, model *AgentDataSourceModel) {
	model.ID = types.StringValue(a.ID)
	model.Type = types.StringValue(a.Type)
	model.Name = types.StringValue(a.Name)
	model.Description = types.StringValue(a.Description)
	model.DataPlaneURL = types.StringValue(a.DataPlaneURL)
	model.TaskCount = types.Int64Value(int64(a.TaskCount))
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
