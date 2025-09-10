// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package sidecar

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

var _ datasource.DataSource = &SidecarDataSource{}

func NewSidecarDataSource() datasource.DataSource {
	return &SidecarDataSource{}
}

type SidecarDataSource struct {
	client *client.Client
}

type SidecarDataSourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Hostname                 types.String `tfsdk:"hostname"`
	OrgID                    types.String `tfsdk:"org_id"`
	DataPlaneURL             types.String `tfsdk:"data_plane_url"`
	ListenerCount            types.Int64  `tfsdk:"listener_count"`
	ListenerRepoBindingCount types.Int64  `tfsdk:"listener_repo_binding_count"`
	PublicKey1               types.String `tfsdk:"public_key_1"`
	PublicKey2               types.String `tfsdk:"public_key_2"`
	UnsupportedQueryBypass   types.Bool   `tfsdk:"unsupported_query_bypass"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
}

func (d *SidecarDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sidecar"
}

func (d *SidecarDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving sidecar information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the sidecar.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the sidecar.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the sidecar.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the sidecar.",
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "Organization ID that owns this sidecar.",
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
			"public_key_1": schema.StringAttribute{
				Description: "First public key for the sidecar.",
				Computed:    true,
			},
			"public_key_2": schema.StringAttribute{
				Description: "Second public key for the sidecar.",
				Computed:    true,
			},
			"unsupported_query_bypass": schema.BoolAttribute{
				Description: "If true, the sidecar will bypass the query parser and return all results without applying policy.",
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

func (d *SidecarDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SidecarDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SidecarDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get sidecar from API
	sidecar, err := d.client.GetSidecar(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sidecar",
			"Could not read sidecar "+config.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	// If sidecar doesn't exist, return error
	if sidecar == nil {
		resp.Diagnostics.AddError(
			"Sidecar not found",
			"Sidecar with ID '"+config.ID.ValueString()+"' does not exist.",
		)

		return
	}

	// Map response to the model
	d.mapSidecarToModel(sidecar, &config)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to map API response to Terraform model
func (d *SidecarDataSource) mapSidecarToModel(sidecar *client.Sidecar, model *SidecarDataSourceModel) {
	model.ID = types.StringValue(sidecar.ID)
	model.Name = types.StringValue(sidecar.Name)
	model.Description = types.StringValue(sidecar.Description)
	model.Hostname = types.StringValue(sidecar.Hostname)
	model.OrgID = types.StringValue(sidecar.OrgID)
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
