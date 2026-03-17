package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

// ServerResource defines the resource implementation.
type ServerResource struct {
	client *client.Client
}

// ServerResourceModel describes the resource data model.
type ServerResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Network                 types.String `tfsdk:"network"`
	Port                    types.Int64  `tfsdk:"port"`
	Protocol                types.String `tfsdk:"protocol"`
	Cipher                  types.String `tfsdk:"cipher"`
	Hash                    types.String `tfsdk:"hash"`
	InterClient             types.Bool   `tfsdk:"inter_client"`
	PingInterval            types.Int64  `tfsdk:"ping_interval"`
	PingTimeout             types.Int64  `tfsdk:"ping_timeout"`
	Status                  types.String `tfsdk:"status"`
	AttachedOrganizationIDs types.List   `tfsdk:"attached_organization_ids"`
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pritunl VPN server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Server ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Server name.",
				Required:    true,
			},
			"network": schema.StringAttribute{
				Description: "VPN network CIDR.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Server port.",
				Required:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "Protocol (udp or tcp).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("udp"),
			},
			"cipher": schema.StringAttribute{
				Description: "Encryption cipher.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("aes256"),
			},
			"hash": schema.StringAttribute{
				Description: "Hash algorithm.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("sha256"),
			},
			"inter_client": schema.BoolAttribute{
				Description: "Allow inter-client routing.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ping_interval": schema.Int64Attribute{
				Description: "Ping interval in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
			},
			"ping_timeout": schema.Int64Attribute{
				Description: "Ping timeout in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
			},
			"status": schema.StringAttribute{
				Description: "Server status (online/offline).",
				Computed:    true,
			},
			"attached_organization_ids": schema.ListAttribute{
				Description: "List of organization IDs attached to this server.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := &client.Server{
		Name:         data.Name.ValueString(),
		Network:      data.Network.ValueString(),
		Port:         int(data.Port.ValueInt64()),
		Protocol:     data.Protocol.ValueString(),
		Cipher:       data.Cipher.ValueString(),
		Hash:         data.Hash.ValueString(),
		InterClient:  data.InterClient.ValueBool(),
		PingInterval: int(data.PingInterval.ValueInt64()),
		PingTimeout:  int(data.PingTimeout.ValueInt64()),
	}

	tflog.Debug(ctx, "Creating server", map[string]interface{}{"name": server.Name})

	created, err := r.client.CreateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create server: %s", err))
		return
	}

	data.ID = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Network = types.StringValue(created.Network)
	data.Port = types.Int64Value(int64(created.Port))
	data.Protocol = types.StringValue(created.Protocol)
	data.Cipher = types.StringValue(created.Cipher)
	data.Hash = types.StringValue(created.Hash)
	data.InterClient = types.BoolValue(created.InterClient)
	data.PingInterval = types.Int64Value(int64(created.PingInterval))
	data.PingTimeout = types.Int64Value(int64(created.PingTimeout))
	data.Status = types.StringValue(created.Status)

	// Attach organizations if specified
	if !data.AttachedOrganizationIDs.IsNull() && !data.AttachedOrganizationIDs.IsUnknown() {
		var orgIDs []string
		diags := data.AttachedOrganizationIDs.ElementsAs(ctx, &orgIDs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, orgID := range orgIDs {
			tflog.Debug(ctx, "Attaching organization to server", map[string]interface{}{
				"server_id": created.ID,
				"org_id":    orgID,
			})
			if err := r.client.AttachOrganization(created.ID, orgID); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to attach organization %s to server: %s", orgID, err))
				return
			}
		}
	}

	tflog.Trace(ctx, "Created server", map[string]interface{}{"id": created.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(data.ID.ValueString())
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

	// Get attached organizations
	orgs, err := r.client.GetServerOrganizations(server.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server organizations: %s", err))
		return
	}

	if len(orgs) > 0 {
		orgIDs := make([]string, len(orgs))
		for i, org := range orgs {
			orgIDs[i] = org.ID
		}
		orgIDsList, diags := types.ListValueFrom(ctx, types.StringType, orgIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.AttachedOrganizationIDs = orgIDsList
	} else {
		data.AttachedOrganizationIDs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServerResourceModel
	var state ServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := &client.Server{
		Name:         data.Name.ValueString(),
		Network:      data.Network.ValueString(),
		Port:         int(data.Port.ValueInt64()),
		Protocol:     data.Protocol.ValueString(),
		Cipher:       data.Cipher.ValueString(),
		Hash:         data.Hash.ValueString(),
		InterClient:  data.InterClient.ValueBool(),
		PingInterval: int(data.PingInterval.ValueInt64()),
		PingTimeout:  int(data.PingTimeout.ValueInt64()),
	}

	tflog.Debug(ctx, "Updating server", map[string]interface{}{"id": data.ID.ValueString(), "name": server.Name})

	updated, err := r.client.UpdateServer(data.ID.ValueString(), server)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server: %s", err))
		return
	}

	// Handle organization attachments
	var plannedOrgIDs []string
	var currentOrgIDs []string

	if !data.AttachedOrganizationIDs.IsNull() && !data.AttachedOrganizationIDs.IsUnknown() {
		diags := data.AttachedOrganizationIDs.ElementsAs(ctx, &plannedOrgIDs, false)
		resp.Diagnostics.Append(diags...)
	}

	if !state.AttachedOrganizationIDs.IsNull() && !state.AttachedOrganizationIDs.IsUnknown() {
		diags := state.AttachedOrganizationIDs.ElementsAs(ctx, &currentOrgIDs, false)
		resp.Diagnostics.Append(diags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Detach orgs no longer in plan
	for _, currentID := range currentOrgIDs {
		found := false
		for _, plannedID := range plannedOrgIDs {
			if currentID == plannedID {
				found = true
				break
			}
		}
		if !found {
			tflog.Debug(ctx, "Detaching organization from server", map[string]interface{}{
				"server_id": updated.ID,
				"org_id":    currentID,
			})
			if err := r.client.DetachOrganization(updated.ID, currentID); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to detach organization %s from server: %s", currentID, err))
				return
			}
		}
	}

	// Attach new orgs
	for _, plannedID := range plannedOrgIDs {
		found := false
		for _, currentID := range currentOrgIDs {
			if plannedID == currentID {
				found = true
				break
			}
		}
		if !found {
			tflog.Debug(ctx, "Attaching organization to server", map[string]interface{}{
				"server_id": updated.ID,
				"org_id":    plannedID,
			})
			if err := r.client.AttachOrganization(updated.ID, plannedID); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to attach organization %s to server: %s", plannedID, err))
				return
			}
		}
	}

	data.Name = types.StringValue(updated.Name)
	data.Network = types.StringValue(updated.Network)
	data.Port = types.Int64Value(int64(updated.Port))
	data.Protocol = types.StringValue(updated.Protocol)
	data.Cipher = types.StringValue(updated.Cipher)
	data.Hash = types.StringValue(updated.Hash)
	data.InterClient = types.BoolValue(updated.InterClient)
	data.PingInterval = types.Int64Value(int64(updated.PingInterval))
	data.PingTimeout = types.Int64Value(int64(updated.PingTimeout))
	data.Status = types.StringValue(updated.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting server", map[string]interface{}{"id": data.ID.ValueString()})

	// Stop server before deleting
	_ = r.client.StopServer(data.ID.ValueString())

	err := r.client.DeleteServer(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete server: %s", err))
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	server, err := r.client.GetServer(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import server: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), server.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), server.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network"), server.Network)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), server.Port)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("protocol"), server.Protocol)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cipher"), server.Cipher)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hash"), server.Hash)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inter_client"), server.InterClient)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ping_interval"), server.PingInterval)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ping_timeout"), server.PingTimeout)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("status"), server.Status)...)
}
