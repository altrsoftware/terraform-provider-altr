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

var _ datasource.DataSource = &AgentTaskDataSource{}

func NewAgentTaskDataSource() datasource.DataSource {
	return &AgentTaskDataSource{}
}

type AgentTaskDataSource struct {
	client *client.Client
}

type AgentTaskDataSourceModel struct {
	ID            types.String                           `tfsdk:"id"`
	AgentID       types.String                           `tfsdk:"agent_id"`
	Name          types.String                           `tfsdk:"name"`
	Description   types.String                           `tfsdk:"description"`
	RepoName      types.String                           `tfsdk:"repo_name"`
	ServiceUser   types.String                           `tfsdk:"service_user"`
	Configuration *AgentTaskConfigurationDataSourceModel `tfsdk:"configuration"`
	Schedule      *AgentTaskScheduleDataSourceModel      `tfsdk:"schedule"`
	CreatedAt     types.String                           `tfsdk:"created_at"`
	UpdatedAt     types.String                           `tfsdk:"updated_at"`
}

type AgentTaskConfigurationDataSourceModel struct {
	ClassificationType types.Int64               `tfsdk:"classification_type"`
	SampleStrategy     types.String              `tfsdk:"sample_strategy"`
	CollectionName     types.String              `tfsdk:"collection_name"`
	SslConfig          *SslConfigDataSourceModel `tfsdk:"ssl_config"`
}

type SslConfigDataSourceModel struct {
	Enabled                types.Bool   `tfsdk:"enabled"`
	HostnameInCertificate  types.String `tfsdk:"hostname_in_certificate"`
	TrustServerCertificate types.Bool   `tfsdk:"trust_server_certificate"`
	TrustStorePasswordARN  types.String `tfsdk:"trust_store_password_arn"`
	TrustStorePath         types.String `tfsdk:"trust_store_path"`
}

type AgentTaskScheduleDataSourceModel struct {
	Type        types.String `tfsdk:"type"`
	Value       types.String `tfsdk:"value"`
	MaxDuration types.String `tfsdk:"max_duration"`
	Timezone    types.String `tfsdk:"timezone"`
}

func (d *AgentTaskDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_task"
}

func (d *AgentTaskDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for retrieving information about a task assigned to an ALTR agent.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Task UUID.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			"agent_id": schema.StringAttribute{
				Description: "UUID of the agent this task belongs to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(service.UUIDv4Regex),
						"must be a valid UUIDv4",
					),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the task.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the task.",
				Computed:    true,
			},
			"repo_name": schema.StringAttribute{
				Description: "Name of the target repository this task runs against.",
				Computed:    true,
			},
			"service_user": schema.StringAttribute{
				Description: "Username of the service user the agent authenticates as when connecting to the repository.",
				Computed:    true,
			},
			"configuration": schema.SingleNestedAttribute{
				Description: "CLASSIFIER task configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"classification_type": schema.Int64Attribute{
						Description: "Classification engine identifier.",
						Computed:    true,
					},
					"sample_strategy": schema.StringAttribute{
						Description: "Sampling strategy used when collecting data for classification.",
						Computed:    true,
					},
					"collection_name": schema.StringAttribute{
						Description: "Name of the classifier collection to use.",
						Computed:    true,
					},
					"ssl_config": schema.SingleNestedAttribute{
						Description: "SSL/TLS configuration used when connecting to the repository.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Whether SSL/TLS is enabled for the connection.",
								Computed:    true,
							},
							"hostname_in_certificate": schema.StringAttribute{
								Description: "Expected hostname in the server certificate.",
								Computed:    true,
							},
							"trust_server_certificate": schema.BoolAttribute{
								Description: "Whether to trust the server certificate without validation.",
								Computed:    true,
							},
							"trust_store_password_arn": schema.StringAttribute{
								Description: "ARN of the secret holding the trust store password.",
								Computed:    true,
							},
							"trust_store_path": schema.StringAttribute{
								Description: "Path to the trust store used to validate the server certificate.",
								Computed:    true,
							},
						},
					},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				Description: "Schedule controlling when the task runs.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Schedule type (e.g. CRON).",
						Computed:    true,
					},
					"value": schema.StringAttribute{
						Description: "Cron expression (5 fields: minute hour dom month dow).",
						Computed:    true,
					},
					"max_duration": schema.StringAttribute{
						Description: "ISO 8601 duration capping how long a single run may take.",
						Computed:    true,
					},
					"timezone": schema.StringAttribute{
						Description: "IANA timezone name the cron expression is evaluated in.",
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

func (d *AgentTaskDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AgentTaskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AgentTaskDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	task, err := d.client.GetAgentTask(config.AgentID.ValueString(), config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading agent task",
			"Could not read agent task "+config.ID.ValueString()+": "+err.Error(),
		)

		return
	}

	if task == nil {
		resp.Diagnostics.AddError(
			"Agent task not found",
			"Agent task '"+config.ID.ValueString()+"' for agent '"+config.AgentID.ValueString()+"' does not exist.",
		)

		return
	}

	d.mapTaskToModel(task, &config)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func (d *AgentTaskDataSource) mapTaskToModel(task *client.AgentTask, model *AgentTaskDataSourceModel) {
	model.ID = types.StringValue(task.ID)
	model.AgentID = types.StringValue(task.AgentID)
	model.Name = types.StringValue(task.Name)
	model.Description = types.StringValue(task.Description)
	model.RepoName = types.StringValue(task.RepoName)
	model.ServiceUser = types.StringValue(task.ServiceUser)
	model.CreatedAt = types.StringValue(task.CreatedAt)
	model.UpdatedAt = types.StringValue(task.UpdatedAt)

	config := &AgentTaskConfigurationDataSourceModel{
		SampleStrategy: types.StringValue(task.Configuration.SampleStrategy),
		CollectionName: types.StringValue(task.Configuration.CollectionName),
	}

	if task.Configuration.ClassificationType != nil {
		config.ClassificationType = types.Int64Value(int64(*task.Configuration.ClassificationType))
	} else {
		config.ClassificationType = types.Int64Null()
	}

	if task.Configuration.SslConfig != nil {
		config.SslConfig = &SslConfigDataSourceModel{
			Enabled:                types.BoolValue(task.Configuration.SslConfig.Enabled),
			HostnameInCertificate:  types.StringValue(task.Configuration.SslConfig.HostnameInCertificate),
			TrustServerCertificate: types.BoolValue(task.Configuration.SslConfig.TrustServerCertificate),
			TrustStorePasswordARN:  types.StringValue(task.Configuration.SslConfig.TrustStorePasswordARN),
			TrustStorePath:         types.StringValue(task.Configuration.SslConfig.TrustStorePath),
		}
	}

	model.Configuration = config

	model.Schedule = &AgentTaskScheduleDataSourceModel{
		Type:        types.StringValue(task.Schedule.Type),
		Value:       types.StringValue(task.Schedule.Value),
		MaxDuration: types.StringValue(task.Schedule.MaxDuration),
		Timezone:    types.StringValue(task.Schedule.Timezone),
	}
}
