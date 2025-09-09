package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-altr/internal/client"
	"terraform-provider-altr/internal/service/policy"
	"terraform-provider-altr/internal/service/repo"
	"terraform-provider-altr/internal/service/sidecar"
)

var _ provider.Provider = &SidecarProvider{}

type SidecarProvider struct {
	version string
}

type SidecarProviderModel struct {
	OrgID   types.String `tfsdk:"org_id"`
	ApiKey  types.String `tfsdk:"api_key"`
	Secret  types.String `tfsdk:"secret"`
	BaseURL types.String `tfsdk:"base_url"`
}

func (p *SidecarProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "altr"
	resp.Version = p.version
}

func (p *SidecarProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "ALTR Provider for managing ALTR SaaS resources.",
		Attributes: map[string]schema.Attribute{
			"org_id": schema.StringAttribute{
				Description: "ALTR Organization ID",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API Key for authentication",
				Optional:    true,
				Sensitive:   true,
			},
			"secret": schema.StringAttribute{
				Description: "API Secret for authentication",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "ALTR base URL",
				Optional:    true,
			},
		},
	}
}

func (p *SidecarProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SidecarProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults from environment variables
	orgID := os.Getenv("ALTR_ORG_ID")
	if !config.OrgID.IsNull() {
		orgID = config.OrgID.ValueString()
	}

	apiKey := os.Getenv("ALTR_API_KEY")
	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	secret := os.Getenv("ALTR_SECRET")
	if !config.Secret.IsNull() {
		secret = config.Secret.ValueString()
	}

	baseURL := os.Getenv("ALTR_BASE_URL")
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if orgID == "" {
		resp.Diagnostics.AddError(
			"Missing Organization ID",
			"Organization ID must be provided via org_id attribute or ALTR_ORG_ID environment variable",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"API Key must be provided via api_key attribute or ALTR_API_KEY environment variable",
		)
	}

	if secret == "" {
		resp.Diagnostics.AddError(
			"Missing Secret",
			"Secret must be provided via secret attribute or ALTR_SECRET environment variable",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API client
	client, err := client.NewClient(orgID, apiKey, secret, baseURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Sidecar API Client",
			"An unexpected error occurred when creating the Sidecar API client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SidecarProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		sidecar.NewSidecarResource,
		repo.NewRepoResource,
		repo.NewRepoUserResource,
		sidecar.NewSidecarListenerResource,
		repo.NewRepoSidecarBindingResource,
		policy.NewAccessManagementOltpPolicyDataResource,
		policy.NewAccessManagementSnowflakePolicyDataResource,
		policy.NewImpersonationPolicyResource,
	}
}

func (p *SidecarProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		sidecar.NewSidecarDataSource,
		sidecar.NewSidecarListenerDataSource,
		repo.NewRepoDataSource,
		repo.NewRepoUserDataSource,
		repo.NewRepoSidecarBindingDataSource,
		policy.NewAccessManagementOltpPolicyDataSource,
		policy.NewAccessManagementSnowflakePolicyDataSource,
		policy.NewImpersonationPolicyDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SidecarProvider{
			version: version,
		}
	}
}
