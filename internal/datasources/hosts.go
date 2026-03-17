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
var _ datasource.DataSource = &HostsDataSource{}

func NewHostsDataSource() datasource.DataSource {
	return &HostsDataSource{}
}

// HostsDataSource defines the data source implementation.
type HostsDataSource struct {
	client *client.Client
}

// HostsDataSourceModel describes the data source data model.
type HostsDataSourceModel struct {
	Hosts []HostModel `tfsdk:"hosts"`
}

// HostModel describes a single host.
type HostModel struct {
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

func (d *HostsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts"
}

func (d *HostsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Pritunl hosts.",
		Attributes: map[string]schema.Attribute{
			"hosts": schema.ListNestedAttribute{
				Description: "List of hosts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Host ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Host name.",
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
				},
			},
		},
	}
}

func (d *HostsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HostsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hosts, err := d.client.ListHosts()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read hosts: %s", err))
		return
	}

	data.Hosts = make([]HostModel, len(hosts))
	for i, host := range hosts {
		data.Hosts[i] = HostModel{
			ID:          types.StringValue(host.ID),
			Name:        types.StringValue(host.Name),
			Hostname:    types.StringValue(host.Hostname),
			Status:      types.StringValue(host.Status),
			PublicAddr:  types.StringValue(host.PublicAddr),
			PublicAddr6: types.StringValue(host.PublicAddr6),
			LocalAddr:   types.StringValue(host.LocalAddr),
			LocalAddr6:  types.StringValue(host.LocalAddr6),
			CPUUsage:    types.Float64Value(host.CPUUsage),
			MemUsage:    types.Float64Value(host.MemUsage),
			Version:     types.StringValue(host.Version),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
