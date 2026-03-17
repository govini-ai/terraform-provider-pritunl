package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &HostDataSource{}

func NewHostDataSource() datasource.DataSource {
	return &HostDataSource{}
}

// HostDataSource defines the data source implementation.
type HostDataSource struct {
	client *client.Client
}

// HostDataSourceModel describes the data source data model.
type HostDataSourceModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Hostname    types.String  `tfsdk:"hostname"`
	Status      types.String  `tfsdk:"status"`
	PublicAddr  types.String  `tfsdk:"public_addr"`
	PublicAddr6 types.String  `tfsdk:"public_addr6"`
	LocalAddr   types.String  `tfsdk:"local_addr"`
	LocalAddr6  types.String  `tfsdk:"local_addr6"`
	CPUUsage    types.Float64 `tfsdk:"cpu_usage"`
	MemUsage    types.Float64 `tfsdk:"mem_usage"`
	Version     types.String  `tfsdk:"version"`
}

func (d *HostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *HostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Pritunl host by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Host ID. Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Host name. Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Host hostname.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Host status.",
				Computed:    true,
			},
			"public_addr": schema.StringAttribute{
				Description: "Public IPv4 address.",
				Computed:    true,
			},
			"public_addr6": schema.StringAttribute{
				Description: "Public IPv6 address.",
				Computed:    true,
			},
			"local_addr": schema.StringAttribute{
				Description: "Local IPv4 address.",
				Computed:    true,
			},
			"local_addr6": schema.StringAttribute{
				Description: "Local IPv6 address.",
				Computed:    true,
			},
			"cpu_usage": schema.Float64Attribute{
				Description: "CPU usage percentage.",
				Computed:    true,
			},
			"mem_usage": schema.Float64Attribute{
				Description: "Memory usage percentage.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Pritunl version.",
				Computed:    true,
			},
		},
	}
}

func (d *HostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var host *client.Host
	var err error

	if !data.ID.IsNull() {
		host, err = d.client.GetHost(data.ID.ValueString())
	} else if !data.Name.IsNull() {
		host, err = d.client.GetHostByName(data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host: %s", err))
		return
	}

	data.ID = types.StringValue(host.ID)
	data.Name = types.StringValue(host.Name)
	data.Hostname = types.StringValue(host.Hostname)
	data.Status = types.StringValue(host.Status)
	data.PublicAddr = types.StringValue(host.PublicAddr)
	data.PublicAddr6 = types.StringValue(host.PublicAddr6)
	data.LocalAddr = types.StringValue(host.LocalAddr)
	data.LocalAddr6 = types.StringValue(host.LocalAddr6)
	data.CPUUsage = types.Float64Value(host.CPUUsage)
	data.MemUsage = types.Float64Value(host.MemUsage)
	data.Version = types.StringValue(host.Version)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
