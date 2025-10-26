package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

// Ensure TruenasProvider satisfies various provider interfaces.
var _ provider.Provider = &TruenasProvider{}

// TruenasProvider defines the provider implementation.
type TruenasProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TruenasProviderModel describes the provider data model.
type TruenasProviderModel struct {
	BaseURL types.String `tfsdk:"base_url"`
	APIKey  types.String `tfsdk:"api_key"`
}

func (p *TruenasProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "truenas"
	resp.Version = p.version
}

func (p *TruenasProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for TrueNAS Scale 24.04 REST API",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "The base URL of the TrueNAS server (e.g., http://10.0.0.213:81 or https://truenas.local). Can also be set via TRUENAS_BASE_URL environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for authenticating with TrueNAS. Can also be set via TRUENAS_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *TruenasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TruenasProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// If practitioners don't configure a value, check environment variables.
	baseURL := os.Getenv("TRUENAS_BASE_URL")
	apiKey := os.Getenv("TRUENAS_API_KEY")

	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing TrueNAS Base URL",
			"The provider cannot create the TrueNAS API client as there is a missing or empty value for the TrueNAS base URL. "+
				"Set the base_url value in the configuration or use the TRUENAS_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing TrueNAS API Key",
			"The provider cannot create the TrueNAS API client as there is a missing or empty value for the TrueNAS API key. "+
				"Set the api_key value in the configuration or use the TRUENAS_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new TrueNAS client using the configuration values
	client, err := truenas.NewClient(baseURL, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create TrueNAS API Client",
			"An unexpected error occurred when creating the TrueNAS API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"TrueNAS Client Error: "+err.Error(),
		)
		return
	}

	// Make the TrueNAS client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TruenasProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatasetResource,
		NewNFSShareResource,
		NewSMBShareResource,
		NewUserResource,
		NewGroupResource,
	}
}

func (p *TruenasProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDatasetDataSource,
		NewPoolDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TruenasProvider{
			version: version,
		}
	}
}

