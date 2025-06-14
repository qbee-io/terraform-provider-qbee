package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.qbee.io/client"
	"os"
)

// Ensure QbeeProvider satisfies various provider interfaces.
var _ provider.Provider = &QbeeProvider{}

// QbeeProvider defines the provider implementation.
type QbeeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// qbeeProviderModel describes the provider data model.
type qbeeProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	BaseURL  types.String `tfsdk:"base_url"`
}

func (p *QbeeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "qbee"
	resp.Version = p.version
}

func (p *QbeeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Qbee username. Can also be set using the QBEE_USERNAME environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Qbee password. Can also be set using the QBEE_PASSWORD environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Qbee base URL. Defaults to `https://www.app.qbee.io`. Can also be set using the QBEE_BASE_URL environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *QbeeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config qbeeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Qbee API Username",
			"The provider cannot create the Qbee API client as there is an unknown configuration value for the Qbee API username. "+
				"Either target apply the source of the value first, set the QBEE_USERNAME environment variable or set the value statically in the configuration.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Qbee API Password",
			"The provider cannot create the Qbee API client as there is an unknown configuration value for the Qbee API password. "+
				"Either target apply the source of the value first, set the QBEE_PASSWORD environment variable or set the value statically in the configuration",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	username := os.Getenv("QBEE_USERNAME")
	password := os.Getenv("QBEE_PASSWORD")
	baseUrl := os.Getenv("QBEE_BASE_URL")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.BaseURL.IsNull() {
		baseUrl = config.BaseURL.ValueString()
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(path.Root("username"), "Missing Qbee API username",
			"the Qbee API client can not be created because the username is missing."+
				"Set the username in the provider config or use the QBEE_USERNAME environment variable.")
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Missing Qbee API username",
			"the Qbee API client can not be created because the username is missing."+
				"Set the password in the provider config or use the QBEE_PASSWORD environment variable.")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	qbeeClient := client.New()

	if baseUrl != "" {
		qbeeClient = qbeeClient.WithBaseURL(baseUrl)
	}

	err := qbeeClient.Authenticate(ctx, username, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Qbee API Client",
			"An unexpected error occurred when creating the Qbee API client: "+err.Error())
		return
	}

	resp.DataSourceData = qbeeClient
	resp.ResourceData = qbeeClient
}

func (p *QbeeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewConnectivityWatchdogResource,
		NewDockerContainersResource,
		NewFiledistributionResource,
		NewFilemanagerDirectoryResource,
		NewFilemanagerFileResource,
		NewFirewallResource,
		NewGrouptreeGroupResource,
		NewMetricsMonitorResource,
		NewPackageManagementResource,
		NewParametersResource,
		NewPasswordResource,
		NewPodmanContainersResource,
		NewProcessWatchResource,
		NewRaucResource,
		NewSSHKeysResource,
		NewSettingsResource,
		NewSoftwareManagementResource,
		NewUsersResource,
		NewBootstrapKeyResource,
		NewRoleResource,
	}
}

func (p *QbeeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &QbeeProvider{
			version: version,
		}
	}
}
