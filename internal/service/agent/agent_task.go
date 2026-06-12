// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &AgentTaskResource{}
	_ resource.ResourceWithImportState = &AgentTaskResource{}
)

func NewAgentTaskResource() resource.Resource {
	return &AgentTaskResource{}
}

type AgentTaskResource struct {
	client *client.Client
}

type AgentTaskResourceModel struct {
	ID            types.String          `tfsdk:"id"`
	AgentID       types.String          `tfsdk:"agent_id"`
	Name          types.String          `tfsdk:"name"`
	Description   types.String          `tfsdk:"description"`
	RepoName      types.String          `tfsdk:"repo_name"`
	ServiceUser   types.String          `tfsdk:"service_user"`
	Configuration basetypes.ObjectValue `tfsdk:"configuration"`
	Schedule      basetypes.ObjectValue `tfsdk:"schedule"`
	CreatedAt     types.String          `tfsdk:"created_at"`
	UpdatedAt     types.String          `tfsdk:"updated_at"`
}

var sslConfigAttrTypes = map[string]attr.Type{
	"enabled":                  types.BoolType,
	"hostname_in_certificate":  types.StringType,
	"trust_server_certificate": types.BoolType,
	"trust_store_password_arn": types.StringType,
	"trust_store_path":         types.StringType,
}

var configAttrTypes = map[string]attr.Type{
	"collection_name":     types.StringType,
	"classification_type": types.Int64Type,
	"sample_strategy":     types.StringType,
	"ssl_config":          types.ObjectType{AttrTypes: sslConfigAttrTypes},
}

var scheduleAttrTypes = map[string]attr.Type{
	"type":         types.StringType,
	"value":        types.StringType,
	"max_duration": types.StringType,
	"timezone":     types.StringType,
}

func (r *AgentTaskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_task"
}

func (r *AgentTaskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a task assigned to an ALTR CLASSIFIER agent. Tasks run against a repository on a schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Task UUID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent_id": schema.StringAttribute{
				Description: "UUID of the agent this task belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the task.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the task.",
				Optional:    true,
				Computed:    true,
			},
			"repo_name": schema.StringAttribute{
				Description: "Name of the target repository this task runs against.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_user": schema.StringAttribute{
				Description: "Username of the service user the agent authenticates as when connecting to the repository.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configuration": schema.SingleNestedAttribute{
				Description: "CLASSIFIER task configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"classification_type": schema.Int64Attribute{
						Description: "Classification engine to use: 1 (GOOGLE_DLP), 2 (SNOWFLAKE_NATIVE), 3 (SNOWFLAKE_OBJECT_TAG_IMPORT), 4 (SNOWFLAKE_NATIVE_AND_TAG_IMPORT), or 5 (ALTR_NATIVE).",
						Required:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(1, 2, 3, 4, 5),
						},
					},
					"sample_strategy": schema.StringAttribute{
						Description: "Sampling strategy: ROWS (row data only), METADATA (column metadata only), or COMBINED (both).",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ROWS", "METADATA", "COMBINED"),
						},
					},
					"collection_name": schema.StringAttribute{
						Description: "Name of the classifier collection to use. May only be set when classification_type is 5 (ALTR_NATIVE).",
						Optional:    true,
						Computed:    true,
					},
					"ssl_config": schema.SingleNestedAttribute{
						Description: "SSL/TLS configuration used when connecting to the repository.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Whether SSL/TLS is enabled for the connection.",
								Optional:    true,
								Computed:    true,
							},
							"hostname_in_certificate": schema.StringAttribute{
								Description: "Expected hostname in the server certificate.",
								Optional:    true,
								Computed:    true,
							},
							"trust_server_certificate": schema.BoolAttribute{
								Description: "Whether to trust the server certificate without validation.",
								Optional:    true,
								Computed:    true,
							},
							"trust_store_password_arn": schema.StringAttribute{
								Description: "ARN of the secret holding the trust store password.",
								Optional:    true,
								Computed:    true,
							},
							"trust_store_path": schema.StringAttribute{
								Description: "Path to the trust store used to validate the server certificate.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				Description: "Schedule controlling when the task runs.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Schedule type. Must be 'CRON'.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("CRON"),
						},
					},
					"value": schema.StringAttribute{
						Description: "Cron expression (5 fields: minute hour dom month dow).",
						Required:    true,
					},
					"max_duration": schema.StringAttribute{
						Description: "ISO 8601 duration capping how long a single run may take (e.g. PT30M).",
						Optional:    true,
						Computed:    true,
					},
					"timezone": schema.StringAttribute{
						Description: "IANA timezone name the cron expression is evaluated in (e.g. America/New_York). Defaults to UTC.",
						Optional:    true,
						Computed:    true,
					},
				},
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

func (r *AgentTaskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = c
}

func (r *AgentTaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AgentTaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.validateConfiguration(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.CreateAgentTaskInput{
		Name:          plan.Name.ValueString(),
		Description:   plan.Description.ValueString(),
		RepoName:      plan.RepoName.ValueString(),
		ServiceUser:   plan.ServiceUser.ValueString(),
		Configuration: r.configFromModel(plan.Configuration),
		Schedule:      r.scheduleFromModel(plan.Schedule),
	}

	task, err := r.client.CreateAgentTask(plan.AgentID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating agent task",
			"Could not create agent task, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapTaskToModel(task, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AgentTaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AgentTaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	task, err := r.client.GetAgentTask(state.AgentID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading agent task",
			"Could not read agent task "+state.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	if task == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	r.mapTaskToModel(task, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *AgentTaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  AgentTaskResourceModel
		state AgentTaskResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.validateConfiguration(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())

		return
	}

	input := client.UpdateAgentTaskInput{}

	if !plan.Name.Equal(state.Name) {
		input.Name = plan.Name.ValueStringPointer()
	}

	if !plan.Description.Equal(state.Description) {
		input.Description = plan.Description.ValueStringPointer()
	}

	if !plan.Configuration.Equal(state.Configuration) {
		cfg := r.configFromModel(plan.Configuration)
		input.Configuration = &cfg
	}

	if !plan.Schedule.Equal(state.Schedule) {
		sched := r.scheduleFromModel(plan.Schedule)
		input.Schedule = &sched
	}

	task, err := r.client.UpdateAgentTask(state.AgentID.ValueString(), state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating agent task",
			"Could not update agent task, unexpected error: "+err.Error(),
		)

		return
	}

	r.mapTaskToModel(task, &plan)

	// agent_id has RequiresReplace so it never changes, but preserve the known
	// state value in case the API response omits the field.
	if plan.AgentID.ValueString() == "" {
		plan.AgentID = state.AgentID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AgentTaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AgentTaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAgentTask(state.AgentID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting agent task",
			"Could not delete agent task, unexpected error: "+err.Error(),
		)

		return
	}
}

func (r *AgentTaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: agent_id:task_id
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID in format: agent_id:task_id",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// collectionNameClassificationType is the only classification_type
// (ALTR_NATIVE) for which the API accepts a collection_name.
const collectionNameClassificationType = 5

func (r *AgentTaskResource) validateConfiguration(model *AgentTaskResourceModel) error {
	attrs := model.Configuration.Attributes()

	collectionName := attrs["collection_name"].(types.String)
	hasCollection := !collectionName.IsNull() && !collectionName.IsUnknown() && collectionName.ValueString() != ""

	classificationType := attrs["classification_type"].(types.Int64)

	if hasCollection && (classificationType.IsNull() || classificationType.ValueInt64() != collectionNameClassificationType) {
		return fmt.Errorf("'collection_name' may only be set when 'classification_type' is %d (ALTR_NATIVE)", collectionNameClassificationType)
	}

	return nil
}

func (r *AgentTaskResource) configFromModel(obj basetypes.ObjectValue) client.AgentTaskConfiguration {
	attrs := obj.Attributes()

	cfg := client.AgentTaskConfiguration{
		CollectionName: attrs["collection_name"].(types.String).ValueString(),
		SampleStrategy: attrs["sample_strategy"].(types.String).ValueString(),
	}

	if ct := attrs["classification_type"].(types.Int64); !ct.IsNull() && !ct.IsUnknown() {
		v := int(ct.ValueInt64())
		cfg.ClassificationType = &v
	}

	if ssl, ok := attrs["ssl_config"].(basetypes.ObjectValue); ok && !ssl.IsNull() && !ssl.IsUnknown() {
		sslAttrs := ssl.Attributes()
		cfg.SslConfig = &client.SslConfig{
			Enabled:                sslAttrs["enabled"].(types.Bool).ValueBool(),
			HostnameInCertificate:  sslAttrs["hostname_in_certificate"].(types.String).ValueString(),
			TrustServerCertificate: sslAttrs["trust_server_certificate"].(types.Bool).ValueBool(),
			TrustStorePasswordARN:  sslAttrs["trust_store_password_arn"].(types.String).ValueString(),
			TrustStorePath:         sslAttrs["trust_store_path"].(types.String).ValueString(),
		}
	}

	return cfg
}

func (r *AgentTaskResource) scheduleFromModel(obj basetypes.ObjectValue) client.AgentTaskSchedule {
	attrs := obj.Attributes()

	return client.AgentTaskSchedule{
		Type:        attrs["type"].(types.String).ValueString(),
		Value:       attrs["value"].(types.String).ValueString(),
		MaxDuration: attrs["max_duration"].(types.String).ValueString(),
		Timezone:    attrs["timezone"].(types.String).ValueString(),
	}
}

func (r *AgentTaskResource) mapTaskToModel(task *client.AgentTask, model *AgentTaskResourceModel) {
	model.ID = types.StringValue(task.ID)
	model.AgentID = types.StringValue(task.AgentID)
	model.Name = types.StringValue(task.Name)
	model.Description = types.StringValue(task.Description)
	model.RepoName = types.StringValue(task.RepoName)
	model.ServiceUser = types.StringValue(task.ServiceUser)
	model.CreatedAt = types.StringValue(task.CreatedAt)
	model.UpdatedAt = types.StringValue(task.UpdatedAt)

	classificationType := types.Int64Null()
	if task.Configuration.ClassificationType != nil {
		classificationType = types.Int64Value(int64(*task.Configuration.ClassificationType))
	}

	sslConfig := basetypes.NewObjectNull(sslConfigAttrTypes)
	if task.Configuration.SslConfig != nil {
		sslConfig = basetypes.NewObjectValueMust(sslConfigAttrTypes, map[string]attr.Value{
			"enabled":                  types.BoolValue(task.Configuration.SslConfig.Enabled),
			"hostname_in_certificate":  types.StringValue(task.Configuration.SslConfig.HostnameInCertificate),
			"trust_server_certificate": types.BoolValue(task.Configuration.SslConfig.TrustServerCertificate),
			"trust_store_password_arn": types.StringValue(task.Configuration.SslConfig.TrustStorePasswordARN),
			"trust_store_path":         types.StringValue(task.Configuration.SslConfig.TrustStorePath),
		})
	}

	model.Configuration = basetypes.NewObjectValueMust(configAttrTypes, map[string]attr.Value{
		"collection_name":     types.StringValue(task.Configuration.CollectionName),
		"classification_type": classificationType,
		"sample_strategy":     types.StringValue(task.Configuration.SampleStrategy),
		"ssl_config":          sslConfig,
	})

	model.Schedule = basetypes.NewObjectValueMust(scheduleAttrTypes, map[string]attr.Value{
		"type":         types.StringValue(task.Schedule.Type),
		"value":        types.StringValue(task.Schedule.Value),
		"max_duration": types.StringValue(task.Schedule.MaxDuration),
		"timezone":     types.StringValue(task.Schedule.Timezone),
	})
}
