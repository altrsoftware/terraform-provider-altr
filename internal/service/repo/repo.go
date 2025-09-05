package repo

import (
	"context"
	"fmt"
	"regexp"

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

var _ resource.Resource = &RepoResource{}
var _ resource.ResourceWithImportState = &RepoResource{}

func NewRepoResource() resource.Resource {
	return &RepoResource{}
}

type RepoResource struct {
	client *client.Client
}

type RepoResourceModel struct {
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	Hostname     types.String `tfsdk:"hostname"`
	Port         types.Int64  `tfsdk:"port"`
	UserCount    types.Int64  `tfsdk:"user_count"`
	BindingCount types.Int64  `tfsdk:"binding_count"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *RepoResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repo"
}

func (r *RepoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a repository.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the repository.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.AlphanumericAndUnderscoreRegex),
						"must be alphanumeric or underscore",
					),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the repository.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(100),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the repository (e.g., Oracle).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(service.OltpDatabaseTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname of the repository.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.HostnameRegexStringRFC1123),
						"must be a valid hostname",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port number of the repository.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"user_count": schema.Int64Attribute{
				Description: "Number of users associated with this repository.",
				Computed:    true,
			},
			"binding_count": schema.Int64Attribute{
				Description: "Number of sidecar bindings for this repository.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp.",
				Computed:    true,
			},
		},
	}
}

func (r *RepoResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RepoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepoResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the input for the API call
	input := client.CreateRepoInput{
		Name:        plan.Name.ValueString(),
		Type:        plan.Type.ValueString(),
		Hostname:    plan.Hostname.ValueString(),
		Port:        int(plan.Port.ValueInt64()),
		Description: plan.Description.ValueString(),
	}

	// Call the API to create the repo
	repo, err := r.client.CreateRepo(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			"Could not create repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to the model
	r.mapRepoToModel(repo, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepoResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get repo from API
	repo, err := r.client.GetRepo(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading repository",
			"Could not read repository "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// If repo doesn't exist, remove it from state
	if repo == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to the model
	r.mapRepoToModel(repo, &state)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *RepoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RepoResourceModel
	var state RepoResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the input for the API call
	input := client.UpdateRepoInput{}

	// Only description can be updated according to the API spec
	if !plan.Description.Equal(state.Description) {
		input.Description = plan.Description.ValueString()
	}

	// Call the API to update the repo
	repo, err := r.client.UpdateRepo(state.Name.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating repository",
			"Could not update repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to the model
	r.mapRepoToModel(repo, &plan)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *RepoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepoResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the repo
	err := r.client.DeleteRepo(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting repository",
			"Could not delete repository. Note: A repository cannot be deleted if it has users or sidecar bindings. Error: "+err.Error(),
		)
		return
	}
}

func (r *RepoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to name attribute (repos are identified by name)
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// Helper function to map API response to Terraform model
func (r *RepoResource) mapRepoToModel(repo *client.Repo, model *RepoResourceModel) {
	model.Name = types.StringValue(repo.Name)
	model.Description = types.StringValue(repo.Description)
	model.Type = types.StringValue(repo.Type)
	model.Hostname = types.StringValue(repo.Hostname)
	model.Port = types.Int64Value(int64(repo.Port))
	model.UserCount = types.Int64Value(int64(repo.UserCount))
	model.BindingCount = types.Int64Value(int64(repo.BindingCount))
	model.CreatedAt = types.StringValue(repo.CreatedAt)
	model.UpdatedAt = types.StringValue(repo.UpdatedAt)
}
