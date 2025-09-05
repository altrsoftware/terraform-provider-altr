package sidecar

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-altr/internal/client"
	"terraform-provider-altr/internal/service"
)

var _ resource.Resource = &SidecarListenerResource{}
var _ resource.ResourceWithImportState = &SidecarListenerResource{}

func NewSidecarListenerResource() resource.Resource {
	return &SidecarListenerResource{}
}

type SidecarListenerResource struct {
	client *client.Client
}

type SidecarListenerResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SidecarID         types.String `tfsdk:"sidecar_id"`
	Port              types.Int64  `tfsdk:"port"`
	DatabaseType      types.String `tfsdk:"database_type"`
	AdvertisedVersion types.String `tfsdk:"advertised_version"`
}

func (r *SidecarListenerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sidecar_listener"
}

func (r *SidecarListenerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a sidecar listener port.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the sidecar listener (sidecar_id:port).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sidecar_id": schema.StringAttribute{
				Description: "ID of the sidecar.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			// 1 to 65535
			"port": schema.Int64Attribute{
				Description: "Port number for the listener.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"database_type": schema.StringAttribute{
				Description: "Type of database (e.g., Oracle, etc.).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(service.OltpDatabaseTypes...),
				},
			},
			"advertised_version": schema.StringAttribute{
				Description: "Advertised version of the database.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
		},
	}
}

func (r *SidecarListenerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SidecarListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SidecarListenerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the input for the API call
	input := client.RegisterSidecarListenerInput{
		Port:         int(plan.Port.ValueInt64()),
		DatabaseType: plan.DatabaseType.ValueString(),
	}

	// Set optional fields
	if !plan.AdvertisedVersion.IsNull() {
		input.AdvertisedVersion = plan.AdvertisedVersion.ValueString()
	}

	// Call the API to register the sidecar listener
	err := r.client.RegisterSidecarListener(plan.SidecarID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error registering sidecar listener",
			"Could not register sidecar listener, unexpected error: "+err.Error(),
		)
		return
	}

	// Set the ID
	plan.ID = types.StringValue(fmt.Sprintf("%s:%d", plan.SidecarID.ValueString(), plan.Port.ValueInt64()))

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SidecarListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SidecarListenerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get sidecar listener from API
	listener, err := r.client.GetSidecarListener(state.SidecarID.ValueString(), int(state.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sidecar listener",
			"Could not read sidecar listener for sidecar "+state.SidecarID.ValueString()+" on port "+strconv.FormatInt(state.Port.ValueInt64(), 10)+": "+err.Error(),
		)
		return
	}

	// If listener doesn't exist, remove it from state
	if listener == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to the model
	r.mapListenerToModel(listener, &state)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SidecarListenerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// According to the API specification, sidecar listeners cannot be updated
	// They can only be registered (created) and deregistered (deleted)
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Sidecar listeners cannot be updated. Please recreate the resource to make changes.",
	)
}

func (r *SidecarListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SidecarListenerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Deregister the sidecar listener
	err := r.client.DeregisterSidecarListener(state.SidecarID.ValueString(), int(state.Port.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deregistering sidecar listener",
			"Could not deregister sidecar listener, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *SidecarListenerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import ID format: "sidecar_id:port"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: sidecar_id:port",
		)
		return
	}

	sidecarID := parts[0]
	portStr := parts[1]

	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Port in Import ID",
			"Port must be a valid integer: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sidecar_id"), sidecarID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), port)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Helper function to map API response to Terraform model
func (r *SidecarListenerResource) mapListenerToModel(listener *client.ListenerPort, model *SidecarListenerResourceModel) {
	model.Port = types.Int64Value(int64(listener.Port))
	model.DatabaseType = types.StringValue(listener.DatabaseType)

	if listener.AdvertisedVersion != "" {
		model.AdvertisedVersion = types.StringValue(listener.AdvertisedVersion)
	} else {
		model.AdvertisedVersion = types.StringNull()
	}

	// Set the ID
	model.ID = types.StringValue(fmt.Sprintf("%s:%d", model.SidecarID.ValueString(), listener.Port))
}
