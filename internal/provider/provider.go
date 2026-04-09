package provider

import (
	"context"
	"os"

	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
	"github.com/govini-ai/terraform-provider-pritunl/internal/datasources"
	"github.com/govini-ai/terraform-provider-pritunl/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure PritunlProvider satisfies various provider interfaces.
var _ provider.Provider = &PritunlProvider{}

// PritunlProvider defines the provider implementation.
type PritunlProvider struct {
	version string
}

// PritunlProviderModel describes the provider data model.
type PritunlProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Token    types.String `tfsdk:"token"`
	Secret   types.String `tfsdk:"secret"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PritunlProvider{
			version: version,
		}
	}
}

func (p *PritunlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pritunl"
	resp.Version = p.version
}

func (p *PritunlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Pritunl VPN server.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "URL of the Pritunl server. Can also be set via PRITUNL_URL environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API token for Pritunl authentication. Can also be set via PRITUNL_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"secret": schema.StringAttribute{
				Description: "API secret for Pritunl authentication. Can also be set via PRITUNL_SECRET environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Defaults to false. Can also be set via PRITUNL_INSECURE environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *PritunlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Pritunl client")

	var config PritunlProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables
	url := os.Getenv("PRITUNL_URL")
	token := os.Getenv("PRITUNL_TOKEN")
	secret := os.Getenv("PRITUNL_SECRET")
	insecure := os.Getenv("PRITUNL_INSECURE") == "true"

	// Override with config values if provided
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	if !config.Secret.IsNull() {
		secret = config.Secret.ValueString()
	}
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// Validate required configuration
	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing Pritunl URL",
			"The provider cannot create the Pritunl API client as there is a missing or empty value for the Pritunl URL. "+
				"Set the url value in the configuration or use the PRITUNL_URL environment variable.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Pritunl API Token",
			"The provider cannot create the Pritunl API client as there is a missing or empty value for the Pritunl API token. "+
				"Set the token value in the configuration or use the PRITUNL_TOKEN environment variable.",
		)
	}

	if secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("secret"),
			"Missing Pritunl API Secret",
			"The provider cannot create the Pritunl API client as there is a missing or empty value for the Pritunl API secret. "+
				"Set the secret value in the configuration or use the PRITUNL_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Pritunl client", map[string]interface{}{
		"url":      url,
		"insecure": insecure,
	})

	// Create client
	c := client.NewClient(url, token, secret, insecure)

	// Verify connection
	if err := c.Status(); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to Pritunl API",
			"An error occurred when connecting to the Pritunl API. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Pritunl Client Error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Configured Pritunl client", map[string]interface{}{"success": true})

	// Make the client available during DataSource and Resource type Configure methods.
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *PritunlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewOrganizationResource,
		resources.NewServerResource,
		resources.NewRouteResource,
		resources.NewUserResource,
	}
}

func (p *PritunlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewOrganizationDataSource,
		datasources.NewOrganizationsDataSource,
		datasources.NewServerDataSource,
		datasources.NewServersDataSource,
		datasources.NewHostDataSource,
		datasources.NewHostsDataSource,
		datasources.NewUserDataSource,
		datasources.NewUsersDataSource,
	}
}
