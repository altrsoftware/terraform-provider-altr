// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package sidecar

import (
	"context"
	"fmt"
	"regexp"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/altrsoftware/terraform-provider-altr/internal/service"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SidecarListenerDataSource{}

func NewSidecarListenerDataSource() datasource.DataSource {
	return &SidecarListenerDataSource{}
}

type SidecarListenerDataSource struct {
	client *client.Client
}

type SidecarListenerDataSourceModel struct {
	SidecarID         types.String `tfsdk:"sidecar_id"`
	Port              types.Int64  `tfsdk:"port"`
	DatabaseType      types.String `tfsdk:"database_type"`
	AdvertisedVersion types.String `tfsdk:"advertised_version"`
}

func (d *SidecarListenerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sidecar_listener"
}

func (d *SidecarListenerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving sidecar listener port information.",
		Attributes: map[string]schema.Attribute{
			"sidecar_id": schema.StringAttribute{
				Description: "ID of the sidecar.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port number of the listener.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"database_type": schema.StringAttribute{
				Description: "Type of database (e.g., Oracle, etc.).",
				Computed:    true,
			},
			"advertised_version": schema.StringAttribute{
				Description: "Advertised version of the database.",
				Computed:    true,
			},
		},
	}
}

func (d *SidecarListenerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SidecarListenerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SidecarListenerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get sidecar listener from API
	listener, err := d.client.GetSidecarListener(config.SidecarID.ValueString(), int(config.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sidecar listener",
			fmt.Sprintf("Could not read sidecar listener for sidecar %s on port %d: %s",
				config.SidecarID.ValueString(),
				config.Port.ValueInt64(),
				err.Error()),
		)

		return
	}

	// If listener doesn't exist, return error
	if listener == nil {
		resp.Diagnostics.AddError(
			"Sidecar listener not found",
			fmt.Sprintf("Sidecar listener for sidecar '%s' on port %d does not exist.",
				config.SidecarID.ValueString(),
				config.Port.ValueInt64()),
		)

		return
	}

	// Map response to the model
	d.mapListenerToModel(listener, &config)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// Helper function to map API response to Terraform model
func (d *SidecarListenerDataSource) mapListenerToModel(listener *client.ListenerPort, model *SidecarListenerDataSourceModel) {
	model.Port = types.Int64Value(int64(listener.Port))
	model.DatabaseType = types.StringValue(listener.DatabaseType)

	if listener.AdvertisedVersion != "" {
		model.AdvertisedVersion = types.StringValue(listener.AdvertisedVersion)
	} else {
		model.AdvertisedVersion = types.StringNull()
	}
}
