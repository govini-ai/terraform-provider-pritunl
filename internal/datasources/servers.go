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
var _ datasource.DataSource = &ServersDataSource{}

func NewServersDataSource() datasource.DataSource {
	return &ServersDataSource{}
}

// ServersDataSource defines the data source implementation.
type ServersDataSource struct {
	client *client.Client
}

// ServersDataSourceModel describes the data source data model.
type ServersDataSourceModel struct {
	Servers []ServerModel `tfsdk:"servers"`
}

// ServerModel describes a single server.
type ServerModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Network      types.String `tfsdk:"network"`
	Port         types.Int64  `tfsdk:"port"`
	Protocol     types.String `tfsdk:"protocol"`
	Cipher       types.String `tfsdk:"cipher"`
	Hash         types.String `tfsdk:"hash"`
	InterClient  types.Bool   `tfsdk:"inter_client"`
	PingInterval types.Int64  `tfsdk:"ping_interval"`
	PingTimeout  types.Int64  `tfsdk:"ping_timeout"`
	Status       types.String `tfsdk:"status"`
}

func (d *ServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Pritunl servers.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Description: "List of servers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Server ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Server name.",
							Computed:    true,
						},
						"network": schema.StringAttribute{
							Description: "VPN network CIDR.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Server port.",
							Computed:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol (udp or tcp).",
							Computed:    true,
						},
						"cipher": schema.StringAttribute{
							Description: "Encryption cipher.",
							Computed:    true,
						},
						"hash": schema.StringAttribute{
							Description: "Hash algorithm.",
							Computed:    true,
						},
						"inter_client": schema.BoolAttribute{
							Description: "Allow inter-client routing.",
							Computed:    true,
						},
						"ping_interval": schema.Int64Attribute{
							Description: "Ping interval in seconds.",
							Computed:    true,
						},
						"ping_timeout": schema.Int64Attribute{
							Description: "Ping timeout in seconds.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Server status (online/offline).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	servers, err := d.client.ListServers()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read servers: %s", err))
		return
	}

	data.Servers = make([]ServerModel, len(servers))
	for i, server := range servers {
		data.Servers[i] = ServerModel{
			ID:           types.StringValue(server.ID),
			Name:         types.StringValue(server.Name),
			Network:      types.StringValue(server.Network),
			Port:         types.Int64Value(int64(server.Port)),
			Protocol:     types.StringValue(server.Protocol),
			Cipher:       types.StringValue(server.Cipher),
			Hash:         types.StringValue(server.Hash),
			InterClient:  types.BoolValue(server.InterClient),
			PingInterval: types.Int64Value(int64(server.PingInterval)),
			PingTimeout:  types.Int64Value(int64(server.PingTimeout)),
			Status:       types.StringValue(server.Status),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
