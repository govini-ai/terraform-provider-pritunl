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
var _ datasource.DataSource = &ServerDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

// ServerDataSource defines the data source implementation.
type ServerDataSource struct {
	client *client.Client
}

// ServerDataSourceModel describes the data source data model.
type ServerDataSourceModel struct {
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

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Pritunl server by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Server ID. Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Server name. Either id or name must be specified.",
				Optional:    true,
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
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var server *client.Server
	var err error

	if !data.ID.IsNull() {
		server, err = d.client.GetServer(data.ID.ValueString())
	} else if !data.Name.IsNull() {
		server, err = d.client.GetServerByName(data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server: %s", err))
		return
	}

	data.ID = types.StringValue(server.ID)
	data.Name = types.StringValue(server.Name)
	data.Network = types.StringValue(server.Network)
	data.Port = types.Int64Value(int64(server.Port))
	data.Protocol = types.StringValue(server.Protocol)
	data.Cipher = types.StringValue(server.Cipher)
	data.Hash = types.StringValue(server.Hash)
	data.InterClient = types.BoolValue(server.InterClient)
	data.PingInterval = types.Int64Value(int64(server.PingInterval))
	data.PingTimeout = types.Int64Value(int64(server.PingTimeout))
	data.Status = types.StringValue(server.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
